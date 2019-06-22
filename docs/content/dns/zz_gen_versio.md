---
title: "Versio.[nl|eu|uk]"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: versio
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/versio/versio.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v2.7.0

Configuration for [Versio.[nl|eu|uk]](https://www.versio.nl/domeinnamen).


<!--more-->

- Code: `versio`

Here is an example bash command using the Versio.[nl|eu|uk] provider:

```bash
VERSIO_USERNAME=<your login> \
VERSIO_PASSWORD=<your password> \
lego --dns versio --domains my.domain.com --email my@email.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `VERSIO_PASSWORD` | Basic authentication password |
| `VERSIO_USERNAME` | Basic authentication username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `VERSIO_ENDPOINT` | The endpoint URL of the API Server |
| `VERSIO_HTTP_TIMEOUT` | API request timeout |
| `VERSIO_POLLING_INTERVAL` | Time between DNS propagation check |
| `VERSIO_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `VERSIO_SEQUENCE_INTERVAL` | Interval between iteration, default 60s |
| `VERSIO_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).

To test with the sandbox environment set ```VERSIO_ENDPOINT=https://www.versio.nl/testapi/v1/```



## More information

- [API documentation](https://www.versio.nl/RESTapidoc/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/versio/versio.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
