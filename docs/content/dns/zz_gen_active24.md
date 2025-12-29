---
title: "Active24"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: active24
dnsprovider:
  since:    "v4.23.0"
  code:     "active24"
  url:      "https://www.active24.cz"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/active24/active24.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Active24](https://www.active24.cz).


<!--more-->

- Code: `active24`
- Since: v4.23.0


Here is an example bash command using the Active24 provider:

```bash
ACTIVE24_API_KEY="xxx" \
ACTIVE24_SECRET="yyy" \
lego --dns active24 -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ACTIVE24_API_KEY` | API key |
| `ACTIVE24_SECRET` | Secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ACTIVE24_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `ACTIVE24_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `ACTIVE24_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `ACTIVE24_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://rest.active24.cz/v2/docs)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/active24/active24.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
