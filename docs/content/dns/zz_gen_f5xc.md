---
title: "F5 XC"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: f5xc
dnsprovider:
  since:    "v4.23.0"
  code:     "f5xc"
  url:      "https://www.f5.com/products/distributed-cloud-services"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/f5xc/f5xc.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [F5 XC](https://www.f5.com/products/distributed-cloud-services).


<!--more-->

- Code: `f5xc`
- Since: v4.23.0


Here is an example bash command using the F5 XC provider:

```bash
F5XC_API_TOKEN="xxx" \
F5XC_TENANT_NAME="yyy" \
F5XC_GROUP_NAME="zzz" \
lego --email you@example.com --dns f5xc -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `F5XC_API_TOKEN` | API token |
| `F5XC_GROUP_NAME` | Group name |
| `F5XC_TENANT_NAME` | XC Tenant shortname |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `F5XC_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `F5XC_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `F5XC_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `F5XC_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.cloud.f5.com/docs-v2/api/dns-zone-rrset)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/f5xc/f5xc.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
