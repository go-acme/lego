---
title: "Cloud.ru"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: cloudru
dnsprovider:
  since:    "v4.14.0"
  code:     "cloudru"
  url:      "https://cloud.ru"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/cloudru/cloudru.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Cloud.ru](https://cloud.ru).


<!--more-->

- Code: `cloudru`
- Since: v4.14.0


Here is an example bash command using the Cloud.ru provider:

```bash
CLOUDRU_SERVICE_INSTANCE_ID=ppp \
CLOUDRU_KEY_ID=xxx \
CLOUDRU_SECRET=yyy \
lego --email you@example.com --dns cloudru -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `CLOUDRU_KEY_ID` | Key ID (login) |
| `CLOUDRU_SECRET` | Key Secret |
| `CLOUDRU_SERVICE_INSTANCE_ID` | Service Instance ID (parentId) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `CLOUDRU_HTTP_TIMEOUT` | API request timeout |
| `CLOUDRU_POLLING_INTERVAL` | Time between DNS propagation check |
| `CLOUDRU_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `CLOUDRU_SEQUENCE_INTERVAL` | Time between sequential requests |
| `CLOUDRU_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://cloud.ru/ru/docs/clouddns/ug/topics/api-ref.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/cloudru/cloudru.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
