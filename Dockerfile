FROM golang:1.12-alpine3.10 as builder

RUN apk --no-cache --no-progress add make git

WORKDIR /go/src/github.com/go-acme/lego
COPY . .
RUN make build

FROM alpine:3.10
RUN apk update \
    && apk add --no-cache ca-certificates tzdata \
    && update-ca-certificates

COPY --from=builder /go/src/github.com/go-acme/lego/dist/lego /usr/bin/lego
ENTRYPOINT [ "/usr/bin/lego" ]
