---
title: "Constellix"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: constellix
dnsprovider:
  since:    "v3.4.0"
  code:     "constellix"
  url:      "https://constellix.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/constellix/constellix.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Constellix](https://constellix.com).


<!--more-->

- Code: `constellix`
- Since: v3.4.0


Here is an example bash command using the Constellix provider:

```bash
CONSTELLIX_API_KEY=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
CONSTELLIX_SECRET_KEY=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
lego --email you@example.com --dns constellix -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `CONSTELLIX_API_KEY` | User API key |
| `CONSTELLIX_SECRET_KEY` | User secret key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `CONSTELLIX_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `CONSTELLIX_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `CONSTELLIX_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `CONSTELLIX_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api-docs.constellix.com)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/constellix/constellix.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
