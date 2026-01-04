---
title: "Lego"
date: 2019-03-03T16:39:46+01:00
draft: false
chapter: false
---

Let's Encrypt client and ACME library written in Go.

{{% notice important %}}
lego is an independent, free, and open-source project, if you value it, consider [supporting it](https://donate.ldez.dev)! ❤️

This project is not owned by a company. I'm not an employee of a company.

I don't have gifted domains/accounts from DNS companies.

I've been maintaining it for about 10 years.
{{% /notice %}}

## Features

- ACME v2 [RFC 8555](https://www.rfc-editor.org/rfc/rfc8555.html)
  - Support [RFC 8737](https://www.rfc-editor.org/rfc/rfc8737.html): TLS Application‑Layer Protocol Negotiation (ALPN) Challenge Extension
  - Support [RFC 8738](https://www.rfc-editor.org/rfc/rfc8738.html): issues certificates for IP addresses
  - Support [RFC 9773](https://www.rfc-editor.org/rfc/rfc9773.html): Renewal Information (ARI) Extension
  - Support [draft-ietf-acme-profiles-00](https://datatracker.ietf.org/doc/draft-ietf-acme-profiles/): Profiles Extension
- Comes with about [180 DNS providers]({{% ref "dns" %}})
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
