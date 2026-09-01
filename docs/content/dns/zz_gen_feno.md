---
title: "FENO"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: feno
dnsprovider:
  since:    "v5.5.0"
  code:     "feno"
  url:      "https://feno.no/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/feno/feno.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [FENO](https://feno.no/).


<!--more-->

- Code: `feno`
- Since: v5.5.0


Here is an example bash command using the FENO provider:

```bash
FENO_API_KEY="feno_live_xxx" \
lego --dns feno -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `FENO_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `FENO_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `FENO_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `FENO_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `FENO_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

The API key should have the `acme:write` scope only, and the domain must use the FENO nameservers.



## More information

- [API documentation](https://github.com/mrerikcodes/feno-api/blob/main/openapi.yaml)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/feno/feno.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
