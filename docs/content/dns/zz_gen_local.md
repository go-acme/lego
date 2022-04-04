---
title: "Local"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: local
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/local/local.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.0.1
Setup local udp dns server that can serve forawrded requests from main dns server


<!--more-->

- Code: `local`

Here is an example bash command using the Local provider:

```bash
LOCAL_LISTEN=:5353 \
lego --email myemail@example.com --dns local --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `LOCAL_LISTEN` | Listen udp dns-server address |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).







<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/local/local.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
