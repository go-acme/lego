---
title: "Stackpath"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: stackpath
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/stackpath/stackpath.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v1.1.0

Configuration for [Stackpath](https://www.stackpath.com/).


<!--more-->

- Code: `stackpath`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `STACKPATH_CLIENT_ID` | Client ID |
| `STACKPATH_CLIENT_SECRET` | Client secret |
| `STACKPATH_STACK_ID` | Stack ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `STACKPATH_POLLING_INTERVAL` | Time between DNS propagation check |
| `STACKPATH_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `STACKPATH_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://developer.stackpath.com/en/api/dns/#tag/Zone)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/stackpath/stackpath.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
