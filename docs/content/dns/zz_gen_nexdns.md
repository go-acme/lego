---
title: "NexDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: nexdns
dnsprovider:
  since:    "v5.4.0"
  code:     "nexdns"
  url:      "https://nexdns.tech/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nexdns/nexdns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [NexDNS](https://nexdns.tech/).


<!--more-->

- Code: `nexdns`
- Since: v5.4.0


Here is an example bash command using the NexDNS provider:

```bash
NEXDNS_API_TOKEN="xxx" \
lego run --dns nexdns -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NEXDNS_API_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NEXDNS_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `NEXDNS_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `NEXDNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `NEXDNS_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

The API token requires the `zones.read` and `records.write` scopes.



## More information

- [API documentation](https://nexdns.tech/docs/api)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nexdns/nexdns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
