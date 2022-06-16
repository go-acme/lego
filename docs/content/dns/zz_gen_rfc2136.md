---
title: "RFC2136"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: rfc2136
dnsprovider:
  since:    "v0.3.0"
  code:     "rfc2136"
  url:      "https://www.rfc-editor.org/rfc/rfc2136.html"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rfc2136/rfc2136.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [RFC2136](https://www.rfc-editor.org/rfc/rfc2136.html).


<!--more-->

- Code: `rfc2136`
- Since: v0.3.0


Here is an example bash command using the RFC2136 provider:

```bash
RFC2136_NAMESERVER=127.0.0.1 \
RFC2136_TSIG_KEY=lego \
RFC2136_TSIG_ALGORITHM=hmac-sha256. \
RFC2136_TSIG_SECRET=YWJjZGVmZGdoaWprbG1ub3BxcnN0dXZ3eHl6MTIzNDU= \
lego --email you@example.com --dns rfc2136 --domains my.example.org run

## ---

keyname=lego; keyfile=lego.key; tsig-keygen $keyname > $keyfile

RFC2136_NAMESERVER=127.0.0.1 \
RFC2136_TSIG_KEY="$keyname" \
RFC2136_TSIG_ALGORITHM="$( awk -F'[ ";]' '/algorithm/ { print $2 }' $keyfile )." \
RFC2136_TSIG_SECRET="$( awk -F'[ ";]' '/secret/ { print $3 }' $keyfile )" \
lego --email you@example.com --dns rfc2136 --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `RFC2136_NAMESERVER` | Network address in the form "host" or "host:port" |
| `RFC2136_TSIG_ALGORITHM` | TSIG algorithm. See [miekg/dns#tsig.go](https://github.com/miekg/dns/blob/master/tsig.go) for supported values. To disable TSIG authentication, leave the `RFC2136_TSIG*` variables unset. |
| `RFC2136_TSIG_KEY` | Name of the secret key as defined in DNS server configuration. To disable TSIG authentication, leave the `RFC2136_TSIG*` variables unset. |
| `RFC2136_TSIG_SECRET` | Secret key payload. To disable TSIG authentication, leave the` RFC2136_TSIG*` variables unset. |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `RFC2136_DNS_TIMEOUT` | API request timeout |
| `RFC2136_POLLING_INTERVAL` | Time between DNS propagation check |
| `RFC2136_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `RFC2136_SEQUENCE_INTERVAL` | Time between sequential requests |
| `RFC2136_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://www.rfc-editor.org/rfc/rfc2136.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rfc2136/rfc2136.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
