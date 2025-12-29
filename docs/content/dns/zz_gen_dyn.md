---
title: "Dyn"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dyn
dnsprovider:
  since:    "v0.3.0"
  code:     "dyn"
  url:      "https://dyn.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dyn/dyn.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Dyn](https://dyn.com/).


<!--more-->

- Code: `dyn`
- Since: v0.3.0


Here is an example bash command using the Dyn provider:

```bash
DYN_CUSTOMER_NAME=xxxxxx \
DYN_USER_NAME=yyyyy \
DYN_PASSWORD=zzzz \
lego --dns dyn -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DYN_CUSTOMER_NAME` | Customer name |
| `DYN_PASSWORD` | Password |
| `DYN_USER_NAME` | User name |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DYN_HTTP_TIMEOUT` | API request timeout in seconds (Default: 10) |
| `DYN_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `DYN_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `DYN_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://help.dyn.com/rest/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dyn/dyn.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
