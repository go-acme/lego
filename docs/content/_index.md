---
title: "Lego"
date: 2019-03-03T16:39:46+01:00
draft: false
chapter: false
---

ACME client and ACME library written in Go.

{{% notice important %}}
lego is an independent, free, and open-source project, if you value it, consider [supporting it](https://donate.ldez.dev)! ❤️
{{% /notice %}}

## Features

{{< cards >}}

{{% card title="Challenges" href="obtain" %}}
- DNS-01
- HTTP-01
- TLS-ALPN-01
- DNS-PERSIST-01
{{% /card %}}

{{% card title="ACME servers" href="advanced/caservers" %}}
Multiple ACME servers support (Let's Encrypt, ZeroSSL, etc.)
{{% /card %}}

{{% card title="Certificate Management" href="obtain" %}}
Obtain, renew, revoke.

SAN certificate support.
{{% /card %}}

{{% card title="DNS providers" href="dns" %}}
Comes with more than 200 DNS providers
{{% /card %}}

{{% card title="CNAME" href="advanced/options/#lego_disable_cname_support" %}}
Supported by default.
{{% /card %}}

{{< /cards >}}

## Supported RFCs

| RFC                                                                                             | Description                                                               |
|-------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------|
| [RFC 8555](https://www.rfc-editor.org/rfc/rfc8555.html)                                         | Automatic Certificate Management Environment (ACME).                      |
| [RFC 8737](https://www.rfc-editor.org/rfc/rfc8737.html)                                         | TLS Application‑Layer Protocol Negotiation (ALPN) Challenge Extension.    |
| [RFC 8738](https://www.rfc-editor.org/rfc/rfc8738.html)                                         | IP Identifier Validation Extension. Issues certificates for IP addresses. |
| [RFC 9773](https://www.rfc-editor.org/rfc/rfc9773.html)                                         | Renewal Information (ARI) Extension.                                      |
| [draft-ietf-acme-profiles-01](https://datatracker.ietf.org/doc/draft-ietf-acme-profiles/)       | Profiles Extension.                                                       |
| [draft-ietf-acme-dns-persist-01](https://datatracker.ietf.org/doc/draft-ietf-acme-dns-persist/) | Challenge for Persistent DNS TXT Record Validation.                       |

## Supporting lego

Special thanks to the organizations sponsoring lego's development.

- [![FairSSL](./images/fairssl.svg?height=20px&classes=inline&lightbox=false)](https://www.fairssl.dk/)

- [![Canonical](./images/canonical.svg?height=30px&classes=inline&lightbox=false)](https://www.fairssl.dk/)
