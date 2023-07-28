---
title: "Manual"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: manual
dnsprovider:
  since: v0.3.0
  code: manual
  url:
---

Solving the DNS-01 challenge using CLI prompt.

<!--more-->

## Example

To start using the CLI prompt "provider", start lego with `--dns manual`:

```console
$ lego --email "you@example.com" --domains="example.com" --dns "manual" run
```

What follows are a few log print-outs, interspersed with some prompts, asking for you to do perform some actions:

```txt
No key found for account you@example.com. Generating a P256 key.
Saved key to ./.lego/accounts/acme-v02.api.letsencrypt.org/you@example.com/keys/you@example.com.key
Please review the TOS at https://letsencrypt.org/documents/LE-SA-v1.2-November-15-2017.pdf
Do you accept the TOS? Y/n
```

If you accept the linked Terms of Service, hit `Enter`.

```txt
[INFO] acme: Registering account for you@example.com
!!!! HEADS UP !!!!

    Your account credentials have been saved in your Let's Encrypt
    configuration directory at "./.lego/accounts".

    You should make a secure backup of this folder now. This
    configuration directory will also contain certificates and
    private keys obtained from Let's Encrypt so making regular
    backups of this folder is ideal.
[INFO] [example.com] acme: Obtaining bundled SAN certificate
[INFO] [example.com] AuthURL: https://acme-v02.api.letsencrypt.org/acme/authz-v3/2345678901
[INFO] [example.com] acme: Could not find solver for: tls-alpn-01
[INFO] [example.com] acme: Could not find solver for: http-01
[INFO] [example.com] acme: use dns-01 solver
[INFO] [example.com] acme: Preparing to solve DNS-01
lego: Please create the following TXT record in your example.com. zone:
_acme-challenge.example.com. 120 IN TXT "hX0dPkG6Gfs9hUvBAchQclkyyoEKbShbpvJ9mY5q2JQ"
lego: Press 'Enter' when you are done
```

Do as instructed, and create the TXT records, and hit `Enter`.

```txt
[INFO] [example.com] acme: Trying to solve DNS-01
[INFO] [example.com] acme: Checking DNS record propagation using [192.168.8.1:53]
[INFO] Wait for propagation [timeout: 1m0s, interval: 2s]
[INFO] [example.com] acme: Waiting for DNS record propagation.
[INFO] [example.com] The server validated our request
[INFO] [example.com] acme: Cleaning DNS-01 challenge
lego: You can now remove this TXT record from your example.com. zone:
_acme-challenge.example.com. 120 IN TXT "hX0dPkG6Gfs9hUvBAchQclkyyoEKbShbpvJ9mY5q2JQ"
[INFO] [example.com] acme: Validations succeeded; requesting certificates
[INFO] [example.com] Server responded with a certificate.
```

As mentioned, you can now remove the TXT record again.
