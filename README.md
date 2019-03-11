# lego

Let's Encrypt client and ACME library written in Go.

[![GoDoc](https://godoc.org/github.com/go-acme/lego?status.svg)](https://godoc.org/github.com/go-acme/lego/acme)
[![Build Status](https://travis-ci.org/go-acme/lego.svg?branch=master)](https://travis-ci.org/go-acme/lego)
[![Docker Pulls](https://img.shields.io/docker/pulls/go-acme/lego.svg)](https://hub.docker.com/r/go-acme/lego/)

## Features

- Register with CA
- Obtain certificates, both from scratch or with an existing CSR
- Renew certificates
- Revoke certificates
- Robust implementation of all ACME challenges
  - HTTP (http-01)
  - DNS (dns-01)
  - TLS (tls-alpn-01)
- SAN certificate support
- Comes with multiple optional [DNS providers](hhttps://go-acme.github.io/lego/dns)
- [Custom challenge solvers](https://go-acme.github.io/lego/usage/library/writing-a-challenge-solver/)
- Certificate bundling
- OCSP helper function

lego introduced support for ACME v2 in [v1.0.0](https://github.com/go-acme/lego/releases/tag/v1.0.0). If you still need to utilize ACME v1, you can do so by using the [v0.5.0](https://github.com/go-acme/lego/releases/tag/v0.5.0) version.

## Installation

How to [install](https://go-acme.github.io/lego/installation/).

## Usage

- as a [CLI](https://go-acme.github.io/lego/usage/cli)
- as a [library](https://go-acme.github.io/lego/usage/lib)

## Documentation

Documentation is hosted live at https://go-acme.github.io/lego/.

## DNS providers

Detailed documentation is available [here](https://go-acme.github.io/lego/dns).

|                                                                |                                                                                |                                                                   |                                                                  |
|----------------------------------------------------------------|--------------------------------------------------------------------------------|-------------------------------------------------------------------|------------------------------------------------------------------|
| [Alibaba Cloud DNS](https://go-acme.github.io/lego/dns/alidns/) | [Amazon Lightsail](https://go-acme.github.io/lego/dns/lightsail/)               | [Amazon Route 53](https://go-acme.github.io/lego/dns/route53/)     | [Aurora DNS](https://go-acme.github.io/lego/dns/auroradns/)       |
| [Azure](https://go-acme.github.io/lego/dns/azure/)              | [Bluecat](https://go-acme.github.io/lego/dns/bluecat/)                          | [ClouDNS](https://go-acme.github.io/lego/dns/cloudns/)             | [CloudXNS](https://go-acme.github.io/lego/dns/cloudxns/)          |
| [Cloudflare](https://go-acme.github.io/lego/dns/cloudflare/)    | [ConoHa](https://go-acme.github.io/lego/dns/conoha/)                            | [DNS Made Easy](https://go-acme.github.io/lego/dns/dnsmadeeasy/)   | [DNSPod](https://go-acme.github.io/lego/dns/dnspod/)              |
| [DNSimple](https://go-acme.github.io/lego/dns/dnsimple/)        | [Designate DNSaaS for Openstack](https://go-acme.github.io/lego/dns/designate/) | [Digital Ocean](https://go-acme.github.io/lego/dns/digitalocean/)  | [DreamHost](https://go-acme.github.io/lego/dns/dreamhost/)        |
| [Duck DNS](https://go-acme.github.io/lego/dns/duckdns/)         | [Dyn](https://go-acme.github.io/lego/dns/dyn/)                                  | [Exoscale](https://go-acme.github.io/lego/dns/exoscale/)           | [External program](https://go-acme.github.io/lego/dns/exec/)      |
| [FastDNS](https://go-acme.github.io/lego/dns/fastdns/)          | [Gandi](https://go-acme.github.io/lego/dns/gandi/)                              | [Gandi Live DNS (v5)](https://go-acme.github.io/lego/dns/gandiv5/) | [Glesys](https://go-acme.github.io/lego/dns/glesys/)              |
| [Go Daddy](https://go-acme.github.io/lego/dns/godaddy/)         | [Google Cloud](https://go-acme.github.io/lego/dns/gcloud/)                      | [HTTP request](https://go-acme.github.io/lego/dns/httpreq/)        | [Hosting.de](https://go-acme.github.io/lego/dns/hostingde/)       |
| [INWX](https://go-acme.github.io/lego/dns/inwx/)                | [Internet Initiative Japan](https://go-acme.github.io/lego/dns/iij/)            | [Joohoi's ACME-DNS](https://go-acme.github.io/lego/dns/acme-dns)   | [Linode (deprecated)](https://go-acme.github.io/lego/dns/linode/) |
| [Linode (v4)](https://go-acme.github.io/lego/dns/linodev4/)     | [Manual](https://go-acme.github.io/lego/dns/manual/)                            | [MyDNS.jp](https://go-acme.github.io/lego/dns/mydnsjp/)            | [NIFCloud](https://go-acme.github.io/lego/dns/nifcloud/)          |
| [NS1](https://go-acme.github.io/lego/dns/ns1/)                  | [Name.com](https://go-acme.github.io/lego/dns/namedotcom/)                      | [Namecheap](https://go-acme.github.io/lego/dns/namecheap/)         | [Netcup](https://go-acme.github.io/lego/dns/netcup/)              |
| [OVH](https://go-acme.github.io/lego/dns/ovh/)                  | [Open Telekom Cloud](https://go-acme.github.io/lego/dns/otc/)                   | [Oracle Cloud](https://go-acme.github.io/lego/dns/oraclecloud/)    | [PowerDNS](https://go-acme.github.io/lego/dns/pdns/)              |
| [RFC2136](https://go-acme.github.io/lego/dns/rfc2136/)          | [Rackspace](https://go-acme.github.io/lego/dns/rackspace/)                      | [Sakura Cloud](https://go-acme.github.io/lego/dns/sakuracloud/)    | [Selectel](https://go-acme.github.io/lego/dns/selectel/)          |
| [Stackpath](https://go-acme.github.io/lego/dns/stackpath/)      | [TransIP](https://go-acme.github.io/lego/dns/transip/)                          | [VegaDNS](https://go-acme.github.io/lego/dns/vegadns/)             | [Vscale](https://go-acme.github.io/lego/dns/vscale/)              |
| [Vultr](https://go-acme.github.io/lego/dns/vultr/)              | [Zone.ee](https://go-acme.github.io/lego/dns/zoneee/)                           |                                                                   |                                                                  |
