---
title: "Efficient IP"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: efficientip
dnsprovider:
  since:    "v4.13.0"
  code:     "efficientip"
  url:      "https://efficientip.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/efficientip/efficientip.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Efficient IP](https://efficientip.com/).


<!--more-->

- Code: `efficientip`
- Since: v4.13.0


Here is an example bash command using the Efficient IP provider:

```bash
EFFICIENTIP_USERNAME="user" \
EFFICIENTIP_PASSWORD="secret" \
EFFICIENTIP_HOSTNAME="ipam.example.org" \
EFFICIENTIP_DNS_NAME="dns.smart" \
lego --email you@example.com --dns efficientip --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `EFFICIENTIP_DNS_NAME` | DNS name (ex: dns.smart) |
| `EFFICIENTIP_HOSTNAME` | Hostname (ex: foo.example.com) |
| `EFFICIENTIP_PASSWORD` | Password |
| `EFFICIENTIP_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `EFFICIENTIP_HTTP_TIMEOUT` | API request timeout |
| `EFFICIENTIP_INSECURE_SKIP_VERIFY` | Whether or not to verify EfficientIP API certificate |
| `EFFICIENTIP_POLLING_INTERVAL` | Time between DNS propagation check |
| `EFFICIENTIP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `EFFICIENTIP_TTL` | The TTL of the TXT record used for the DNS challenge |
| `EFFICIENTIP_VIEW_NAME` | View name (ex: external) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).





<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/efficientip/efficientip.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
