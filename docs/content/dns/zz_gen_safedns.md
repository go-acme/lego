---
title: "SafeDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: safedns
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/safedns/safedns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v4.6.0

Configuration for [SafeDNS](https://www.ukfast.co.uk/dns-hosting.html).


<!--more-->

- Code: `safedns`

Here is an example bash command using the SafeDNS provider:

```bash
SAFEDNS_AUTH_TOKEN=xxxxxx \
lego --email myemail@example.com --dns safedns --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SAFEDNS_AUTH_TOKEN` | Authentication token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SAFEDNS_API_TIMEOUT` | API request timeout in seconds |
| `SAFEDNS_POLLING_INTERVAL` | Time to wait for initial check |
| `SAFEDNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `SAFEDNS_TTL` | TXT record TTL |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://developers.ukfast.io/documentation/safedns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/safedns/safedns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
