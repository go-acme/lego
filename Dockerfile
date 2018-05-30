FROM golang:alpine3.7 as builder

ARG LEGO_VERSION=dev

WORKDIR /go/src/github.com/xenolf/lego
COPY . .
RUN go build -ldflags="-s -X main.version=${LEGO_VERSION}"

FROM alpine:3.7
RUN apk update && apk add --no-cache --virtual ca-certificates
COPY --from=builder /go/src/github.com/xenolf/lego/lego /usr/bin/lego
ENTRYPOINT [ "/usr/bin/lego" ]
