FROM golang:1.7
MAINTAINER Jon Miller <jon@siteidentify.com>

ADD install-proto.sh .
RUN ./install-proto.sh

RUN go get -u github.com/golang/protobuf/proto
RUN go get -u github.com/golang/protobuf/protoc-gen-go
RUN go get -u github.com/vicapow/go-vtile-example

EXPOSE 8080

WORKDIR /go/src/github.com/vicapow/go-vtile-example
CMD go run main.go

