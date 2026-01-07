---
title: "JD Cloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: jdcloud
dnsprovider:
  since:    "v4.31.0"
  code:     "jdcloud"
  url:      "https://www.jdcloud.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/jdcloud/jdcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [JD Cloud](https://www.jdcloud.com/).


<!--more-->

- Code: `jdcloud`
- Since: v4.31.0


Here is an example bash command using the JD Cloud provider:

```bash
JDCLOUD_ACCESS_KEY_ID="xxx" \
JDCLOUD_ACCESS_KEY_SECRET="yyy" \
lego --dns jdcloud -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `JDCLOUD_ACCESS_KEY_ID` | Access key ID |
| `JDCLOUD_ACCESS_KEY_SECRET` | Access key secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `JDCLOUD_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `JDCLOUD_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `JDCLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `JDCLOUD_REGION_ID` | Region ID (Default: cn-north-1) |
| `JDCLOUD_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.jdcloud.com/cn/jd-cloud-dns/api/overview)
- [Go client](https://github.com/jdcloud-api/jdcloud-sdk-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/jdcloud/jdcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
