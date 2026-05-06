---
title: "dns.la"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dnsla
dnsprovider:
  since:    "v5.0.0"
  code:     "dnsla"
  url:      "https://www.dns.la"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnsla/dnsla.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [dns.la](https://www.dns.la).


<!--more-->

- Code: `dnsla`
- Since: v5.0.0


Here is an example bash command using the dns.la provider:

```bash
DNSLA_API_SECRET="xxx" \
DNSLA_API_SECRET="yyy" \
lego run --dns dnsla -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DNSLA_API_ID` | API ID |
| `DNSLA_API_SECRET` | API secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DNSLA_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `DNSLA_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `DNSLA_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `DNSLA_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.dns.la/docs/ApiDoc)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnsla/dnsla.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
