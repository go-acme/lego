---
title: "all-inkl"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: allinkl
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/allinkl/allinkl.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v4.5.0

Configuration for [all-inkl](https://all-inkl.com).


<!--more-->

- Code: `allinkl`

Here is an example bash command using the all-inkl provider:

```bash
ALL_INKL_LOGIN=xxxxxxxxxxxxxxxxxxxxxxxxxx \
ALL_INKL_PASSWORD=yyyyyyyyyyyyyyyyyyyyyyyyyy \
lego --email myemail@example.com --dns allinkl --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ALL_INKL_API_KEY` | API login |
| `ALL_INKL_PASSWORD` | API password |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ALL_INKL_HTTP_TIMEOUT` | API request timeout |
| `ALL_INKL_POLLING_INTERVAL` | Time between DNS propagation check |
| `ALL_INKL_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://kasapi.kasserver.com/dokumentation/phpdoc/index.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/allinkl/allinkl.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
