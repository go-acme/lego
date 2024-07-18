---
title: "directadmin"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: directadmin
dnsprovider:
  since:    ""
  code:     "directadmin"
  url:      "directadmin"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/directadmin/directadmin.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

directadmin api


<!--more-->

- Code: `directadmin`
- Since: 


Here is an example bash command using the directadmin provider:

```bash
DIRECTADMIN_API_URL="http://example.com:2222" \
DIRECTADMIN_USERNAME=xxxx \
DIRECTADMIN_PASSWORD=yyy \
lego --email you@example.com --dns directadmin -d "my.example.org" -d "*.example.org" run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DIRECTADMIN_API_URL` | URL of the API |
| `DIRECTADMIN_PASSWORD` | API password |
| `DIRECTADMIN_USERNAME` | API username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).






## More information

- [API documentation](https://www.directadmin.com/api.php)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/directadmin/directadmin.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
