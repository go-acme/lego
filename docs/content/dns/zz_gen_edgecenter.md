---
title: "EdgeCenter"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: edgecenter
dnsprovider:
  since:    "v4.29.0"
  code:     "edgecenter"
  url:      "https://edgecenter.ru/dns"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/edgecenter/edgecenter.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [EdgeCenter](https://edgecenter.ru/dns).


<!--more-->

- Code: `edgecenter`
- Since: v4.29.0


Here is an example bash command using the EdgeCenter provider:

```bash
EDGECENTER_PERMANENT_API_TOKEN=xxxxx \
lego --dns edgecenter -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `EDGECENTER_PERMANENT_API_TOKEN` | Permanent API token (https://edgecenter.ru/blog/permanent-api-token-explained/) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `EDGECENTER_HTTP_TIMEOUT` | API request timeout in seconds (Default: 10) |
| `EDGECENTER_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 20) |
| `EDGECENTER_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 360) |
| `EDGECENTER_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://apidocs.edgecenter.ru/dns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/edgecenter/edgecenter.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
