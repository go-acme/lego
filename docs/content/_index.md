---
title: "Lego"
date: 2019-03-03T16:39:46+01:00
draft: false
chapter: false
---

Let's Encrypt client and ACME library written in Go.

## Features

- ACME v2 [RFC 8555](https://www.rfc-editor.org/rfc/rfc8555.html)
  - Support [RFC 8737](https://www.rfc-editor.org/rfc/rfc8737.html): TLS Application‑Layer Protocol Negotiation (ALPN) Challenge Extension
  - Support [RFC 8738](https://www.rfc-editor.org/rfc/rfc8738.html): issues certificates for IP addresses
  - Support [draft-ietf-acme-ari-01](https://datatracker.ietf.org/doc/draft-ietf-acme-ari/): Renewal Information (ARI) Extension
- Register with CA
- Obtain certificates, both from scratch or with an existing CSR
- Renew certificates
- Revoke certificates
- Robust implementation of all ACME challenges
  - HTTP (http-01)
  - DNS (dns-01)
  - TLS (tls-alpn-01)
- SAN certificate support
- [CNAME support](https://letsencrypt.org/2019/10/09/onboarding-your-customers-with-lets-encrypt-and-acme.html) by default
- Comes with multiple optional [DNS providers]({{% ref "dns" %}})
- [Custom challenge solvers]({{% ref "usage/library/Writing-a-Challenge-Solver" %}})
- Certificate bundling
- OCSP helper function
