---
title: "CA servers"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 5
slug: caservers
---

This page describes the usage of CA servers (ACME servers).

<!--more-->

{{% notice note %}}
Any CA server that follow [RFC 8855](https://www.rfc-editor.org/rfc/rfc8555.html) can be used with lego.
{{% /notice %}}

## Let's Encrypt ACME server

lego defaults to communicating with the production Let's Encrypt ACME server.

If you'd like to test something without issuing real certificates, consider using the staging endpoint instead:

```bash
lego run --server='letsencrypt-staging' …
```

## CA Server short-codes

To ease the usage of the CA server in most of cases, we provide a short-code for each already known CA server.

| Name | Code | Directory URL |
|---|---|---|
| Actalis  | `actalis`  | https://acme-api.actalis.com/acme/directory  |
| Digicert  | `digicert`  | https://one.digicert.com/mpki/api/v1/acme/v2/directory  |
| FreeSSL  | `freessl`  | https://acmepro.freessl.cn/v2/DV  |
| GlobalSign  | `globalsign`  | https://emea.acme.atlas.globalsign.com/directory  |
| Google Trust  | `googletrust`  | https://dv.acme-v02.api.pki.goog/directory  |
| Google Trust staging  | `googletrust-staging`  | https://dv.acme-v02.test-api.pki.goog/directory  |
| Let's Encrypt  | `letsencrypt`  | https://acme-v02.api.letsencrypt.org/directory  |
| Let's Encrypt staging  | `letsencrypt-staging`  | https://acme-staging-v02.api.letsencrypt.org/directory  |
| LiteSSL  | `litessl`  | https://acme.litessl.com/acme/v2/directory  |
| PeeringHub  | `peeringhub`  | https://stica.peeringhub.io/acme  |
| SSL.com ECDSA  | `sslcomecc`  | https://acme.ssl.com/sslcom-dv-ecc  |
| SSL.com RSA  | `sslcomrsa`  | https://acme.ssl.com/sslcom-dv-rsa  |
| Sectigo DV (Domain Validation)  | `sectigo`  | https://acme.sectigo.com/v2/DV  |
| Sectigo EV (Extended Validation)  | `sectigoev`  | https://acme.sectigo.com/v2/EV  |
| Sectigo OV (Organization Validation)   | `sectigoov`  | https://acme.sectigo.com/v2/OV  |
| ZeroSSL  | `zerossl`  | https://acme.zerossl.com/v2/DV90  |


## ZeroSSL

lego supports three different ways to authenticate with ZeroSSL.

1. Access key: if the environment variable `ZERO_SSL_ACCESS_KEY` is set.
2. Email: if the email address is set and the environment variable `ZERO_SSL_ACCESS_KEY` is not set.
3. External Account Binding (EAB): if none of the above elements are defined.
