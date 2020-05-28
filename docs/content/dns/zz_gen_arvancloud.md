---
title: "ArvanCloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: arvancloud
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/arvancloud/arvancloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v3.8.0

Configuration for [ArvanCloud](https://arvancloud.com).


<!--more-->

- Code: `arvancloud`

Here is an example bash command using the ArvanCloud provider:

```bash
ARVANCLOUD_API_KEY=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
lego --dns arvancloud --domains my.domain.com --email my@email.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ARVANCLOUD_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ARVANCLOUD_HTTP_TIMEOUT` | API request timeout |
| `ARVANCLOUD_POLLING_INTERVAL` | Time between DNS propagation check |
| `ARVANCLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `ARVANCLOUD_SEQUENCE_INTERVAL` | Interval between iteration |
| `ARVANCLOUD_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://www.arvancloud.com/docs/api/cdn/4.0)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/arvancloud/arvancloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
