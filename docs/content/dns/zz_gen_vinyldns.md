---
title: "VinylDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: vinyldns
dnsprovider:
  since:    "v4.4.0"
  code:     "vinyldns"
  url:      "https://www.vinyldns.io"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/vinyldns/vinyldns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [VinylDNS](https://www.vinyldns.io).


<!--more-->

- Code: `vinyldns`
- Since: v4.4.0


Here is an example bash command using the VinylDNS provider:

```bash
VINYLDNS_ACCESS_KEY=xxxxxx \
VINYLDNS_SECRET_KEY=yyyyy \
VINYLDNS_HOST=https://api.vinyldns.example.org:9443 \
lego --email you@example.com --dns vinyldns --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `VINYLDNS_ACCESS_KEY` | The VinylDNS API key |
| `VINYLDNS_HOST` | The VinylDNS API URL |
| `VINYLDNS_SECRET_KEY` | The VinylDNS API Secret key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `VINYLDNS_POLLING_INTERVAL` | Time between DNS propagation check |
| `VINYLDNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `VINYLDNS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

The vinyldns integration makes use of dotted hostnames to ease permission management.
Users are required to have DELETE ACL level or zone admin permissions on the VinylDNS zone containing the target host.



## More information

- [API documentation](https://www.vinyldns.io/api/)
- [Go client](https://github.com/vinyldns/go-vinyldns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/vinyldns/vinyldns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
