---
title: "Baidu Cloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: baiducloud
dnsprovider:
  since:    "v4.23.0"
  code:     "baiducloud"
  url:      "https://cloud.baidu.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/baiducloud/baiducloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Baidu Cloud](https://cloud.baidu.com).


<!--more-->

- Code: `baiducloud`
- Since: v4.23.0


Here is an example bash command using the Baidu Cloud provider:

```bash
BAIDUCLOUD_ACCESS_KEY_ID="xxx" \
BAIDUCLOUD_SECRET_ACCESS_KEY="yyy" \
lego --email you@example.com --dns baiducloud -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `BAIDUCLOUD_ACCESS_KEY_ID` | Access key |
| `BAIDUCLOUD_SECRET_ACCESS_KEY` | Secret access key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `BAIDUCLOUD_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `BAIDUCLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `BAIDUCLOUD_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://cloud.baidu.com/doc/DNS/s/El4s7lssr)
- [Go client](https://github.com/baidubce/bce-sdk-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/baiducloud/baiducloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
