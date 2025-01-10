---
title: "West.cn/西部数码"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: westcn
dnsprovider:
  since:    "v4.21.0"
  code:     "westcn"
  url:      "https://www.west.cn"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/westcn/westcn.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [West.cn/西部数码](https://www.west.cn).


<!--more-->

- Code: `westcn`
- Since: v4.21.0


Here is an example bash command using the West.cn/西部数码 provider:

```bash
WESTCN_USERNAME="xxx" \
WESTCN_PASSWORD="yyy" \
lego --email you@example.com --dns westcn -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `WESTCN_PASSWORD` | API password |
| `WESTCN_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `WESTCN_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `WESTCN_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `WESTCN_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `WESTCN_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.west.cn/CustomerCenter/doc/domain_v2.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/westcn/westcn.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
