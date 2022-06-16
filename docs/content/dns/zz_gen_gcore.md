---
title: "G-Core Labs"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: gcore
dnsprovider:
  since:    "v4.5.0"
  code:     "gcore"
  url:      "https://gcorelabs.com/dns/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gcore/gcore.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [G-Core Labs](https://gcorelabs.com/dns/).


<!--more-->

- Code: `gcore`
- Since: v4.5.0


Here is an example bash command using the G-Core Labs provider:

```bash
GCORE_PERMANENT_API_TOKEN=xxxxx \
lego --email you@example.com --dns gcore --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `GCORE_PERMANENT_API_TOKEN` | Permanent API tokene (https://gcorelabs.com/blog/permanent-api-token-explained/) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GCORE_HTTP_TIMEOUT` | API request timeout |
| `GCORE_POLLING_INTERVAL` | Time between DNS propagation check |
| `GCORE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `GCORE_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://dnsapi.gcorelabs.com/docs#tag/zonesV2)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gcore/gcore.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
