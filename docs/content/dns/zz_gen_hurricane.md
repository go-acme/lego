---
title: "Hurricane Electric DNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: hurricane
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hurricane/hurricane.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: 

Configuration for [Hurricane Electric DNS](https://dns.he.net/).


<!--more-->

- Code: `hurricane`

Here is an example bash command using the Hurricane Electric DNS provider:

```bash
HURRICANE_TOKEN=xxxxxx \
lego --email myemail@example.com --dns he --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `HURRICANE_TOKEN` | Account token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).






## More information

- [API documentation](https://dns.he.org/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hurricane/hurricane.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
