---
title: "FastDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: fastdns
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/fastdns/fastdns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.5.0

Configuration for [FastDNS](https://www.akamai.com/us/en/products/security/fast-dns.jsp).


<!--more-->

- Code: `fastdns`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




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
| `AKAMAI_POLLING_INTERVAL` | Time between DNS propagation check |
| `AKAMAI_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `AKAMAI_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://developer.akamai.com/api/web_performance/fast_dns_record_management/v1.html)
- [Go client](https://github.com/akamai/AkamaiOPEN-edgegrid-golang)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/fastdns/fastdns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
