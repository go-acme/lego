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
  - Support [RFC 9773](https://www.rfc-editor.org/rfc/rfc9773.html): Renewal Information (ARI) Extension
  - Support [draft-aaron-acme-profiles-00](https://datatracker.ietf.org/doc/draft-aaron-acme-profiles/): Profiles Extension
- Comes with about [150 DNS providers]({{% ref "dns" %}})
- Register with CA
- Obtain certificates, both from scratch or with an existing CSR
- Renew certificates
- Revoke certificates
- Robust implementation of ACME challenges:
  - HTTP (http-01)
  - DNS (dns-01)
  - TLS (tls-alpn-01)
- SAN certificate support
- [CNAME support](https://letsencrypt.org/2019/10/09/onboarding-your-customers-with-lets-encrypt-and-acme.html) by default
- [Custom challenge solvers]({{% ref "usage/library/Writing-a-Challenge-Solver" %}})
- Certificate bundling
- OCSP helper function
