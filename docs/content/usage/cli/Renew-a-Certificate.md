---
title: Renew a Certificate
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 3
---

This guide describes how to renew existing certificates.

<!--more-->

Certificates issues by Let's Encrypt are valid for a period of 90 days.
To avoid certificate errors, you need to ensure that you renew your certificate *before* it expires.

In order to renew a certificate, follow the general instructions laid out under [Obtain a Certificate]({{< ref "usage/cli/Obtain-a-Certificate" >}}), and replace `lego ... run` with `lego ... renew`.
Note that the `renew` sub-command supports a slightly different set of some command line flags.

## Using the built-in web server

By default, and following best practices, a certificate is only renewed if its expiry date is less than 30 days in the future.

```bash
lego --email="you@example.com" --domains="example.com" --http renew
```

If the certificate needs to renewed earlier, you can specify the number of remaining days:

```bash
lego --email="you@example.com" --domains="example.com" --http renew --days 45
```

## Using a DNS provider

If you can't or don't want to start a web server, you need to use a DNS provider.
lego comes with [support for many]({{< ref "dns#dns-providers" >}}) providers,
and you need to pick the one where your domain's DNS settings are set up.
Typically, this is the registrar where you bought the domain, but in some cases this can be another third-party provider.

For this example, let's assume you have set up CloudFlare for your domain.

Execute this command:

```bash
CLOUDFLARE_EMAIL="you@example.com" \
CLOUDFLARE_API_KEY="yourprivatecloudflareapikey" \
lego --email "you@example.com" --dns cloudflare --domains "example.org" renew
```

## Running a script afterward

You can easily hook into the certificate-obtaining process by providing the path to a script.
The hook is executed only when the certificates are effectively renewed.

```bash
lego --email="you@example.com" --domains="example.com" --http renew --renew-hook="./myscript.sh"
```

Some information is provided through environment variables:

- `LEGO_ACCOUNT_EMAIL`: the email of the account.
- `LEGO_CERT_DOMAIN`: the main domain of the certificate.
- `LEGO_CERT_PATH`: the path of the certificate.
- `LEGO_CERT_KEY_PATH`: the path of the certificate key.

See [Obtain a Certificate → Use case]({{< ref "usage/cli/Obtain-a-Certificate#use-case" >}}) for an example script.

## Automatic renewal

It is tempting to create a cron job (or systemd timer) to automatically renew all you certificates.

When doing so, please note that some cron defaults will cause measurable load on the ACME provider's infrastructure.
Notably `@daily` jobs run at midnight.

To both counteract load spikes (caused by all lego users) and reduce subsequent renewal failures, we were asked to implement a small random delay for non-interactive renewals.[^loadspikes]
Since v4.8.0, lego will pause for up to 8 minutes to help spread the load.

You can help further, by adjusting your crontab entry, like so:

```ruby
# avoid:
#@daily      /usr/bin/lego ... renew
#@midnight   /usr/bin/lego ... renew
#0 0 * * *   /usr/bin/lego ... renew

# instead, use a randomly chosen time:
35 3 * * *  /usr/bin/lego ... renew
```

If you use systemd timers, consider doing something similar, and/or introduce a `RandomizedDelaySec`:

```ini
[Unit]
Description=Renew certificates

[Timer]
Persistent=true
# avoid:
#OnCalendar=*-*-* 00:00:00
#OnCalendar=daily

# instead, use a randomly chosen time:
OnCalendar=*-*-* 3:35
# add extra delay, here up to 1 hour:
RandomizedDelaySec=1h

[Install]
WantedBy=timers.target
```

[^loadspikes]: See [GitHub issue #1656](https://github.com/go-acme/lego/issues/1656) for an excellent problem description.
