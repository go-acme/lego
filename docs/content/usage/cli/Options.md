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

By default, lego assumes it is able to bind to ports 80 and 443 to solve challenges.
If this is not possible in your environment, you can use the `--http.port` and `--tls.port` options to instruct
lego to listen on that interface:port for any incoming challenges.

If you are using either of these options, make sure you setup a proxy to redirect traffic to the chosen ports.

**HTTP Port:** All plaintext HTTP requests to port **80** which begin with a request path of `/.well-known/acme-challenge/` for the HTTP challenge[^header].

**TLS Port:** All TLS handshakes on port **443** for the TLS-ALPN challenge.

This traffic redirection is only needed as long as lego solves challenges. As soon as you have received your certificates you can deactivate the forwarding.

[^header]: You must ensure that incoming validation requests contains the correct value for the HTTP `Host` header. If you operate lego behind a non-transparent reverse proxy (such as Apache or NGINX), you might need to alter the header field using `--http.proxy-header X-Forwarded-Host`.

## DNS Resolvers and Challenge Verification

When using a DNS challenge provider (via `--dns <name>`), Lego tries to ensure the ACME challenge token is properly setup before instructing the ACME provider to perform the validation.

This involves a few DNS queries to different servers:

1. Determining the DNS zone and resolving CNAMEs.

   The DNS zone for a given domain is determined by the SOA record, which contains the authoritative name server for the domain and all its subdomains.
   For simple domains like `example.com`, this is usually `example.com` itself.
   For other domains (like `fra.eu.cdn.example.com`), this can get complicated, as `cdn.example.com` may be delegated to the CDN provider, which means for `cdn.example.com` must exist a different SOA record.

   To find the correct zone, Lego requests the SOA record for each DNS label (starting on the leaf domain, i.e. the left-most DNS label).
   If there is no SOA record, Lego requests the SOA record of the parent label, then for its parent, etc., until it reaches the apex domain[^apex].
   Should any DNS label on the way be a CNAME, it is resolved as per usual.

   In the default configuration, Lego uses the system name servers for this, and falls back to Google's DNS servers, should they be absent.

2. Verifying the challenge token.

   The `_acme-challenge.<yourdomain>` TXT record must be correctly installed.
   Lego verifies this by directly querying the authoritative name server for this record (as detected in the previous step).

Strictly speaking, this verification step is not necessary, but helps to protect your ACME account.
Remember that some ACME providers impose a rate limit on certain actions (at the time of writing, Let's Encrypt allows 300 new certificate orders per account per 3 hours).

There are also situations, where this verification step doesn't work as expected:

- A "split DNS" setup gives different answers to clients on the internal network (Lego) vs. on the public internet (Let's Encrypt).
- With "hidden master" setups, Lego may be able to directly talk to the primary DNS server, while the `_acme-challenge` record might not have fully propagated to the (public) secondary servers, yet.

The effect is the same: Lego determined the challenge token to be installed correctly, while Let's Encrypt has a different view, and rejects the certificate order.

In these cases, you can instruct Lego to use a different DNS resolver, using the `--dns.resolvers` flag.
You should prefer one on the public internet, otherwise you might be susceptible to the same problem.

[^apex]: The apex domain is the domain you have registered with your domain registrar. For gTLDs (`.com`, `.fyi`) this is the 2nd level domain, but for ccTLDs, this can either be the 2nd level (`.de`) or 3rd level domain (`.co.uk`).

## Other options

### LEGO_CA_CERTIFICATES

The environment variable `LEGO_CA_CERTIFICATES` allows to specify the path to PEM-encoded CA certificates
that can be used to authenticate an ACME server with an HTTPS certificate not issued by a CA in the system-wide trusted root list.

Multiple file paths can be added by using `:` (unix) or `;` (Windows) as a separator.

Example:

```bash
# On Unix system
LEGO_CA_CERTIFICATES=/foo/cert1.pem:/foo/cert2.pem
```

### LEGO_CA_SYSTEM_CERT_POOL

The environment variable `LEGO_CA_SYSTEM_CERT_POOL` can be used to define if the certificates pool must use a copy of the system cert pool.

Example:

```bash
LEGO_CA_SYSTEM_CERT_POOL=true
```

### LEGO_CA_SERVER_NAME

The environment variable `LEGO_CA_SERVER_NAME` allows to specify the CA server name used to authenticate an ACME server
with an HTTPS certificate not issued by a CA in the system-wide trusted root list.

Example:

```bash
LEGO_CA_SERVER_NAME=foo
```

### LEGO_DISABLE_CNAME_SUPPORT

By default, lego follows CNAME, the environment variable `LEGO_DISABLE_CNAME_SUPPORT` allows to disable this support.

Example:

```bash
LEGO_DISABLE_CNAME_SUPPORT=false
```

### LEGO_DEBUG_CLIENT_VERBOSE_ERROR

The environment variable `LEGO_DEBUG_CLIENT_VERBOSE_ERROR` allows to enrich error messages from some of the DNS clients.

Example:

```bash
LEGO_DEBUG_CLIENT_VERBOSE_ERROR=true
```

### LEGO_DEBUG_DNS_API_HTTP_CLIENT

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

### LEGO_DEBUG_ACME_HTTP_CLIENT

The environment variable `LEGO_DEBUG_ACME_HTTP_CLIENT` allows debug the calls to the ACME server.

Example:

```bash
LEGO_DEBUG_ACME_HTTP_CLIENT=true
```
