---
title: "Akamai EdgeDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: edgedns
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/edgedns/edgedns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v3.9.0

Akamai edgedns supercedes FastDNS; implementing a DNS provider for solving the DNS-01 challenge using Akamai EdgeDNS



<!--more-->

- Code: `edgedns`

Here is an example bash command using the Akamai EdgeDNS provider:

```bash
AKAMAI_CLIENT_SECRET=abcdefghijklmnopqrstuvwxyz1234567890ABCDEFG= \
AKAMAI_CLIENT_TOKEN=akab-mnbvcxzlkjhgfdsapoiuytrewq1234567 \
AKAMAI_HOST=akab-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.luna.akamaiapis.net \
AKAMAI_ACCESS_TOKEN=akab-1234567890qwerty-asdfghjklzxcvtnu \
lego --domains="example.zone" --email="testuser@mail.me" --dns="edgedns" -a run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `AKAMAI_ACCESS_TOKEN` | Access token |
| `AKAMAI_CLIENT_SECRET` | Client secret |
| `AKAMAI_CLIENT_TOKEN` | Client token |
| `AKAMAI_HOST` | API host |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `AKAMAI_POLLING_INTERVAL` | Time between DNS propagation check. Default: 15 seconds |
| `AKAMAI_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation. Default: 3 minutes |
| `AKAMAI_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://developer.akamai.com/api/cloud_security/edge_dns_zone_management/v2.html)
- [Go client](https://github.com/akamai/AkamaiOPEN-edgegrid-golang)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/edgedns/edgedns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
