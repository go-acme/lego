FROM alpine:3.6

ENV GOPATH /go

RUN apk update && apk add ca-certificates go git musl-dev && \
    rm -rf /var/cache/apk/* && \
    go get -u github.com/xenolf/lego && \
    cd /go/src/github.com/xenolf/lego && \
    git checkout tags/v0.4.0 && \
    go build -o /usr/bin/lego . && \
    apk del go git musl-dev && \
    rm -rf /var/cache/apk/* && \
    rm -rf /go

ENTRYPOINT [ "/usr/bin/lego" ]
