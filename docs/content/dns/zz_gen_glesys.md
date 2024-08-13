---
title: "Glesys"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: glesys
dnsprovider:
  since:    "v0.5.0"
  code:     "glesys"
  url:      "https://glesys.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/glesys/glesys.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Glesys](https://glesys.com/).


<!--more-->

- Code: `glesys`
- Since: v0.5.0


Here is an example bash command using the Glesys provider:

```bash
GLESYS_API_USER=xxxxx \
GLESYS_API_KEY=yyyyy \
lego --email you@example.com --dns glesys --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `GLESYS_API_KEY` | API key |
| `GLESYS_API_USER` | API user |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GLESYS_HTTP_TIMEOUT` | API request timeout |
| `GLESYS_POLLING_INTERVAL` | Time between DNS propagation check |
| `GLESYS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `GLESYS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://github.com/GleSYS/API/wiki/API-Documentation)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/glesys/glesys.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
