FROM golang:1.10

EXPOSE 8080

COPY . /go/src/github.com/shikloshi/SimiliarWeb/app
WORKDIR /go/src/github.com/shikloshi/SimiliarWeb/app

RUN go get
RUN go build -o replicator

CMD ["./replicator"]
