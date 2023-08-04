---
title: "Yandex 360"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: yandex360
dnsprovider:
  since:    "v4.14.0"
  code:     "yandex360"
  url:      "https://360.yandex.ru"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/yandex360/yandex360.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Yandex 360](https://360.yandex.ru).


<!--more-->

- Code: `yandex360`
- Since: v4.14.0


Here is an example bash command using the Yandex 360 provider:

```bash
YANDEX360_OAUTH_TOKEN=<your OAuth Token> \
YANDEX360_ORG_ID=<your organization ID> \
lego --email you@example.com --dns yandex360 --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `YANDEX360_OAUTH_TOKEN` | The OAuth Token |
| `YANDEX360_ORG_ID` | The organization ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `YANDEX360_HTTP_TIMEOUT` | API request timeout |
| `YANDEX360_POLLING_INTERVAL` | Time between DNS propagation check |
| `YANDEX360_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `YANDEX360_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://yandex.ru/dev/api360/doc/ref/DomainDNSService.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/yandex360/yandex360.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
