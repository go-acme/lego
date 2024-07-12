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
lego --email you@example.com --dns alidns --domains my.example.org run

# Or, using credentials
ALICLOUD_ACCESS_KEY=abcdefghijklmnopqrstuvwx \
ALICLOUD_SECRET_KEY=your-secret-key \
ALICLOUD_SECURITY_TOKEN=your-sts-token \
lego --email you@example.com --dns alidns --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ALICLOUD_ACCESS_KEY` | Access key ID |
| `ALICLOUD_RAM_ROLE` | Your instance RAM role (https://www.alibabacloud.com/help/doc-detail/54579.htm) |
| `ALICLOUD_SECRET_KEY` | Access Key secret |
| `ALICLOUD_SECURITY_TOKEN` | STS Security Token (optional) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ALICLOUD_HTTP_TIMEOUT` | API request timeout |
| `ALICLOUD_POLLING_INTERVAL` | Time between DNS propagation check |
| `ALICLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `ALICLOUD_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://www.alibabacloud.com/help/en/alibaba-cloud-dns/latest/api-alidns-2015-01-09-dir-parsing-records)
- [Go client](https://github.com/aliyun/alibaba-cloud-sdk-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/alidns/alidns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
