# Go library for accessing the Aurora DNS API

[![GoDoc](https://godoc.org/github.com/ldez/go-auroradns?status.svg)](https://godoc.org/github.com/ldez/go-auroradns)
[![Build Status](https://travis-ci.org/ldez/go-auroradns.svg?branch=master)](https://travis-ci.org/ldez/go-auroradns)
[![Go Report Card](https://goreportcard.com/badge/github.com/ldez/go-auroradns)](https://goreportcard.com/report/github.com/ldez/go-auroradns)

An Aurora DNS API client written in Go.

go-auroradns is a Go client library for accessing the Aurora DNS API.

## Available API methods

Zones:
- create
- delete
- list

Records:
- create
- delete
- list

## Example

```go
tr, _ := auroradns.NewTokenTransport("userID", "key")
client, _ := auroradns.NewClient(tr.Client())

zones, _, _ := client.GetZones()

fmt.Println(zones)
```

## API Documentation

- [API endpoint information](https://www.pcextreme.nl/community/d/111-what-is-the-api-endpoint-for-dns-health-checks)
- [API docs](https://libcloud.readthedocs.io/en/latest/dns/drivers/auroradns.html#api-docs)
