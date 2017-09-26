FROM alpine:3.6

ENV GOPATH /go

RUN apk update && apk --no-cache add \
        ca-certificates \
        go \
        git \
        libc-dev \
    && go get github.com/xenolf/lego \
    && mv /go/bin/lego /usr/bin/lego \
    && rm -rf /go \
    && apk del \
        go \
        git \
        libc-dev

ENTRYPOINT [ "/usr/bin/lego" ]
