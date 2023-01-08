---
title: "Options"
date: 2019-03-03T16:39:46+01:00
draft: false
summary: This page describes various command line options.
weight: 4
---

## Usage

{{< clihelp >}}

When using the standard `--path` option, all certificates and account configurations are saved to a folder `.lego` in the current working directory.


## Let's Encrypt ACME server

lego defaults to communicating with the production Let's Encrypt ACME server.
If you'd like to test something without issuing real certificates, consider using the staging endpoint instead:

```bash
lego --server=https://acme-staging-v02.api.letsencrypt.org/directory …
```

## Running without root privileges

The CLI does not require root permissions but needs to bind to port 80 and 443 for certain challenges.
To run the CLI without `sudo`, you have four options:

- Use `setcap 'cap_net_bind_service=+ep' /path/to/lego` (Linux only)
- Pass the `--http.port` or/and the `--tls.port` option and specify a custom port to bind to. In this case you have to forward port 80/443 to these custom ports (see [Port Usage](#port-usage)).
- Pass the `--http.webroot` option and specify the path to your webroot folder. In this case the challenge will be written in a file in `.well-known/acme-challenge/` inside your webroot.
- Pass the `--dns` option and specify a DNS provider.

## Port Usage

By default lego assumes it is able to bind to ports 80 and 443 to solve challenges.
If this is not possible in your environment, you can use the `--http.port` and `--tls.port` options to instruct
lego to listen on that interface:port for any incoming challenges.

If you are using this option, make sure you proxy all of the following traffic to these ports.

**HTTP Port:** All plaintext HTTP requests to port **80** which begin with a request path of `/.well-known/acme-challenge/` for the HTTP challenge.[^header]

**TLS Port:** All TLS handshakes on port **443** for the TLS-ALPN challenge.

This traffic redirection is only needed as long as lego solves challenges. As soon as you have received your certificates you can deactivate the forwarding.

[^header]: You must ensure that incoming validation requests contains the correct value for the HTTP `Host` header. If you operate lego behind a non-transparent reverse proxy (such as Apache or NGINX), you might need to alter the header field using `--http.proxy-header X-Forwarded-Host`.
