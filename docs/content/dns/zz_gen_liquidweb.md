---
title: "Liquid Web"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: liquidweb
dnsprovider:
  since:    "v3.1.0"
  code:     "liquidweb"
  url:      "https://liquidweb.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/liquidweb/liquidweb.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Liquid Web](https://liquidweb.com).


<!--more-->

- Code: `liquidweb`
- Since: v3.1.0


Here is an example bash command using the Liquid Web provider:

```bash
LWAPI_USERNAME=someuser \
LWAPI_PASSWORD=somepass \
lego --email you@example.com --dns liquidweb --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `LWAPI_PASSWORD` | Liquid Web API Password |
| `LWAPI_USERNAME` | Liquid Web API Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `LWAPI_HTTP_TIMEOUT` | Maximum waiting time for the DNS records to be created (not verified) |
| `LWAPI_POLLING_INTERVAL` | Time between DNS propagation check |
| `LWAPI_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `LWAPI_TTL` | The TTL of the TXT record used for the DNS challenge |
| `LWAPI_URL` | Liquid Web API endpoint |
| `LWAPI_ZONE` | DNS Zone |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://api.liquidweb.com/docs/)
- [Go client](https://github.com/liquidweb/liquidweb-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/liquidweb/liquidweb.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
