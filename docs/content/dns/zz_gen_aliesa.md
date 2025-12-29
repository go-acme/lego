---
title: "AlibabaCloud ESA"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: aliesa
dnsprovider:
  since:    "v4.29.0"
  code:     "aliesa"
  url:      "https://www.alibabacloud.com/en/product/esa"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/aliesa/aliesa.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [AlibabaCloud ESA](https://www.alibabacloud.com/en/product/esa).


<!--more-->

- Code: `aliesa`
- Since: v4.29.0


Here is an example bash command using the AlibabaCloud ESA provider:

```bash
# Setup using instance RAM role
ALIESA_RAM_ROLE=lego \
lego --dns aliesa -d '*.example.com' -d example.com run

# Or, using credentials
ALIESA_ACCESS_KEY=abcdefghijklmnopqrstuvwx \
ALIESA_SECRET_KEY=your-secret-key \
ALIESA_SECURITY_TOKEN=your-sts-token \
lego --dns aliesa - -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ALIESA_ACCESS_KEY` | Access key ID |
| `ALIESA_RAM_ROLE` | Your instance RAM role (https://www.alibabacloud.com/help/en/ecs/user-guide/attach-an-instance-ram-role-to-an-ecs-instance) |
| `ALIESA_SECRET_KEY` | Access Key secret |
| `ALIESA_SECURITY_TOKEN` | STS Security Token (optional) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ALIESA_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `ALIESA_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `ALIESA_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `ALIESA_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.alibabacloud.com/help/en/edge-security-acceleration/esa/api-esa-2024-09-10-overview?spm=a2c63.p38356.help-menu-2673927.d_6_0_0.20b224c28PSZDc#:~:text=DNS-,DNS%20records,-DNS%20records)
- [Go client](https://github.com/alibabacloud-go/esa-20240910)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/aliesa/aliesa.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
