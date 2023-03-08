FROM golang:1-alpine as builder

RUN apk --no-cache --no-progress add make git

WORKDIR /go/lego

ENV GO111MODULE on

# Download go modules
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN make build

FROM alpine:3
RUN apk update \
    && apk add --no-cache ca-certificates tzdata \
    && update-ca-certificates

COPY --from=builder /go/lego/dist/lego /usr/bin/lego

ENTRYPOINT [ "/usr/bin/lego" ]
