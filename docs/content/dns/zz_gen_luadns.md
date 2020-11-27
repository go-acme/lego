---
title: "LuaDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: luadns
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/luadns/luadns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v3.7.0

Configuration for [LuaDNS](https://luadns.com).


<!--more-->

- Code: `luadns`

Here is an example bash command using the LuaDNS provider:

```bash
LUADNS_API_USERNAME=youremail \
LUADNS_API_TOKEN=xxxxxxxx \
lego --email myemail@example.com --dns luadns --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `LUADNS_API_TOKEN` | API token |
| `LUADNS_API_USERNAME` | Username (your email) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `LUADNS_HTTP_TIMEOUT` | API request timeout |
| `LUADNS_POLLING_INTERVAL` | Time between DNS propagation check |
| `LUADNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `LUADNS_SEQUENCE_INTERVAL` | Interval between iteration |
| `LUADNS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://luadns.com/api.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/luadns/luadns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
