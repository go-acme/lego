---
title: "Epik"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: epik
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/epik/epik.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v4.5.0

Configuration for [Epik](https://www.epik.com/).


<!--more-->

- Code: `epik`

Here is an example bash command using the Epik provider:

```bash
EPIK_SIGNATURE=xxxxxxxxxxxxxxxxxxxxxxxxxx \
lego --email myemail@example.com --dns epik --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `EPIK_SIGNATURE` | Epik API signature (https://registrar.epik.com/account/api-settings/) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `EPIK_HTTP_TIMEOUT` | API request timeout |
| `EPIK_POLLING_INTERVAL` | Time between DNS propagation check |
| `EPIK_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `EPIK_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://docs.userapi.epik.com/v2/#/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/epik/epik.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
