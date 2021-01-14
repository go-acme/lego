---
title: "Ionos"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: ionos
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ionos/ionos.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v4.2.0

Configuration for [Ionos](https://ionos.com).


<!--more-->

- Code: `ionos`

Here is an example bash command using the Ionos provider:

```bash
IONOS_API_KEY=xxxxxxxx \
lego --email myemail@example.com --dns ionos --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `IONOS_API_KEY` | API key `<prefix>.<secret>` https://developer.hosting.ionos.com/docs/getstarted |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `IONOS_HTTP_TIMEOUT` | API request timeout |
| `IONOS_POLLING_INTERVAL` | Time between DNS propagation check |
| `IONOS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `IONOS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://developer.hosting.ionos.com/docs/dns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ionos/ionos.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
