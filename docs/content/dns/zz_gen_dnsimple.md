---
title: "DNSimple"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dnsimple
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnsimple/dnsimple.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.3.0

Configuration for [DNSimple](https://dnsimple.com/).


<!--more-->

- Code: `dnsimple`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DNSIMPLE_BASE_URL` | API endpoint URL |
| `DNSIMPLE_OAUTH_TOKEN` | OAuth token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DNSIMPLE_POLLING_INTERVAL` | Time between DNS propagation check |
| `DNSIMPLE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DNSIMPLE_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://developer.dnsimple.com/v2/)
- [Go client](https://github.com/dnsimple/dnsimple-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnsimple/dnsimple.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
