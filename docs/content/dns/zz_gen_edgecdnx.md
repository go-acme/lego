---
title: "EdgeCDN-X"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: edgecdnx
dnsprovider:
  since:    "v5.7.0"
  code:     "edgecdnx"
  url:      "https://edgecdnx.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/edgecdnx/edgecdnx.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [EdgeCDN-X](https://edgecdnx.com/).


<!--more-->

- Code: `edgecdnx`
- Since: v5.7.0


Here is an example bash command using the EdgeCDN-X provider:

```bash
EDGECDNX_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
EDGECDNX_PROJECT_ID="project-id" \
lego run --dns edgecdnx -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `EDGECDNX_API_KEY` | Project-scoped API key with Zone read and update permissions |
| `EDGECDNX_PROJECT_ID` | Project resource ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `EDGECDNX_API_URL` | API base URL (Default: https://api.edgecdnx.com) |
| `EDGECDNX_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `EDGECDNX_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `EDGECDNX_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `EDGECDNX_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.edgecdnx.com/swagger/index.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/edgecdnx/edgecdnx.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
