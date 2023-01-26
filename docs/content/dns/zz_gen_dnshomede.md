---
title: "dnsHome.de"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dnshomede
dnsprovider:
  since:    "v4.10.0"
  code:     "dnshomede"
  url:      "https://www.dnshome.de"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnshomede/dnshomede.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [dnsHome.de](https://www.dnshome.de).


<!--more-->

- Code: `dnshomede`
- Since: v4.10.0


Here is an example bash command using the dnsHome.de provider:

```bash
DNSHOMEDE_CREDENTIALS=sub.example.org:password \
lego --email you@example.com --dns dnshomede --domains example.org --domains '*.example.org' run

DNSHOMEDE_CREDENTIALS=my.example.org:password1,demo.example.org:password2 \
lego --email you@example.com --dns dnshomede --domains my.example.org --domains demo.example.org
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DNSHOMEDE_TOKENS` | TXT record names and tokens |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).







<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnshomede/dnshomede.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
