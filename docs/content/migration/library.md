---
title: "Library Guide"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 2
---

This guide describes the changes between the v4 and v5 versions of the library.

<!--more-->

## Context

Most of the functions and methods are now using a context.

Example:

```go
// Before
client.Certificate.Obtain(request)
```

```go
// After
client.Certificate.Obtain(context.TODO(), request)
```

## Logger

The logger is now `slog` and can be set using the `log.SetDefault(logger)` function.

## Method and function changes

| v4                             | v5                             |
|--------------------------------|--------------------------------|
| `crypto.GenerateCSR`           | `crypto.CreateCSR`             |
| `crypto.GetKeyType`            | `crypto.ToKeyType`             |
| `Certifier.RenewWithOption`    | `Certifier.Renew`              |
| `OrderService.NewWithOptions`  | `OrderService.New`             |
| `acmedns.NewDNSProviderClient` | `acmedns.NewDNSProviderConfig` |
| `scaleway.Config.Token`        | `scaleway.Config.SecretKey`    |

The functions and methods related to the private key are now using the `crypto.Signer` interface instead of the `crypto.PrivateKey` type.

The following methods now return an `*acme.ExtendedAccount` instead of an `*registration.Ressouce`.

- `registration.Registrar.Register`
- `registration.Registrar.RegisterWithExternalAccountBinding`
- `registration.Registrar.QueryRegistration`
- `registration.Registrar.UpdateRegistration`
- `registration.Registrar.ResolveAccountByKey`

The structure `registration.Ressouce` has been removed.

The method `http01.ProviderServer.SetProxyHeader` is removed and replaced by an option `http01.Options.ProxyHeaderName`.

## Fields changes

The field `RetryAfter` of `acme.RateLimitedError` and `acme.ExtendedChallenge` is now a `time.Duration` instead of a `string`.

## CertifierOptions

### CommonName

The support of the common name is disabled by default.

The field `DisableCommonName` of `certificate.CertifierOptions` has been removed.

The option is now determined by the `EnableCommonName` field of the `certificate.ObtainRequest` and `certificate.ObtainForCSRRequest`.

### KeyType

The field `KeyType` of `certificate.CertifierOptions` has been removed.

The key type is now determined by the `KeyType` field of the `certificate.ObtainRequest`.

## certcrypto.KeyType

The string values of the `certcrypto.KeyType` enum have been changed:

| v4       | v5        |
|----------|-----------|
| `P256`   | `EC256`   |
| `P384`   | `EC384`   |
| `2048`   | `RSA2048` |
| `3072`   | `RSA3072` |
| `4096`   | `RSA4096` |
| `8192`   | `RSA8192` |

## Removed elements

The following elements have been removed without replacements:

- `selectel.Config.BaseURL`
- `selectel.EnvBaseURL`
- `SELECTEL_BASE_URL`
- `vscale.Config.BaseURL`
- `vscale.EnvBaseURL`
- `VSCALE_BASE_URL`
- `ipv64.Config.SequenceInterval`
- `netcup.Config.TTL`
- `netcup.EnvTTL`
- `vultr.Config.HTTPTimeout`

## PEM encoding

It uses `PKCS#8` instead of `PKCS#1` for PEM encoding.
