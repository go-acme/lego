---
title: "DreamHost"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dreamhost
dnsprovider:
  since:    "v1.1.0"
  code:     "dreamhost"
  url:      "https://www.dreamhost.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dreamhost/dreamhost.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [DreamHost](https://www.dreamhost.com).


<!--more-->

- Code: `dreamhost`
- Since: v1.1.0


Here is an example bash command using the DreamHost provider:

```bash
DREAMHOST_API_KEY="YOURAPIKEY" \
lego --email you@example.com --dns dreamhost -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DREAMHOST_API_KEY` | The API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DREAMHOST_HTTP_TIMEOUT` | API request timeout |
| `DREAMHOST_POLLING_INTERVAL` | Time between DNS propagation check |
| `DREAMHOST_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DREAMHOST_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://help.dreamhost.com/hc/en-us/articles/217560167-API_overview)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dreamhost/dreamhost.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
