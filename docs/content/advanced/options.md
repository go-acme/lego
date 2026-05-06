---
title: "Options"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 4
aliases:
  - usage/cli/options
---

This section describes advanced options that can be used to configure lego.

<!--more-->

## LEGO_CA_CERTIFICATES

The environment variable `LEGO_CA_CERTIFICATES` allows to specify the path to PEM-encoded CA certificates
that can be used to authenticate an ACME server with an HTTPS certificate not issued by a CA in the system-wide trusted root list.

Multiple file paths can be added by using `:` (unix) or `;` (Windows) as a separator.

Example:

```bash
# On Unix system
LEGO_CA_CERTIFICATES=/foo/cert1.pem:/foo/cert2.pem
```

## LEGO_CA_SYSTEM_CERT_POOL

The environment variable `LEGO_CA_SYSTEM_CERT_POOL` can be used to define if the certificates pool must use a copy of the system cert pool.

Example:

```bash
LEGO_CA_SYSTEM_CERT_POOL=true
```

## LEGO_CA_SERVER_NAME

The environment variable `LEGO_CA_SERVER_NAME` allows to specify the CA server name used to authenticate an ACME server
with an HTTPS certificate not issued by a CA in the system-wide trusted root list.

Example:

```bash
LEGO_CA_SERVER_NAME=foo
```

## LEGO_DISABLE_CNAME_SUPPORT

By default, lego follows CNAME, the environment variable `LEGO_DISABLE_CNAME_SUPPORT` allows to disable this support.

Example:

```bash
LEGO_DISABLE_CNAME_SUPPORT=true
```

There is a Let's Encrypt [blog post](https://letsencrypt.org/2019/10/09/onboarding-your-customers-with-lets-encrypt-and-acme.html) about the behavior of CNAMEs.

## LEGO_DEBUG_CLIENT_VERBOSE_ERROR

The environment variable `LEGO_DEBUG_CLIENT_VERBOSE_ERROR` allows to enrich error messages from some of the DNS clients.

Example:

```bash
LEGO_DEBUG_CLIENT_VERBOSE_ERROR=true
```

## LEGO_DEBUG_DNS_API_HTTP_CLIENT

> **⚠️ WARNING: This will expose credentials in the log output! ⚠️**
>
> Do not run this in production environments, or if you can't be sure that logs aren't accessed by third parties or tools (like log collectors).
>
> You have been warned. Here be dragons.

The environment variable `LEGO_DEBUG_DNS_API_HTTP_CLIENT` allows debugging the DNS API interaction.
It will dump the full request and response to the log output.

Some DNS providers don't support this option.

Example:

```bash
LEGO_DEBUG_DNS_API_HTTP_CLIENT=true
```

## LEGO_DEBUG_ACME_HTTP_CLIENT

The environment variable `LEGO_DEBUG_ACME_HTTP_CLIENT` allows debug the calls to the ACME server.

Example:

```bash
LEGO_DEBUG_ACME_HTTP_CLIENT=true
```
