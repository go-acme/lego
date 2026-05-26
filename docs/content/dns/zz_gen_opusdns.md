---
title: "OpusDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: opusdns
dnsprovider:
  since:    "v5.2.0"
  code:     "opusdns"
  url:      "https://www.opusdns.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/opusdns/opusdns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [OpusDNS](https://www.opusdns.com).


<!--more-->

- Code: `opusdns`
- Since: v5.2.0


Here is an example bash command using the OpusDNS provider:

```bash
OPUSDNS_API_KEY=opk_xxxxxxxxxxxxxxxxxxxxxxxx \
lego run --dns opusdns -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `OPUSDNS_API_KEY` | API key (format: opk_...) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `OPUSDNS_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `OPUSDNS_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 4) |
| `OPUSDNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `OPUSDNS_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://developers.opusdns.com/products/dns/manage-records)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/opusdns/opusdns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
