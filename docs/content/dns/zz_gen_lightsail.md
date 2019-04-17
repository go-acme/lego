---
title: "Amazon Lightsail"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: lightsail
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/lightsail/lightsail.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.5.0

Configuration for [Amazon Lightsail](https://aws.amazon.com/lightsail/).


<!--more-->

- Code: `lightsail`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `AWS_ACCESS_KEY_ID` | Access key ID |
| `AWS_SECRET_ACCESS_KEY` | Secret access key |
| `DNS_ZONE` | DNS zone |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `LIGHTSAIL_POLLING_INTERVAL` | Time between DNS propagation check |
| `LIGHTSAIL_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information


- [Go client](https://github.com/aws/aws-sdk-go/aws)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/lightsail/lightsail.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
