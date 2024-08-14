---
title: Obtain a Certificate
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 2
---

This guide explains various ways to obtain a new certificate.

<!--more-->

## Using the built-in web server

Open a terminal, and execute the following command (insert your own email address and domain):

```bash
lego --email="you@example.com" --domains="example.com" --http run
```

You will find your certificate in the `.lego` folder of the current working directory:

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


## Using a DNS provider

If you can't or don't want to start a web server, you need to use a DNS provider.
lego comes with [support for many]({{% ref "dns#dns-providers" %}}) providers,
and you need to pick the one where your domain's DNS settings are set up.
Typically, this is the registrar where you bought the domain, but in some cases this can be another third-party provider.

For this example, let's assume you have set up CloudFlare for your domain.

Execute this command:

```bash
CLOUDFLARE_EMAIL="you@example.com" \
CLOUDFLARE_API_KEY="yourprivatecloudflareapikey" \
lego --email "you@example.com" --dns cloudflare --domains "example.org" run
```


## Using a custom certificate signing request (CSR)

The first step in the process of obtaining certificates involves creating a signing request.
This CSR bundles various information, including the domain name(s) and a public key.
By default, lego will hide this step from you, but if you already have a CSR, you can easily reuse it:

```bash
lego --email="you@example.com" --http --csr="/path/to/csr.pem" run
```

lego will infer the domains to be validated based on the contents of the CSR, so make sure the CSR's Common Name and optional SubjectAltNames are set correctly.


## Using an existing, running web server

If you have an existing server running on port 80, the `--http` option also requires the `--http.webroot` option.
This just writes the http-01 challenge token to the given directory in the folder `.well-known/acme-challenge` and does not start a server.

The given directory **should** be publicly served as `/` on the domain(s) for the validation to complete.

If the given directory is not publicly served you will have to support rewriting the request to the directory;

You could also implement a rewrite to rewrite `.well-known/acme-challenge` to the given directory `.well-known/acme-challenge`.

You should be able to run an existing webserver on port 80 and have lego write the token file with the HTTP-01 challenge key authorization to `<webroot dir>/.well-known/acme-challenge/` by running something like:

```bash
lego --accept-tos --email you@example.com --http --http.webroot /path/to/webroot --domains example.com run
```

## Running a script afterward

You can easily hook into the certificate-obtaining process by providing the path to a script:

```bash
lego --email="you@example.com" --domains="example.com" --http run --run-hook="./myscript.sh"
```

Some information is provided through environment variables:

- `LEGO_ACCOUNT_EMAIL`: the email of the account.
- `LEGO_CERT_DOMAIN`: the main domain of the certificate.
- `LEGO_CERT_PATH`: the path of the certificate.
- `LEGO_CERT_KEY_PATH`: the path of the certificate key.
- `LEGO_CERT_PEM_PATH`: (only with `--pem`) the path to the PEM certificate.
- `LEGO_CERT_PFX_PATH`: (only with `--pfx`) the path to the PFX certificate.

### Use case

A typical use case is distribute the certificate for other services and reload them if necessary.
Since PEM-formatted TLS certificates are understood by many programs, it is relatively simple to use certificates for more than a web server.

This example script installs the new certificate for a mail server, and reloads it.
Beware: this is just a starting point, error checking is omitted for brevity.

```bash
#!/bin/bash

# copy certificates to a directory controlled by Postfix
postfix_cert_dir="/etc/postfix/certificates"

# our Postfix server only handles mail for @example.com domain
if [ "$LEGO_CERT_DOMAIN" = "example.com" ]; then
  install -u postfix -g postfix -m 0644 "$LEGO_CERT_PATH" "$postfix_cert_dir"
  install -u postfix -g postfix -m 0640 "$LEGO_CERT_KEY_PATH"  "$postfix_cert_dir"

  systemctl reload postfix@-service
fi
```
