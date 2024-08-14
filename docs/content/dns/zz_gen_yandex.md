---
title: "Yandex PDD"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: yandex
dnsprovider:
  since:    "v3.7.0"
  code:     "yandex"
  url:      "https://pdd.yandex.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/yandex/yandex.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Yandex PDD](https://pdd.yandex.com).


<!--more-->

- Code: `yandex`
- Since: v3.7.0


Here is an example bash command using the Yandex PDD provider:

```bash
YANDEX_PDD_TOKEN=<your PDD Token> \
lego --email you@example.com --dns yandex --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `YANDEX_PDD_TOKEN` | Basic authentication username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `YANDEX_HTTP_TIMEOUT` | API request timeout |
| `YANDEX_POLLING_INTERVAL` | Time between DNS propagation check |
| `YANDEX_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `YANDEX_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://yandex.com/dev/domain/doc/concepts/api-dns.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/yandex/yandex.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
