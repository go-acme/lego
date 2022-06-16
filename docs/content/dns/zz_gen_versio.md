---
title: "Versio.[nl|eu|uk]"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: versio
dnsprovider:
  since:    "v2.7.0"
  code:     "versio"
  url:      "https://www.versio.nl/domeinnamen"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/versio/versio.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Versio.[nl|eu|uk]](https://www.versio.nl/domeinnamen).


<!--more-->

- Code: `versio`
- Since: v2.7.0


Here is an example bash command using the Versio.[nl|eu|uk] provider:

```bash
VERSIO_USERNAME=<your login> \
VERSIO_PASSWORD=<your password> \
lego --email you@example.com --dns versio --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `VERSIO_PASSWORD` | Basic authentication password |
| `VERSIO_USERNAME` | Basic authentication username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `VERSIO_ENDPOINT` | The endpoint URL of the API Server |
| `VERSIO_HTTP_TIMEOUT` | API request timeout |
| `VERSIO_POLLING_INTERVAL` | Time between DNS propagation check |
| `VERSIO_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `VERSIO_SEQUENCE_INTERVAL` | Time between sequential requests, default 60s |
| `VERSIO_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).

To test with the sandbox environment set ```VERSIO_ENDPOINT=https://www.versio.nl/testapi/v1/```



## More information

- [API documentation](https://www.versio.nl/RESTapidoc/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/versio/versio.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
