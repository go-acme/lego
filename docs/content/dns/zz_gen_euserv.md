---
title: "EUserv"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: euserv
dnsprovider:
  since:    "v4.32.0"
  code:     "euserv"
  url:      "https://www.euserv.com/en/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/euserv/euserv.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [EUserv](https://www.euserv.com/en/).


<!--more-->

- Code: `euserv`
- Since: v4.32.0


Here is an example bash command using the EUserv provider:

```bash
EUSERV_EMAIL="user@example.com" \
EUSERV_PASSWORD="xxx" \
EUSERV_ORDER_ID="yyy" \
lego --email you@example.com --dns euserv -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `EUSERV_EMAIL` | The customer email address. You can also use the customer id instead. |
| `EUSERV_ORDER_ID` | The order ID of the API contract that you want to use for this login session. |
| `EUSERV_PASSWORD` | The customer account password. |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `EUSERV_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `EUSERV_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `EUSERV_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `EUSERV_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://support.euserv.com/api-doc/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/euserv/euserv.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
