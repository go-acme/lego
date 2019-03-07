# lego

Let's Encrypt client and ACME library written in Go.

[![GoDoc](https://godoc.org/github.com/xenolf/lego?status.svg)](https://godoc.org/github.com/xenolf/lego/acme)
[![Build Status](https://travis-ci.org/xenolf/lego.svg?branch=master)](https://travis-ci.org/xenolf/lego)
[![Docker Pulls](https://img.shields.io/docker/pulls/xenolf/lego.svg)](https://hub.docker.com/r/xenolf/lego/)
[![Dev Chat](https://img.shields.io/badge/dev%20chat-gitter-blue.svg?label=dev+chat)](https://gitter.im/xenolf/lego)
[![Beerpay](https://beerpay.io/xenolf/lego/badge.svg)](https://beerpay.io/xenolf/lego)

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
- Comes with multiple optional [DNS providers](https://github.com/xenolf/lego/tree/master/providers/dns)
- [Custom challenge solvers](https://github.com/xenolf/lego/wiki/Writing-a-Challenge-Solver)
- Certificate bundling
- OCSP helper function

lego introduced support for ACME v2 in [v1.0.0](https://github.com/xenolf/lego/releases/tag/v1.0.0). If you still need to utilize ACME v1, you can do so by using the [v0.5.0](https://github.com/xenolf/lego/releases/tag/v0.5.0) version.

## Installation

How to [install](https://xenolf.github.io/lego/installation/).

## Usage

- as a [CLI](https://xenolf.github.io/lego/usage/cli)
- as a [library](https://xenolf.github.io/lego/usage/lib)

## Documentation

Documentation is hosted live at https://xenolf.github.io/lego/.

## DNS providers

Detailed documentation is available [here](https://xenolf.github.io/lego/dns).

|                                                                |                                                                                |                                                                   |                                                                  |
|----------------------------------------------------------------|--------------------------------------------------------------------------------|-------------------------------------------------------------------|------------------------------------------------------------------|
| [Alibaba Cloud DNS](https://xenolf.github.io/lego/dns/alidns/) | [Amazon Lightsail](https://xenolf.github.io/lego/dns/lightsail/)               | [Amazon Route 53](https://xenolf.github.io/lego/dns/route53/)     | [Aurora DNS](https://xenolf.github.io/lego/dns/auroradns/)       |
| [Azure](https://xenolf.github.io/lego/dns/azure/)              | [Bluecat](https://xenolf.github.io/lego/dns/bluecat/)                          | [ClouDNS](https://xenolf.github.io/lego/dns/cloudns/)             | [CloudXNS](https://xenolf.github.io/lego/dns/cloudxns/)          |
| [Cloudflare](https://xenolf.github.io/lego/dns/cloudflare/)    | [ConoHa](https://xenolf.github.io/lego/dns/conoha/)                            | [DNS Made Easy](https://xenolf.github.io/lego/dns/dnsmadeeasy/)   | [DNSPod](https://xenolf.github.io/lego/dns/dnspod/)              |
| [DNSimple](https://xenolf.github.io/lego/dns/dnsimple/)        | [Designate DNSaaS for Openstack](https://xenolf.github.io/lego/dns/designate/) | [Digital Ocean](https://xenolf.github.io/lego/dns/digitalocean/)  | [DreamHost](https://xenolf.github.io/lego/dns/dreamhost/)        |
| [Duck DNS](https://xenolf.github.io/lego/dns/duckdns/)         | [Dyn](https://xenolf.github.io/lego/dns/dyn/)                                  | [Exoscale](https://xenolf.github.io/lego/dns/exoscale/)           | [External program](https://xenolf.github.io/lego/dns/exec/)      |
| [FastDNS](https://xenolf.github.io/lego/dns/fastdns/)          | [Gandi](https://xenolf.github.io/lego/dns/gandi/)                              | [Gandi Live DNS (v5)](https://xenolf.github.io/lego/dns/gandiv5/) | [Glesys](https://xenolf.github.io/lego/dns/glesys/)              |
| [Go Daddy](https://xenolf.github.io/lego/dns/godaddy/)         | [Google Cloud](https://xenolf.github.io/lego/dns/gcloud/)                      | [HTTP request](https://xenolf.github.io/lego/dns/httpreq/)        | [Hosting.de](https://xenolf.github.io/lego/dns/hostingde/)       |
| [INWX](https://xenolf.github.io/lego/dns/inwx/)                | [Internet Initiative Japan](https://xenolf.github.io/lego/dns/iij/)            | [Joohoi's ACME-DNS](https://xenolf.github.io/lego/dns/acme-dns)   | [Linode (deprecated)](https://xenolf.github.io/lego/dns/linode/) |
| [Linode (v4)](https://xenolf.github.io/lego/dns/linodev4/)     | [Manual](https://xenolf.github.io/lego/dns/manual/)                            | [MyDNS.jp](https://xenolf.github.io/lego/dns/mydnsjp/)            | [NIFCloud](https://xenolf.github.io/lego/dns/nifcloud/)          |
| [NS1](https://xenolf.github.io/lego/dns/ns1/)                  | [Name.com](https://xenolf.github.io/lego/dns/namedotcom/)                      | [Namecheap](https://xenolf.github.io/lego/dns/namecheap/)         | [Netcup](https://xenolf.github.io/lego/dns/netcup/)              |
| [OVH](https://xenolf.github.io/lego/dns/ovh/)                  | [Open Telekom Cloud](https://xenolf.github.io/lego/dns/otc/)                   | [Oracle Cloud](https://xenolf.github.io/lego/dns/oraclecloud/)    | [PowerDNS](https://xenolf.github.io/lego/dns/pdns/)              |
| [RFC2136](https://xenolf.github.io/lego/dns/rfc2136/)          | [Rackspace](https://xenolf.github.io/lego/dns/rackspace/)                      | [Sakura Cloud](https://xenolf.github.io/lego/dns/sakuracloud/)    | [Selectel](https://xenolf.github.io/lego/dns/selectel/)          |
| [Stackpath](https://xenolf.github.io/lego/dns/stackpath/)      | [TransIP](https://xenolf.github.io/lego/dns/transip/)                          | [VegaDNS](https://xenolf.github.io/lego/dns/vegadns/)             | [Vscale](https://xenolf.github.io/lego/dns/vscale/)              |
| [Vultr](https://xenolf.github.io/lego/dns/vultr/)              | [Zone.ee](https://xenolf.github.io/lego/dns/zoneee/)                           |                                                                   |                                                                  |
