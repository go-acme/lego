---
title: "Welcome"
date: 2019-03-03T16:39:46+01:00
draft: false
chapter: true
---

# Lego

Let's Encrypt client and ACME library written in Go.

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
- Comes with multiple optional [DNS providers](dns)
- [Custom challenge solvers](usage/library/writing-a-challenge-solver/)
- Certificate bundling
- OCSP helper function


lego introduced support for ACME v2 in [v1.0.0](https://github.com/go-acme/lego/releases/tag/v1.0.0).  
If you still need to utilize ACME v1, you can do so by using the [v0.5.0](https://github.com/go-acme/lego/releases/tag/v0.5.0) version.
