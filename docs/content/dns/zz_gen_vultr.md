---
title: "Vultr"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: vultr
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/vultr/vultr.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.3.1

Configuration for [Vultr](https://www.vultr.com/).


<!--more-->

- Code: `vultr`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `VULTR_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `VULTR_HTTP_TIMEOUT` | API request timeout |
| `VULTR_POLLING_INTERVAL` | Time between DNS propagation check |
| `VULTR_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `VULTR_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://www.vultr.com/api/#dns)
- [Go client](https://github.com/JamesClonk/vultr)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/vultr/vultr.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
