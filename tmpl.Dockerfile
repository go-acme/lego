# Dockerfile template used by Seihon to create multi-arch images.
# https://github.com/ldez/seihon
FROM golang:1-alpine as builder

RUN apk --update upgrade \
    && apk --no-cache --no-progress add git make ca-certificates tzdata

WORKDIR /go/lego

ENV GO111MODULE on

# Download go modules
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN GOARCH={{ .GoARCH }} GOARM={{ .GoARM }} make build

FROM {{ .RuntimeImage }}

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/lego/dist/lego /usr/bin/lego

ENTRYPOINT [ "/usr/bin/lego" ]
