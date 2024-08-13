---
title: "Tencent Cloud DNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: tencentcloud
dnsprovider:
  since:    "v4.6.0"
  code:     "tencentcloud"
  url:      "https://cloud.tencent.com/product/cns"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/tencentcloud/tencentcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Tencent Cloud DNS](https://cloud.tencent.com/product/cns).


<!--more-->

- Code: `tencentcloud`
- Since: v4.6.0


Here is an example bash command using the Tencent Cloud DNS provider:

```bash
TENCENTCLOUD_SECRET_ID=abcdefghijklmnopqrstuvwx \
TENCENTCLOUD_SECRET_KEY=your-secret-key \
lego --email you@example.com --dns tencentcloud --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `TENCENTCLOUD_SECRET_ID` | Access key ID |
| `TENCENTCLOUD_SECRET_KEY` | Access Key secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `TENCENTCLOUD_HTTP_TIMEOUT` | API request timeout |
| `TENCENTCLOUD_POLLING_INTERVAL` | Time between DNS propagation check |
| `TENCENTCLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `TENCENTCLOUD_REGION` | Region |
| `TENCENTCLOUD_SESSION_TOKEN` | Access Key token |
| `TENCENTCLOUD_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://cloud.tencent.com/document/product/1427/56153)
- [Go client](https://github.com/tencentcloud/tencentcloud-sdk-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/tencentcloud/tencentcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
