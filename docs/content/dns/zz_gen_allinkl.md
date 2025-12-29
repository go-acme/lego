---
title: "all-inkl"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: allinkl
dnsprovider:
  since:    "v4.5.0"
  code:     "allinkl"
  url:      "https://all-inkl.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/allinkl/allinkl.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [all-inkl](https://all-inkl.com).


<!--more-->

- Code: `allinkl`
- Since: v4.5.0


Here is an example bash command using the all-inkl provider:

```bash
ALL_INKL_LOGIN=xxxxxxxxxxxxxxxxxxxxxxxxxx \
ALL_INKL_PASSWORD=yyyyyyyyyyyyyyyyyyyyyyyyyy \
lego --dns allinkl -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ALL_INKL_LOGIN` | KAS login |
| `ALL_INKL_PASSWORD` | KAS password |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ALL_INKL_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `ALL_INKL_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `ALL_INKL_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://kasapi.kasserver.com/dokumentation/phpdoc/index.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/allinkl/allinkl.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
