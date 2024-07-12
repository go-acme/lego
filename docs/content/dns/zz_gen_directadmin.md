---
title: "directadmin"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: directadmin
dnsprovider:
  since:    "v0.3.0"
  code:     "directadmin"
  url:      "directadmin"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/directadmin/directadmin.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [directadmin](directadmin).


<!--more-->

- Code: `directadmin`
- Since: v0.3.0


Here is an example bash command using the directadmin provider:

```bash
DO_AUTH_TOKEN=xxxxxx \
lego --email you@example.com --dns directadmin --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `password` | yourpassword |
| `username` | yourusername |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `api_url` | The URL of the API |
| `timeout` | API request timeout |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](directadmin)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/directadmin/directadmin.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
