---
title: "Linode (v4)"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: linode
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/linode/linode.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v1.1.0

Configuration for [Linode (v4)](https://www.linode.com/).


<!--more-->

- Code: `linode`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `LINODE_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `LINODE_HTTP_TIMEOUT` | API request timeout |
| `LINODE_POLLING_INTERVAL` | Time between DNS propagation check |
| `LINODE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `LINODE_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://developers.linode.com/api/v4)
- [Go client](https://github.com/linode/linodego)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/linode/linode.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
