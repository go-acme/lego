---
title: "Yandex"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: yandex
dnsprovider:
  since:    "v3.7.0"
  code:     "yandex"
  url:      "https://yandex.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/yandex/yandex.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Yandex](https://yandex.com/).


<!--more-->

- Code: `yandex`
- Since: v3.7.0


Here is an example bash command using the Yandex provider:

```bash
YANDEX_PDD_TOKEN=<your PDD Token> \
lego --email you@example.com --dns yandex --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `YANDEX_PDD_TOKEN` | Basic authentication username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `YANDEX_HTTP_TIMEOUT` | API request timeout |
| `YANDEX_POLLING_INTERVAL` | Time between DNS propagation check |
| `YANDEX_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `YANDEX_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://tech.yandex.com/domain/doc/concepts/api-dns-docpage/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/yandex/yandex.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
