---
title: "DonDominio/MrDomain"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dondominio
dnsprovider:
  since:    "v5.7.0"
  code:     "dondominio"
  url:      "https://dondominio.dev"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dondominio/dondominio.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [DonDominio/MrDomain](https://dondominio.dev).


<!--more-->

- Code: `dondominio`
- Since: v5.7.0


Here is an example bash command using the DonDominio/MrDomain provider:

```bash
DONDOMINIO_API_USER="xxx" \
DONDOMINIO_API_PASSWORD="yyy" \
lego run --dns dondominio -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DONDOMINIO_API_PASSWORD` | API password |
| `DONDOMINIO_API_USER` | API user |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DONDOMINIO_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `DONDOMINIO_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `DONDOMINIO_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `DONDOMINIO_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://dondominio.dev/en/api/docs/api/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dondominio/dondominio.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
