---
title: "DNS.services"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dnsservices
dnsprovider:
  since:    "v5.0.0"
  code:     "dnsservices"
  url:      "https://dns.services/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnsservices/dnsservices.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [DNS.services](https://dns.services/).


<!--more-->

- Code: `dnsservices`
- Since: v5.0.0


Here is an example bash command using the DNS.services provider:

```bash
DNSSERVICES_USERNAME="xxxxxxxxxxxxxxxxxxxxx" \
DNSSERVICES_PASSWORD="xxxxxxxxxxxxxxxxxxxxx" \
lego --dns dnsservices -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DNSSERVICES_PASSWORD` | Password |
| `DNSSERVICES_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DNSSERVICES_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `DNSSERVICES_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `DNSSERVICES_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `DNSSERVICES_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://dns.services/userapi#dns-85)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnsservices/dnsservices.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
