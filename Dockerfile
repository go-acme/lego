FROM golang:alpine3.7 as builder

ARG LEGO_VERSION=master

RUN apk update && \
    apk add --no-cache --virtual git && \
    go get -u github.com/xenolf/lego && \
    cd ${GOPATH}/src/github.com/xenolf/lego && \
    git checkout ${LEGO_VERSION} && \
    go build -o /usr/bin/lego .

FROM alpine:3.7
RUN apk update && apk add --no-cache --virtual ca-certificates
COPY --from=builder /usr/bin/lego /usr/bin/lego

ENTRYPOINT [ "/usr/bin/lego" ]
