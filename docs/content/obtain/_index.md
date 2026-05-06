---
title: "Obtain or renew certificates"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 2
aliases:
  - usage/cli
  - usage/cli/general-instructions
  - usage/cli/obtain-a-certificate
  - usage/cli/renew-a-certificate
---

This guide explains various ways to get and renew a certificate.

<!--more-->


These examples assume you have [lego installed]({{% ref "install" %}}).
You can get a pre-built binary from the [releases](https://github.com/go-acme/lego/releases) page.

## Quickstart


{{< tabs groupid="usage-examples" >}}
{{% tab title="Classic Way" %}}

Execute the following command:

```bash
lego run -d 'example.com' --http
```

{{% /tab %}}
{{% tab title="With a Configuration File" %}}

Create a `.lego.yml` file with the following content:

```yaml
certificates:
  foo:
    challenge: http-01
    domains:
      - example.com
```

And execute:

```bash
lego
```

{{% /tab %}}
{{< /tabs >}}

## Wildcard Certificates

{{< tabs groupid="usage-examples" >}}
{{% tab title="Classic Way" %}}

Execute the following command:

```bash
CLOUDFLARE_EMAIL="you@example.com" \
CLOUDFLARE_API_KEY="yourprivatecloudflareapikey" \
lego run --dns cloudflare -d 'example.org' -d '*.example.org'
```

{{% /tab %}}
{{% tab title="With a Configuration File" %}}

Create a `.lego.yml` file with the following content:

```yaml
challenges:
  cf:
    dns:
      provider: cloudflare

certificates:
  foo:
    domains:
      - example.com
      - '*.example.com'
```

And execute:

```bash
CLOUDFLARE_EMAIL="you@example.com" \
CLOUDFLARE_API_KEY="yourprivatecloudflareapikey" \
lego
```

{{% /tab %}}
{{< /tabs >}}

## Certificates

You will find your certificates in the `.lego` folder of the current working directory:

```console
$ ls -1 ./.lego/certificates
example.com.crt
example.com.issuer.crt
example.com.json
example.com.key
[maybe more files for different domains...]
```

where

- `example.com.crt` is the server certificate (including the CA certificate),
- `example.com.key` is the private key needed for the server certificate,
- `example.com.issuer.crt` is the CA certificate, and
- `example.com.json` contains some JSON encoded meta information.

For each domain, you will have a set of these four files.
For wildcard certificates (`*.example.com`), the filenames will look like `_.example.com.crt`.

The `.crt` and `.key` files are PEM-encoded x509 certificates and private keys.
If you're looking for a `cert.pem` and `privkey.pem`, you can just use `example.com.crt` and `example.com.key`.

## Using a custom certificate signing request (CSR)

The first step in the process of obtaining certificates involves creating a signing request.
This CSR bundles various information, including the domain name(s) and a public key.
By default, lego will hide this step from you, but if you already have a CSR, you can easily reuse it:

{{< tabs groupid="usage-examples" >}}
{{% tab title="Classic Way" %}}

Execute the following command:

```bash
lego run --http --csr="/path/to/csr.pem"
```

{{% /tab %}}
{{% tab title="With a Configuration File" %}}

Create a `.lego.yml` file with the following content:

```yaml
certificates:
  foo:
    csr: /path/to/csr.pem
    challenge: http-01
    domains:
      - example.com
```

And execute:

```bash
lego
```

{{% /tab %}}
{{< /tabs >}}

lego will infer the domains to be validated based on the contents of the CSR, so make sure the CSR's Common Name and SubjectAltNames are set correctly.

## Challenge Types

{{% children type="card" description="true" %}}
