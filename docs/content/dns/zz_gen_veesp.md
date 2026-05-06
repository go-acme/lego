---
title: "Veesp"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: veesp
dnsprovider:
  since:    "v5.0.0"
  code:     "veesp"
  url:      "https://veesp.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/veesp/veesp.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Veesp](https://veesp.com).


<!--more-->

- Code: `veesp`
- Since: v5.0.0


Here is an example bash command using the Veesp provider:

```bash
VEESP_USERNAME="xxxxxxxxxxxxxxxxxxxxx" \
VEESP_PASSWORD="xxxxxxxxxxxxxxxxxxxxx" \
lego run --dns vessp -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `VEESP_PASSWORD` | Password |
| `VEESP_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `VEESP_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `VEESP_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `VEESP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `VEESP_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://secure.veesp.com/userapi#dns-96)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/veesp/veesp.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
