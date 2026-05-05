---
title: "Ngenix"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: ngenix
dnsprovider:
  since:    "v5.0.0"
  code:     "ngenix"
  url:      "https://ngenix.net"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ngenix/ngenix.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Ngenix](https://ngenix.net).


<!--more-->

- Code: `ngenix`
- Since: v5.0.0


Here is an example bash command using the Ngenix provider:

```bash
NGENIX_USERNAME="xxx" \
NGENIX_TOKEN="yyy" \
NGENIX_CUSTOMER_ID="zzz" \
lego run --dns ngenix -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NGENIX_CUSTOMER_ID` | Customer ID |
| `NGENIX_TOKEN` | API token |
| `NGENIX_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NGENIX_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `NGENIX_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 20) |
| `NGENIX_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 600) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://help.ngenix.net/articles/#!apidocs/upravlenie-dns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ngenix/ngenix.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
