---
title: "Huawei Cloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: huaweicloud
dnsprovider:
  since:    "v4.19"
  code:     "huaweicloud"
  url:      "https://huaweicloud.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/huaweicloud/huaweicloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Huawei Cloud](https://huaweicloud.com).


<!--more-->

- Code: `huaweicloud`
- Since: v4.19


Here is an example bash command using the Huawei Cloud provider:

```bash
HUAWEICLOUD_ACCESS_KEY_ID=your-access-key-id \
HUAWEICLOUD_SECRET_ACCESS_KEY=your-secret-access-key \
HUAWEICLOUD_REGION=cn-south-1 \
lego --email you@example.com --dns huaweicloud -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `HUAWEICLOUD_ACCESS_KEY_ID` | Access key ID |
| `HUAWEICLOUD_REGION` | Region |
| `HUAWEICLOUD_SECRET_ACCESS_KEY` | Access Key secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `HUAWEICLOUD_HTTP_TIMEOUT` | API request timeout |
| `HUAWEICLOUD_POLLING_INTERVAL` | Time between DNS propagation check |
| `HUAWEICLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `HUAWEICLOUD_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://console-intl.huaweicloud.com/apiexplorer/#/openapi/DNS/doc?locale=en-us)
- [Go client](https://github.com/huaweicloud/huaweicloud-sdk-go-v3)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/huaweicloud/huaweicloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
