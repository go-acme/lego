---
title: "Alibaba Cloud DNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: alidns
dnsprovider:
  since:    "v1.1.0"
  code:     "alidns"
  url:      "https://www.alibabacloud.com/product/dns"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/alidns/alidns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Alibaba Cloud DNS](https://www.alibabacloud.com/product/dns).


<!--more-->

- Code: `alidns`
- Since: v1.1.0


Here is an example bash command using the Alibaba Cloud DNS provider:

```bash
# Setup using instance RAM role
ALICLOUD_RAM_ROLE=lego \
lego --email you@example.com --dns alidns -d '*.example.com' -d example.com run

# Or, using credentials
ALICLOUD_ACCESS_KEY=abcdefghijklmnopqrstuvwx \
ALICLOUD_SECRET_KEY=your-secret-key \
ALICLOUD_SECURITY_TOKEN=your-sts-token \
lego --email you@example.com --dns alidns - -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ALICLOUD_ACCESS_KEY` | Access key ID |
| `ALICLOUD_RAM_ROLE` | Your instance RAM role (https://www.alibabacloud.com/help/en/ecs/user-guide/attach-an-instance-ram-role-to-an-ecs-instance) |
| `ALICLOUD_SECRET_KEY` | Access Key secret |
| `ALICLOUD_SECURITY_TOKEN` | STS Security Token (optional) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ALICLOUD_HTTP_TIMEOUT` | API request timeout in seconds (Default: 10) |
| `ALICLOUD_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `ALICLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `ALICLOUD_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 600) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.alibabacloud.com/help/en/alibaba-cloud-dns/latest/api-alidns-2015-01-09-dir-parsing-records)
- [Go client](https://github.com/alibabacloud-go/alidns-20150109)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/alidns/alidns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
