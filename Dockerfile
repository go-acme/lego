FROM alpine:3.5

ENV GOPATH /go
ENV SYSROOT /go

RUN apk update && apk add ca-certificates go git musl-dev && \
    rm -rf /var/cache/apk/* && \
    go get -u github.com/xenolf/lego && \
    cd /go/src/github.com/xenolf/lego && \
    go build -o /usr/bin/lego . && \
    apk del go git && \
    rm -rf /var/cache/apk/* && \
    rm -rf /go

ENTRYPOINT [ "/usr/bin/lego" ]
