# lego

Let's Encrypt client and ACME library written in Go.

[![GoDoc](https://godoc.org/github.com/go-acme/lego?status.svg)](https://godoc.org/github.com/go-acme/lego/acme)
[![Build Status](https://travis-ci.com/go-acme/lego.svg?branch=master)](https://travis-ci.com/go-acme/lego)
[![Docker Pulls](https://img.shields.io/docker/pulls/goacme/lego.svg)](https://hub.docker.com/r/goacme/lego/)

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
- Comes with multiple optional [DNS providers](https://go-acme.github.io/lego/dns)
- [Custom challenge solvers](https://go-acme.github.io/lego/usage/library/writing-a-challenge-solver/)
- Certificate bundling
- OCSP helper function

lego introduced support for ACME v2 in [v1.0.0](https://github.com/go-acme/lego/releases/tag/v1.0.0). If you still need to utilize ACME v1, you can do so by using the [v0.5.0](https://github.com/go-acme/lego/releases/tag/v0.5.0) version.

## Installation

How to [install](https://go-acme.github.io/lego/installation/).

## Usage

- as a [CLI](https://go-acme.github.io/lego/usage/cli)
- as a [library](https://go-acme.github.io/lego/usage/library)

## Documentation

Documentation is hosted live at https://go-acme.github.io/lego/.

## DNS providers

Detailed documentation is available [here](https://go-acme.github.io/lego/dns).

|                                                                                 |                                                                                 |                                                                                 |                                                                                 |
|---------------------------------------------------------------------------------|---------------------------------------------------------------------------------|---------------------------------------------------------------------------------|---------------------------------------------------------------------------------|
| [Alibaba Cloud DNS](https://go-acme.github.io/lego/dns/alidns/)                 | [Amazon Lightsail](https://go-acme.github.io/lego/dns/lightsail/)               | [Amazon Route 53](https://go-acme.github.io/lego/dns/route53/)                  | [Aurora DNS](https://go-acme.github.io/lego/dns/auroradns/)                     |
| [Azure](https://go-acme.github.io/lego/dns/azure/)                              | [Bindman](https://go-acme.github.io/lego/dns/bindman/)                          | [Bluecat](https://go-acme.github.io/lego/dns/bluecat/)                          | [Cloudflare](https://go-acme.github.io/lego/dns/cloudflare/)                    |
| [ClouDNS](https://go-acme.github.io/lego/dns/cloudns/)                          | [CloudXNS](https://go-acme.github.io/lego/dns/cloudxns/)                        | [ConoHa](https://go-acme.github.io/lego/dns/conoha/)                            | [Designate DNSaaS for Openstack](https://go-acme.github.io/lego/dns/designate/) |
| [Digital Ocean](https://go-acme.github.io/lego/dns/digitalocean/)               | [DNS Made Easy](https://go-acme.github.io/lego/dns/dnsmadeeasy/)                | [DNSimple](https://go-acme.github.io/lego/dns/dnsimple/)                        | [DNSPod](https://go-acme.github.io/lego/dns/dnspod/)                            |
| [Domain Offensive (do.de)](https://go-acme.github.io/lego/dns/dode/)            | [DreamHost](https://go-acme.github.io/lego/dns/dreamhost/)                      | [Duck DNS](https://go-acme.github.io/lego/dns/duckdns/)                         | [Dyn](https://go-acme.github.io/lego/dns/dyn/)                                  |
| [EasyDNS](https://go-acme.github.io/lego/dns/easydns/)                          | [Exoscale](https://go-acme.github.io/lego/dns/exoscale/)                        | [External program](https://go-acme.github.io/lego/dns/exec/)                    | [FastDNS](https://go-acme.github.io/lego/dns/fastdns/)                          |
| [Gandi Live DNS (v5)](https://go-acme.github.io/lego/dns/gandiv5/)              | [Gandi](https://go-acme.github.io/lego/dns/gandi/)                              | [Glesys](https://go-acme.github.io/lego/dns/glesys/)                            | [Go Daddy](https://go-acme.github.io/lego/dns/godaddy/)                         |
| [Google Cloud](https://go-acme.github.io/lego/dns/gcloud/)                      | [Hosting.de](https://go-acme.github.io/lego/dns/hostingde/)                     | [HTTP request](https://go-acme.github.io/lego/dns/httpreq/)                     | [Internet Initiative Japan](https://go-acme.github.io/lego/dns/iij/)            |
| [INWX](https://go-acme.github.io/lego/dns/inwx/)                                | [Joker](https://go-acme.github.io/lego/dns/joker/)                              | [Joohoi's ACME-DNS](https://go-acme.github.io/lego/dns/acme-dns)                | [Linode (deprecated)](https://go-acme.github.io/lego/dns/linode/)               |
| [Linode (v4)](https://go-acme.github.io/lego/dns/linodev4/)                     | [Manual](https://go-acme.github.io/lego/dns/manual/)                            | [MyDNS.jp](https://go-acme.github.io/lego/dns/mydnsjp/)                         | [Name.com](https://go-acme.github.io/lego/dns/namedotcom/)                      |
| [Namecheap](https://go-acme.github.io/lego/dns/namecheap/)                      | [Namesilo](https://go-acme.github.io/lego/dns/namesilo/)                        | [Netcup](https://go-acme.github.io/lego/dns/netcup/)                            | [NIFCloud](https://go-acme.github.io/lego/dns/nifcloud/)                        |
| [NS1](https://go-acme.github.io/lego/dns/ns1/)                                  | [Open Telekom Cloud](https://go-acme.github.io/lego/dns/otc/)                   | [Oracle Cloud](https://go-acme.github.io/lego/dns/oraclecloud/)                 | [OVH](https://go-acme.github.io/lego/dns/ovh/)                                  |
| [PowerDNS](https://go-acme.github.io/lego/dns/pdns/)                            | [Rackspace](https://go-acme.github.io/lego/dns/rackspace/)                      | [RFC2136](https://go-acme.github.io/lego/dns/rfc2136/)                          | [Sakura Cloud](https://go-acme.github.io/lego/dns/sakuracloud/)                 |
| [Selectel](https://go-acme.github.io/lego/dns/selectel/)                        | [Stackpath](https://go-acme.github.io/lego/dns/stackpath/)                      | [TransIP](https://go-acme.github.io/lego/dns/transip/)                          | [VegaDNS](https://go-acme.github.io/lego/dns/vegadns/)                          |
| [Vscale](https://go-acme.github.io/lego/dns/vscale/)                            | [Versio](https://go-acme.github.io/lego/dns/versio/)                            | [Vultr](https://go-acme.github.io/lego/dns/vultr/)                              | [Zone.ee](https://go-acme.github.io/lego/dns/zoneee/)                           |
