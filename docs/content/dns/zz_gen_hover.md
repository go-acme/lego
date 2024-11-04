---
title: "Hover (tucows)"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: hover
dnsprovider:
  since:    "v4.20.0"
  code:     "hover"
  url:      "https://hover.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hover/hover.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Hover (tucows)](https://hover.com).


<!--more-->

- Code: `hover`
- Since: v4.20.0


Here is an example bash command using the Hover (tucows) provider:

```bash
HOVER_USERNAME=xxx \
HOVER_PASSWORD=yyy \
lego --email you@example.com --dns hover --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `HOVER_PASSWORD` | Password (plaintext) |
| `HOVER_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `HOVER_HTTP_TIMEOUT` | API request timeout |
| `HOVER_POLLING_INTERVAL` | Time between DNS propagation check |
| `HOVER_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).





<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hover/hover.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
