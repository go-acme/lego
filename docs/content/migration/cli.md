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
lego --dns foo -d '*.example.com' -d example.com -s https://acme-staging-v02.api.letsencrypt.org/directory run
```

```bash
# After
lego run --dns foo -d '*.example.com' -d example.com -s https://acme-staging-v02.api.letsencrypt.org/directory
```

The command `renew` has been removed because the command `run` is able to renew certificates.

The command `list` has been removed and replaced by `accounts list` and `certificates list`.

The command `revoke` has been removed and replaced by `certificates revoke`.

## Flags

Some flags have been changed, renamed or removed:

- `--disable-cn` is removed and replaced by `--enable-cn`.
- `--dns.disable-cp` is removed and replaced by  `--dns.propagation.wait`.
- `--dns.propagation-wait` is renamed to `--dns.propagation.wait`.
- `--dns.propagation-disable-ans` is renamed to `--dns.propagation.disable-ans`.
- `--dns.propagation-rns` is removed and replaced by `--dns.propagation.disable-rns`.
- `--dns-timeout` is renamed to `--dns.timeout`.
- `--kid` is renamed to `--eab-kid`.
- `--hmac` is renamed to `--eab-hmac`.
- `--days` is renamed to `--renew-days`. By default, the renewal time is dynamically computed (the behavior of the previous `--dynamic` flag).
- `--dynamic` is removed (because this is the default behavior now).
- `--run-hook` and `--renew-hook` are replaced by `--deploy-hook`.
- `--tls.port` and `--http.port` are renamed to `--tls.address` and `--http.address`.
- `--pfx.pass` is renamed to `pfx.password`.

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
