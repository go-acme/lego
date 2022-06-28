---
title: "WEDOS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: wedos
dnsprovider:
  since:    "v4.4.0"
  code:     "wedos"
  url:      "https://www.wedos.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/wedos/wedos.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [WEDOS](https://www.wedos.com).


<!--more-->

- Code: `wedos`
- Since: v4.4.0


Here is an example bash command using the WEDOS provider:

```bash
WEDOS_USERNAME=xxxxxxxx \
WEDOS_WAPI_PASSWORD=xxxxxxxx \
lego --email you@example.com --dns wedos --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `WEDOS_USERNAME` | Username is the same as for the admin account |
| `WEDOS_WAPI_PASSWORD` | Password needs to be generated and IP allowed in the admin interface |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `WEDOS_HTTP_TIMEOUT` | API request timeout |
| `WEDOS_POLLING_INTERVAL` | Time between DNS propagation check |
| `WEDOS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `WEDOS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://kb.wedos.com/en/kategorie/wapi-api-interface/wdns-en/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/wedos/wedos.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
