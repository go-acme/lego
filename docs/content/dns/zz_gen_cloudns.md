---
title: "ClouDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: cloudns
dnsprovider:
  since:    "v2.3.0"
  code:     "cloudns"
  url:      "https://www.cloudns.net"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/cloudns/cloudns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [ClouDNS](https://www.cloudns.net).


<!--more-->

- Code: `cloudns`
- Since: v2.3.0


Here is an example bash command using the ClouDNS provider:

```bash
CLOUDNS_AUTH_ID=xxxx \
CLOUDNS_AUTH_PASSWORD=yyyy \
lego --email you@example.com --dns cloudns -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `CLOUDNS_AUTH_ID` | The API user ID |
| `CLOUDNS_AUTH_PASSWORD` | The password for API user ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `CLOUDNS_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `CLOUDNS_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `CLOUDNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 180) |
| `CLOUDNS_SUB_AUTH_ID` | The API sub user ID |
| `CLOUDNS_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.cloudns.net/wiki/article/42/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/cloudns/cloudns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
