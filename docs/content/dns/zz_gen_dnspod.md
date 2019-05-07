---
title: "DNSPod"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dnspod
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnspod/dnspod.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.4.0

Configuration for [DNSPod](http://www.dnspod.com/).


<!--more-->

- Code: `dnspod`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DNSPOD_API_KEY` | The user token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DNSPOD_HTTP_TIMEOUT` | API request timeout |
| `DNSPOD_POLLING_INTERVAL` | Time between DNS propagation check |
| `DNSPOD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DNSPOD_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://www.dnspod.com/docs/index.html)
- [Go client](https://github.com/decker502/dnspod-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnspod/dnspod.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
