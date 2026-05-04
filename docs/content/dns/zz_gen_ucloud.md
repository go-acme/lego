---
title: "UCloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: ucloud
dnsprovider:
  since:    "v4.34.0"
  code:     "ucloud"
  url:      "https://www.ucloud.cn/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ucloud/ucloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [UCloud](https://www.ucloud.cn/).


<!--more-->

- Code: `ucloud`
- Since: v4.34.0


Here is an example bash command using the UCloud provider:

```bash
UCLOUD_PUBLIC_KEY="xxx" \
UCLOUD_PRIVATE_KEY="yyy" \
lego run --dns ucloud -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `UCLOUD_PRIVATE_KEY` | Private key |
| `UCLOUD_PUBLIC_KEY` | Public key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `UCLOUD_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `UCLOUD_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `UCLOUD_PROJECT_ID` | Project ID |
| `UCLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `UCLOUD_REGION` | Region |
| `UCLOUD_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 600) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.ucloud.cn/api/udnr-api/README)
- [Go client](https://github.com/ucloud/ucloud-sdk-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ucloud/ucloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
