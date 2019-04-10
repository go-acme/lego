---
title: "Selectel"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: selectel
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/selectel/selectel.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v1.2.0

Configuration for [Selectel](https://kb.selectel.com/).


<!--more-->

- Code: `selectel`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SELECTEL_API_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SELECTEL_BASE_URL` | API endpoint URL |
| `SELECTEL_HTTP_TIMEOUT` | API request timeout |
| `SELECTEL_POLLING_INTERVAL` | Time between DNS propagation check |
| `SELECTEL_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `SELECTEL_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://kb.selectel.com/23136054.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/selectel/selectel.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
