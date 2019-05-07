---
title: "Alibaba Cloud DNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: alidns
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/alidns/alidns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v1.1.0

Configuration for [Alibaba Cloud DNS](https://www.alibabacloud.com/product/dns).


<!--more-->

- Code: `alidns`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ALICLOUD_ACCESS_KEY` | Access key ID |
| `ALICLOUD_SECRET_KEY` | Access Key secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ALICLOUD_HTTP_TIMEOUT` | API request timeout |
| `ALICLOUD_POLLING_INTERVAL` | Time between DNS propagation check |
| `ALICLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `ALICLOUD_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://www.alibabacloud.com/help/doc-detail/42875.htm)
- [Go client](https://github.com/aliyun/alibaba-cloud-sdk-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/alidns/alidns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
