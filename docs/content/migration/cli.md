---
title: "CLI Guide"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 1
---

This guide describes the changes between the v4 and v5 versions of the command line.

<!--more-->

{{% notice style="caution" %}}

Before upgrading to v5, please run the command `lego migrate`.

This command will migrate the file structure to the new one.

**This is a requirement.**

This command will not work if you were using the deprecated `--filename` flag.

If you need help, please open a [discussion](https://github.com/go-acme/lego/discussions/new?category=q-a).

{{% /notice %}}
 
## Commands

The global flags have been moved to flags of the commands.

Example:

```bash
# Before
lego --dns foo -d '*.example.com' -d example.com run
```

```bash
# After
lego run --dns foo -d '*.example.com' -d example.com
```

The command `renew` has been removed because the command `run` is able to renew certificates.

The command `list` has been removed and replaced by `accounts list` and `certificates list`.

The command `revoke` has been removed and replaced by `certificates revoke`.

## Flags

Some flags have been changed, renamed or removed:

| v4                              | Change Type                                                                       | v5                                |
|---------------------------------|-----------------------------------------------------------------------------------|-----------------------------------|
| `--disable-cn`                  | {{% icon icon="arrow-down-up-across-line" color="orange" %}} removed and replaced | `--enable-cn`                     |
| `--dns.disable-cp`              | {{% icon icon="arrow-down-up-across-line" color="orange" %}} removed and replaced | `--dns.propagation.wait`          |
| `--dns.propagation-wait`        | {{% icon icon="right-left" color="green" %}} renamed                              | `--dns.propagation.wait`          |
| `--dns.propagation-disable-ans` | {{% icon icon="right-left" color="green" %}} renamed                              | `--dns.propagation.disable-ans`   |
| `--dns.propagation-rns`         | {{% icon icon="arrow-down-up-across-line" color="orange" %}} removed and replaced | `--dns.propagation.disable-rns`   |
| `--dns-timeout`                 | {{% icon icon="right-left" color="green" %}} renamed                              | `--dns.timeout`                   |
| `--kid`                         | {{% icon icon="right-left" color="green" %}} renamed                              | `--eab-kid`                       |
| `--hmac`                        | {{% icon icon="right-left" color="green" %}} renamed                              | `--eab-hmac`                      |
| `--days`                        | {{% icon icon="right-left" color="green" %}} renamed                              | `--renew-days`[^1]                |
| `--dynamic`                     | {{% icon icon="xmark" color="red" %}} removed                                     | This is the default behavior now. |
| `--run-hook`                    | {{% icon icon="right-left" color="green" %}} renamed                              | `--deploy-hook`                   |
| `--renew-hook`                  | {{% icon icon="right-left" color="green" %}} renamed                              | `--deploy-hook`                   |
| `--tls.port`                    | {{% icon icon="right-left" color="green" %}} renamed                              | `--tls.address`                   |
| `--http.port`                   | {{% icon icon="right-left" color="green" %}} renamed                              | `--http.address`                  |
| `--pfx.pass`                    | {{% icon icon="right-left" color="green" %}} renamed                              | `--pfx.password`                  |

[^1]: By default, the renewal time is dynamically computed (the behavior of the previous `--dynamic` flag). 

## Directory structure

The directory structure has been changed.

{{< tabs groupid="migration-examples" >}}
{{% tab title="v4" %}}


```
.
├── accounts
│   └── <server-name-1>
│       ├── <account-name-1>
│       │   ├── account.json
│       │   └── keys
│       │       └── <account-name-1>.key
│       └── <account-name-2>
│           ├── account.json
│           └── keys
│               └── <account-name-2>.key
└── certificates
    ├── example.com.crt
    ├── example.com.issuer.crt
    ├── example.com.json
    ├── example.com.key
    ├── example.org.crt
    ├── example.org.issuer.crt
    ├── example.org.json
    └── example.org.key
```

{{% /tab %}}
{{% tab title="v5" %}}


```
.
├── accounts
│   └── <server-name-1>
│       ├── <account-name-1>
│       │   ├── account.json
│       │   └── <account-name-1>.key
│       └── <account-name-2>
│           ├── account.json
│           └── <account-name-2>.key
└── certificates
    ├── example.com.crt
    ├── example.com.issuer.crt
    ├── example.com.json
    ├── example.com.key
    ├── example.org.crt
    ├── example.org.issuer.crt
    ├── example.org.json
    └── example.org.key
```

{{% /tab %}}
{{< /tabs >}}

## Environment variables

The following environment variables have been removed without replacement:

- `SELECTEL_BASE_URL`
- `VSCALE_BASE_URL`

The following environment variables related to the hook have been renamed:

| v4                   | v5                        |
|----------------------|---------------------------|
| `LEGO_ACCOUNT_EMAIL` | `LEGO_HOOK_ACCOUNT_EMAIL` |
| `LEGO_CERT_DOMAIN`   | `LEGO_HOOK_CERT_NAME`     |
| `LEGO_CERT_PATH`     | `LEGO_HOOK_CERT_PATH`     |
| `LEGO_CERT_KEY_PATH` | `LEGO_HOOK_CERT_KEY_PATH` |
| `LEGO_CERT_PEM_PATH` | `LEGO_HOOK_CERT_PEM_PATH` |
| `LEGO_CERT_PFX_PATH` | `LEGO_HOOK_CERT_PFX_PATH` |

## CommonName

The support of the common name is disabled by default.

## PEM encoding

Lego uses `PKCS#8` instead of `PKCS#1` for PEM encoding.
