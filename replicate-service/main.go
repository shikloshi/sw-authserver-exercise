package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	//"expvar"

	"github.com/caarlos0/env"
)

var Config = struct {
	Backends                string `env:"BACKENDS" envDefault:"localhost:8081,localhost:8082"`
	MaxRetries              int    `env:"MAX_RETRIES" envDefault:"5"`
	FailedRequestsQueueSize int    `env:"FAILED_REQUESTS_QUEUE_SIZE" envDefault:"1000"`
	RetryWorkersCount       int    `env:"RETRY_WORKERS_COUNT" envDefault:"10"`
	InitialRetryWait        int    `env:"INIT_RETRY_WAIT" envDefault:"100"` // in ms
}{}

var failedRequests chan *http.Request
var backendAddresses []string

func main() {
	if err := env.Parse(&Config); err != nil {
		log.Fatal("Failed to parse configuration - %v", err)
	}

	var backends string
	flag.StringVar(&backends, "backends", Config.Backends, "Comma separated list of backend addresses")
	flag.Parse()

	//log.Printf("Running replicate-service with config - %v", Config)
	backendAddresses = strings.Split(backends, ",")
	log.Printf("Running replicate-service for backends- %v", backendAddresses)
	failedRequests = make(chan *http.Request, Config.FailedRequestsQueueSize)

	// Number of workers to handle request retry in the background
	for i := 0; i < Config.RetryWorkersCount; i++ {
		go failedRequestsWorker(failedRequests)
	}

	http.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func failedRequestsWorker(failedRequests chan *http.Request) {
	for req := range failedRequests {
		retryRequest(req)
	}
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	//responses := make(chan *http.Response, len(Config.Backends))
	responses := make(chan *http.Response)
	done := make(chan bool)
	// send copy of the requset to each configured backend
	for _, backendAddr := range backendAddresses {
		// getting lazy here: I'm not using channels that can cause some out of memory issues cuase number of goroutines are not bound
		go sendRequestUpstream(req, backendAddr, responses)
	}
	// here we should only have a one size chan, the first one should insert, all others can drop and error should go to the bla bla
	go sendFirstResponseDownstream(w, responses, done)
	<-done
	// not closing to avoid panic with sending other responses
	responses = nil
}

func sendFirstResponseDownstream(w http.ResponseWriter, backendResponses chan *http.Response, done chan bool) {
	// send first response back to client
	// for the rest we just need to close body
	// wait to only one blah blah here
	isFirstResponse := true
	//for resp := range backendResponses {
	resp := <-backendResponses
	// assuming at least one response returned okay
	if resp != nil {
		if isFirstResponse {
			log.Println("first response returned, writing back to client")
			isFirstResponse = false
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
			w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
			// TODO: copy all response headers to frontend response
			io.Copy(w, resp.Body)
			done <- true
		}
		//resp.Body.Close()
	}
}

func sendRequestUpstream(req *http.Request, backendAddr string, responses chan *http.Response) {
	log.Printf("Sending request to: [%s]", backendAddr)
	newReq, err := copyRequest(req, backendAddr)
	if err != nil {
		log.Printf("Could not copy request - %v\n", err)
	}
	//TODO: check new req nil
	c := &http.Client{}
	resp, err := c.Do(newReq)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Printf("error while sending request: %v, we are going to retry\n", err)
		// For simplicty: retrying failed request for all errors (which might not be that good)
		reqCopy, err := copyRequest(req, backendAddr)
		if err != nil {
			log.Printf("Could not copy request - %v\n", err)
		} else {
			failedRequests <- reqCopy
		}
		// close the body here
	} else {
		select {
		case responses <- resp:
		default:
			resp.Body.Close()
		}
	}
}

// very-very-very naive implementation fo exponential backoff
func retryRequest(r *http.Request) {
	log.Println("Going to retry request with exponential back off")
	wait := Config.InitialRetryWait
	c := &http.Client{}
	for i := 0; i < Config.MaxRetries; i++ {
		// ignoring err her
		resp, _ := c.Do(r)
		if resp != nil && resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			log.Printf("Retry succeeded after: %d retries\n", i)
			break
		}
		wait *= 2
		log.Printf("Retry failed, waiting for: %d ms before retrying again\n", wait)
		time.Sleep(time.Duration(wait) * time.Millisecond)
	}
}

func copyRequest(req *http.Request, backendAddr string) (*http.Request, error) {
	url := fmt.Sprintf("http://%s%s", backendAddr, req.RequestURI)
	newReq, err := http.NewRequest(req.Method, url, req.Body)
	if err != nil {
		return nil, err
	}
	// Shallow copying the headers due to the fact we are not going to change them for now
	newReq.Header = req.Header
	return newReq, nil
}
