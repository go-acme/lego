#
# First, export the ops.asc key locally.
#
#   gpg --export-secret-key E458F9F85608DF5A22ECCD158B58C61D4FFE0C86 > ops.asc
#
# Build the container
#
#   docker build -t egoscale .
#
# Prepare a snapshot release
#
#   docker run -v $PWD:/go/src/github.com/exoscale/egoscale egoscale goreleaser --snapshot
#
# Publish egoscale exposing a valid GITHUB_TOKEN
#
#   git tag -a v0.10
#   git push --tag
#   docker run -v $PWD:/go/src/github.com/exoscale/egoscale -e GITHUB_TOKEN=... egoscale goreleaser
#
#
# ⚠ do not push this container anywhere ⚠
#
FROM golang:1.10-stretch

ARG DEBIAN_FRONTEND=noninteractive

RUN go get -u github.com/golang/dep/cmd/dep \
 && go get -u -d github.com/goreleaser/goreleaser/... \
 && go get -u -d github.com/goreleaser/nfpm/... \
 && apt-get update -q \
 && apt-get upgrade -qy \
 && apt-get install -qy \
        rpm \
 && cd $GOPATH/src/github.com/goreleaser/nfpm \
 && dep ensure -v -vendor-only \
 && go install \
 && cd ../goreleaser \
 && dep ensure -v -vendor-only \
 && go install \
 && cd /

ADD ops.asc ops.asc
RUN gpg --allow-secret-key-import --import ops.asc

VOLUME /go/src/github.com/exoscale/egoscale
WORKDIR /go/src/github.com/exoscale/egoscale

CMD ['goreleaser', '--snapshot']
