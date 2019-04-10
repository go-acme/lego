---
title: "Namecheap"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: namecheap
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/namecheap/namecheap.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.3.0

Configuration for [Namecheap](https://www.namecheap.com).


<!--more-->

- Code: `namecheap`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NAMECHEAP_API_KEY` | API key |
| `NAMECHEAP_API_USER` | API user |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NAMECHEAP_HTTP_TIMEOUT` | API request timeout |
| `NAMECHEAP_POLLING_INTERVAL` | Time between DNS propagation check |
| `NAMECHEAP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `NAMECHEAP_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://www.namecheap.com/support/api/methods.aspx)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/namecheap/namecheap.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
