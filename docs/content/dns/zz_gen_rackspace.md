---
title: "Rackspace"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: rackspace
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rackspace/rackspace.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.4.0

Configuration for [Rackspace](https://www.rackspace.com/).


<!--more-->

- Code: `rackspace`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `RACKSPACE_API_KEY` | API key |
| `RACKSPACE_USER` | API user |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `RACKSPACE_HTTP_TIMEOUT` | API request timeout |
| `RACKSPACE_POLLING_INTERVAL` | Time between DNS propagation check |
| `RACKSPACE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `RACKSPACE_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://developer.rackspace.com/docs/cloud-dns/v1/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rackspace/rackspace.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
