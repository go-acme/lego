---
title: "35.com/三五互联"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: com35
dnsprovider:
  since:    "v4.31.0"
  code:     "com35"
  url:      "https://www.35.cn/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/com35/com35.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [35.com/三五互联](https://www.35.cn/).


<!--more-->

- Code: `com35`
- Since: v4.31.0


Here is an example bash command using the 35.com/三五互联 provider:

```bash
COM35_USERNAME="xxx" \
COM35_PASSWORD="yyy" \
lego --dns com35 -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `COM35_PASSWORD` | API password |
| `COM35_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `COM35_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `COM35_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `COM35_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `COM35_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.35.cn/CustomerCenter/doc/domain_v2.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/com35/com35.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
