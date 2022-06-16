---
title: "Njalla"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: njalla
dnsprovider:
  since:    "v4.3.0"
  code:     "njalla"
  url:      "https://njal.la"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/njalla/njalla.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Njalla](https://njal.la).


<!--more-->

- Code: `njalla`
- Since: v4.3.0


Here is an example bash command using the Njalla provider:

```bash
NJALLA_TOKEN=xxxxxxxxxxxxxxxxxxxxxxxxxx \
lego --email you@example.com --dns njalla --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NJALLA_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NJALLA_HTTP_TIMEOUT` | API request timeout |
| `NJALLA_POLLING_INTERVAL` | Time between DNS propagation check |
| `NJALLA_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `NJALLA_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://njal.la/api/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/njalla/njalla.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
