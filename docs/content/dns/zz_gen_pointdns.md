---
title: "PointDNS/PointHQ"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: pointdns
dnsprovider:
  since:    "v5.0.0"
  code:     "pointdns"
  url:      "https://pointhq.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/pointdns/pointdns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [PointDNS/PointHQ](https://pointhq.com/).


<!--more-->

- Code: `pointdns`
- Since: v5.0.0


Here is an example bash command using the PointDNS/PointHQ provider:

```bash
POINTDNS_USERNAME="xxx" \
POINTDNS_PASSWORD="yyy" \
lego run --dns pointdns -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `POINTDNS_PASSWORD` | Password |
| `POINTDNS_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `POINTDNS_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `POINTDNS_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `POINTDNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `POINTDNS_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://w.pointhq.com/api/docs)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/pointdns/pointdns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
