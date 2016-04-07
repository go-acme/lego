FROM alpine:3.3

ENV GOPATH /go

RUN apk update && apk add ca-certificates go git && \
    rm -rf /var/cache/apk/*

COPY . /go/src/github.com/xenolf/lego

RUN cd /go/src/github.com/xenolf/lego && \
	go get ./... && \
	go build -o /usr/bin/lego . && \
	apk del ca-certificates go git && \
	rm -rf /var/cache/apk/* && \
	rm -rf /go

ENTRYPOINT [ "/usr/bin/lego" ]
