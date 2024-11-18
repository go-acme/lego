---
title: "Rain Yun/雨云"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: rainyun
dnsprovider:
  since:    "v4.21.0"
  code:     "rainyun"
  url:      "https://www.rainyun.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rainyun/rainyun.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Rain Yun/雨云](https://www.rainyun.com).


<!--more-->

- Code: `rainyun`
- Since: v4.21.0


Here is an example bash command using the Rain Yun/雨云 provider:

```bash
RAINYUN_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego --email you@example.com --dns rainyun -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `RAINYUN_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `RAINYUN_HTTP_TIMEOUT` | API request timeout |
| `RAINYUN_POLLING_INTERVAL` | Time between DNS propagation check |
| `RAINYUN_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `RAINYUN_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.apifox.cn/apidoc/shared-a4595cc8-44c5-4678-a2a3-eed7738dab03/api-151416609)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rainyun/rainyun.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
