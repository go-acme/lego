---
title: "ArvanCloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: arvancloud
dnsprovider:
  since:    "v3.8.0"
  code:     "arvancloud"
  url:      "https://arvancloud.ir"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/arvancloud/arvancloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [ArvanCloud](https://arvancloud.ir).


<!--more-->

- Code: `arvancloud`
- Since: v3.8.0


Here is an example bash command using the ArvanCloud provider:

```bash
ARVANCLOUD_API_KEY="Apikey xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
lego --dns arvancloud -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ARVANCLOUD_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ARVANCLOUD_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `ARVANCLOUD_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `ARVANCLOUD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `ARVANCLOUD_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 600) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.arvancloud.ir/docs/api/cdn/4.0)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/arvancloud/arvancloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
