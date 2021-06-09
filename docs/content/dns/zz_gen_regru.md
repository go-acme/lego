---
title: "reg.ru"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: regru
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/regru/regru.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v3.5.0

Configuration for [reg.ru](https://www.reg.ru/).


<!--more-->

- Code: `regru`

Here is an example bash command using the reg.ru provider:

```bash
REGRU_USERNAME=xxxxxx \
REGRU_PASSWORD=yyyyyy \
lego --email myemail@example.com --dns regru --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `REGRU_PASSWORD` | API password |
| `REGRU_USERNAME` | API username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `REGRU_HTTP_TIMEOUT` | API request timeout |
| `REGRU_POLLING_INTERVAL` | Time between DNS propagation check |
| `REGRU_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `REGRU_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://www.reg.ru/support/help/api2)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/regru/regru.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
