FROM golang:alpine3.9 as builder

RUN apk --update upgrade \
    && apk --no-cache --no-progress add make git

WORKDIR /go/src/github.com/go-acme/lego
COPY . .
RUN make build

FROM alpine:3.9
RUN apk update \
    && apk add --no-cache ca-certificates tzdata \
    && update-ca-certificates

COPY --from=builder /go/src/github.com/go-acme/lego/dist/lego /usr/bin/lego
ENTRYPOINT [ "/usr/bin/lego" ]
