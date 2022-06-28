---
title: "Welcome"
date: 2019-03-03T16:39:46+01:00
draft: false
chapter: true
---

# Lego

Let's Encrypt client and ACME library written in Go.

## Features

- ACME v2 [RFC 8555](https://www.rfc-editor.org/rfc/rfc8555.html)
- Register with CA
- Obtain certificates, both from scratch or with an existing CSR
- Renew certificates
- Revoke certificates
- Robust implementation of all ACME challenges
  - HTTP (http-01)
  - DNS (dns-01)
  - TLS (tls-alpn-01)
- SAN certificate support
- Comes with multiple optional [DNS providers]({{< ref "dns" >}})
- [Custom challenge solvers]({{< ref "usage/library/Writing-a-Challenge-Solver" >}})
- Certificate bundling
- OCSP helper function
