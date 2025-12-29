---
title: "DynDnsFree.de"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dyndnsfree
dnsprovider:
  since:    "v4.23.0"
  code:     "dyndnsfree"
  url:      "https://www.dyndnsfree.de"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dyndnsfree/dyndnsfree.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [DynDnsFree.de](https://www.dyndnsfree.de).


<!--more-->

- Code: `dyndnsfree`
- Since: v4.23.0


Here is an example bash command using the DynDnsFree.de provider:

```bash
DYNDNSFREE_USERNAME="xxx" \
DYNDNSFREE_PASSWORD="yyy" \
lego --dns dyndnsfree -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DYNDNSFREE_PASSWORD` | Password |
| `DYNDNSFREE_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DYNDNSFREE_HTTP_TIMEOUT` | Request timeout in seconds (Default: 30) |
| `DYNDNSFREE_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `DYNDNSFREE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.dyndnsfree.de/user/hilfe.php?hsm=2)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dyndnsfree/dyndnsfree.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
