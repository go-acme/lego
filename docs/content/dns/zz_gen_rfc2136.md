---
title: "RFC2136"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: rfc2136
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rfc2136/rfc2136.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.3.0

Configuration for [RFC2136](https://tools.ietf.org/html/rfc2136).


<!--more-->

- Code: `rfc2136`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `RFC2136_NAMESERVER` | Network address in the form "host" or "host:port" |
| `RFC2136_TSIG_ALGORITHM` | TSIG algorythm. See [miekg/dns#tsig.go](https://github.com/miekg/dns/blob/master/tsig.go) for supported values. To disable TSIG authentication, leave the `RFC2136_TSIG*` variables unset. |
| `RFC2136_TSIG_KEY` | Name of the secret key as defined in DNS server configuration. To disable TSIG authentication, leave the `RFC2136_TSIG*` variables unset. |
| `RFC2136_TSIG_SECRET` | Secret key payload. To disable TSIG authentication, leave the` RFC2136_TSIG*` variables unset. |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `RFC2136_DNS_TIMEOUT` | API request timeout |
| `RFC2136_POLLING_INTERVAL` | Time between DNS propagation check |
| `RFC2136_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `RFC2136_SEQUENCE_INTERVAL` | Interval between iteration |
| `RFC2136_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://tools.ietf.org/html/rfc2136)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rfc2136/rfc2136.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
