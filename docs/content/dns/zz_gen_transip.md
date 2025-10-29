---
title: "TransIP"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: transip
dnsprovider:
  since:    "v2.0.0"
  code:     "transip"
  url:      "https://www.transip.nl/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/transip/transip.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [TransIP](https://www.transip.nl/).


<!--more-->

- Code: `transip`
- Since: v2.0.0


Here is an example bash command using the TransIP provider:

```bash
TRANSIP_ACCOUNT_NAME = "Account name" \
TRANSIP_PRIVATE_KEY_PATH = "transip.key" \
lego --email you@example.com --dns transip -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `TRANSIP_ACCOUNT_NAME` | Account name |
| `TRANSIP_PRIVATE_KEY_PATH` | Private key path |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `TRANSIP_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `TRANSIP_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `TRANSIP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 600) |
| `TRANSIP_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 10) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.transip.eu/rest/docs.html)
- [Go client](https://github.com/transip/gotransip)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/transip/transip.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
