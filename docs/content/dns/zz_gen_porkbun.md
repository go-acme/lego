---
title: "Porkbun"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: porkbun
dnsprovider:
  since:    "v4.4.0"
  code:     "porkbun"
  url:      "https://porkbun.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/porkbun/porkbun.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Porkbun](https://porkbun.com/).


<!--more-->

- Code: `porkbun`
- Since: v4.4.0


Here is an example bash command using the Porkbun provider:

```bash
PORKBUN_SECRET_API_KEY=xxxxxx \
PORKBUN_API_KEY=yyyyyy \
lego --email you@example.com --dns porkbun --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `PORKBUN_API_KEY` | API key |
| `PORKBUN_SECRET_API_KEY` | secret API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `PORKBUN_HTTP_TIMEOUT` | API request timeout |
| `PORKBUN_POLLING_INTERVAL` | Time between DNS propagation check |
| `PORKBUN_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `PORKBUN_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://porkbun.com/api/json/v3/documentation)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/porkbun/porkbun.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
