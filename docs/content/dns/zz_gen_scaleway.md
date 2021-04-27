---
title: "Scaleway"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: scaleway
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/scaleway/scaleway.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v3.4.0

Configuration for [Scaleway](https://developers.scaleway.com/).


<!--more-->

- Code: `scaleway`

Here is an example bash command using the Scaleway provider:

```bash
SCALEWAY_API_TOKEN=xxxxxxx-xxxxx-xxxx-xxx-xxxxxx \
lego --email myemail@example.com --dns scaleway --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SCALEWAY_API_TOKEN` | API token |
| `SCALEWAY_PROJECT_ID` | Project to use (optional) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SCALEWAY_POLLING_INTERVAL` | Time between DNS propagation check |
| `SCALEWAY_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `SCALEWAY_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://developers.scaleway.com/en/products/domain/dns/api/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/scaleway/scaleway.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
