package main

import (
    "io"
	"fmt"
	"log"
	"net/http"
)

type clientsPool chan *http.Client

var requestQueue chan http.Request
var retriesQueue chan *http.Request

//var clients clientsPool

var backendsAddr []string
var clients map[string]clientsPool

func main() {
	requestQueueSize := 1000
	retriesQueueSize := 1000
	numOfClients := 1000

	backendsAddr = []string{
		"localhost:8081",
	}

	requestQueue := make(chan http.Request, requestQueueSize)
	retryQueue := make(chan http.Request, retriesQueueSize)
	//clients := make(chan http.Request, numOfClients)

	log.Printf("%v", requestQueue)
	log.Printf("%v", retryQueue)
	log.Printf("%v", clients)

	clients = make(map[string]clientsPool, len(backendsAddr))

	for _, backendAddr := range backendsAddr {
        //clients[backendAddr] = make(clientsPool, numOfClients)
        backendClients := make(clientsPool, numOfClients)
        for i := 0; i < numOfClients; i++ {
            backendClients <- &http.Client{}
            //backendClients[i] = &Client{} // fill this with all options relevant
		//log.Println(i)
        }
		// Maybe there is a way to init all the clients inside of the cilent pool
		//clients[backendAddr] = make(clientsPool, numOfClients)
        clients[backendAddr] = backendClients
	}

	http.HandleFunc("/", writtenFlagWrapper(handleRequest))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func writtenFlagWrapper(f func(w UseAwareResponseWriter, r *http.Request)) http.HandlerFunction {
    return func (w http.ResponseWriter, r *http.Request) {
        for _, backendAddr := range backendsAddr {
            // getting lazy here: I'm not using channels here cause the go routinesl
            // are bound to the connection queue size * N (number of servers)
            go func() {
                log.Printf("Sending request to: [%s]", backendAddr)
                newReq := copyRequest(req, backendAddr)
                c := <- clients[backendAddr] 
                resp, err := c.Do(newReq)
                if err != nil {
                    // For simplicty: retrying failed request for all errors (which might not be that good)
                    retryReqCopy := copyRequest(req, backendAddr)
                    retriesQueue <- retryReqCopy
                }
                if !w.ResponseWritten() {
                    w.WriteHeader(http.StatusCreated)
                    w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
                    w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
                    io.Copy(w, resp.Body)
                } 
                resp.Body.Close()
            }()
        }
    }
}

//func handleRequest(w http.ResponseWriter, req *http.Request) {
func handleRequest(w http.ResponseWriter, req *http.Request) {
    for _, backendAddr := range backendsAddr {
        // getting lazy here: I'm not using channels here cause the go routinesl
        // are bound to the connection queue size * N (number of servers)
        go func() {
            log.Printf("Sending request to: [%s]", backendAddr)
            newReq := copyRequest(req, backendAddr)
            c := <- clients[backendAddr] 
            resp, err := c.Do(newReq)
            if err != nil {
                // For simplicty: retrying failed request for all errors (which might not be that good)
                retryReqCopy := copyRequest(req, backendAddr)
                retriesQueue <- retryReqCopy
            }
            if !w.ResponseWritten() {
                w.WriteHeader(http.StatusCreated)
                w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
                w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
                io.Copy(w, resp.Body)
            } 
            resp.Body.Close()
        }()
    }
}

//func dispatchSingleRequestToBackends(backends []string, r *httpRequest) {
//client := &http.Client{
//Timeout: time.Second * 10,
//}
//for backend := range backends {

//}

//}
//func retry_send(backend string, http.Request *r) {
//d := 1; // ms
//for {

//d = d*2
//}
//}

func copyRequest(req *http.Request, backendAddr string) *http.Request {
    // we need to buffer the body if we want to read it here and send it
    // in the request.
    //body, err := ioutil.ReadAll(req.Body)
    //if err != nil {
    //http.Error(w, err.Error(), http.StatusInternalServerError)
    //return
    //}

    // you can reassign the body if you need to parse it as multipart
    //req.Body = ioutil.NopCloser(bytes.NewReader(body))

    // create a new url from the raw RequestURI sent by the client
    url := fmt.Sprintf("http://%s%s", backendAddr, req.RequestURI)
    proxyReq, err := http.NewRequest(req.Method, url, req.Body)
    if err != nil {
        return nil
    }

    // We may want to filter some headers, otherwise we could just use a shallow copy
    proxyReq.Header = req.Header
    //proxyReq.Header = make(http.Header)
    //for h, val := range req.Header {
    //proxyReq.Header[h] = val
    //}

    //resp, err := httpClient.Do(proxyReq)
    return proxyReq
}

//func handler(w http.ResponseWriter, req *http.Request) {
//// we need to buffer the body if we want to read it here and send it
//// in the request.
//body, err := ioutil.ReadAll(req.Body)
//if err != nil {
//http.Error(w, err.Error(), http.StatusInternalServerError)
//return
//}

//// you can reassign the body if you need to parse it as multipart
//req.Body = ioutil.NopCloser(bytes.NewReader(body))

//// create a new url from the raw RequestURI sent by the client
//url := fmt.Sprintf("%s://%s%s", proxyScheme, proxyHost, req.RequestURI)

//proxyReq, err := http.NewRequest(req.Method, url, bytes.NewReader(body))

//// We may want to filter some headers, otherwise we could just use a shallow copy
//// proxyReq.Header = req.Header
//proxyReq.Header = make(http.Header)
//for h, val := range req.Header {
//proxyReq.Header[h] = val
//}

//resp, err := httpClient.Do(proxyReq)
//if err != nil {
//http.Error(w, err.Error(), http.StatusBadGateway)
//return
//}
//defer resp.Body.Close()

type UseAwareResponseWriter struct {
    //status int 
    responseWritten bool
    http.ResponseWriter
}

func NewUseAwareResponseWriter(res http.ResponseWriter) *UseAwareResponseWriter {
    // Default the status code to 200
    return &UseAwareResponseWriter{responseWritten: false, ResponseWriter: res}
}

// Give a way to get the status
func (w UseAwareResponseWriter) ResponseWritten() bool {
    return w.responseWritten
}

// Satisfy the http.ResponseWriter interface
func (w UseAwareResponseWriter) Header() http.Header {
    return w.ResponseWriter.Header()
}

func (w UseAwareResponseWriter) Write(data []byte) (int, error) {
    return w.ResponseWriter.Write(data)
}

func (w UseAwareResponseWriter) WriteHeader(statusCode int) {
    w.responseWritten = true
    // Write the status code onward.
    w.ResponseWriter.WriteHeader(statusCode)
}
