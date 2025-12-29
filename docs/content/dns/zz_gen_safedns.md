---
title: "UKFast SafeDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: safedns
dnsprovider:
  since:    "v4.6.0"
  code:     "safedns"
  url:      "https://www.ukfast.co.uk/dns-hosting.html"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/safedns/safedns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [UKFast SafeDNS](https://www.ukfast.co.uk/dns-hosting.html).


<!--more-->

- Code: `safedns`
- Since: v4.6.0


Here is an example bash command using the UKFast SafeDNS provider:

```bash
SAFEDNS_AUTH_TOKEN=xxxxxx \
lego --dns safedns -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SAFEDNS_AUTH_TOKEN` | Authentication token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SAFEDNS_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `SAFEDNS_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `SAFEDNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `SAFEDNS_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://developers.ukfast.io/documentation/safedns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/safedns/safedns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
