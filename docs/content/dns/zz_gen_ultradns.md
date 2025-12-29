---
title: "Ultradns"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: ultradns
dnsprovider:
  since:    "v4.10.0"
  code:     "ultradns"
  url:      "https://vercara.com/authoritative-dns"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ultradns/ultradns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Ultradns](https://vercara.com/authoritative-dns).


<!--more-->

- Code: `ultradns`
- Since: v4.10.0


Here is an example bash command using the Ultradns provider:

```bash
ULTRADNS_USERNAME=username \
ULTRADNS_PASSWORD=password \
lego --dns ultradns -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ULTRADNS_PASSWORD` | API Password |
| `ULTRADNS_USERNAME` | API Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ULTRADNS_ENDPOINT` | API endpoint URL, defaults to https://api.ultradns.com/ |
| `ULTRADNS_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 4) |
| `ULTRADNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `ULTRADNS_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://ultra-portalstatic.ultradns.com/static/docs/REST-API_User_Guide.pdf)
- [Go client](https://github.com/ultradns/ultradns-go-sdk)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ultradns/ultradns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
