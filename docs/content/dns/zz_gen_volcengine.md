---
title: "Volcano Engine/火山引擎"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: volcengine
dnsprovider:
  since:    "v4.19.0"
  code:     "volcengine"
  url:      "https://www.volcengine.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/volcengine/volcengine.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Volcano Engine/火山引擎](https://www.volcengine.com/).


<!--more-->

- Code: `volcengine`
- Since: v4.19.0


Here is an example bash command using the Volcano Engine/火山引擎 provider:

```bash
VOLC_ACCESSKEY=xxx \
VOLC_SECRETKEY=yyy \
lego --email you@example.com --dns volcengine -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `VOLC_ACCESSKEY` | Access Key ID (AK) |
| `VOLC_SECRETKEY` | Secret Access Key (SK) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `VOLC_HOST` | API host |
| `VOLC_HTTP_TIMEOUT` | API request timeout |
| `VOLC_POLLING_INTERVAL` | Time between DNS propagation check |
| `VOLC_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `VOLC_REGION` | Region |
| `VOLC_SCHEME` | API scheme |
| `VOLC_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.volcengine.com/docs/6758/155086)
- [Go client](https://github.com/volcengine/volc-sdk-golang)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/volcengine/volcengine.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
