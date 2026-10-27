---
title: "CloudPress"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: cloudpress
dnsprovider:
  since:    "v5.6.0"
  code:     "cloudpress"
  url:      "https://cloudpress.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/cloudpress/cloudpress.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [CloudPress](https://cloudpress.com/).


<!--more-->

- Code: `cloudpress`
- Since: v5.6.0


Here is an example bash command using the CloudPress provider:

```bash
CLOUDPRESS_API_KEY="xxx" \
CLOUDPRESS_BASE_URL="https://example.org" \
lego run --dns cloudpress -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `CLOUDPRESS_API_KEY` | API key |
| `CLOUDPRESS_BASE_URL` | API URL |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `CLOUDPRESS_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `CLOUDPRESS_PACCOUNT_ID` | Account ID |
| `CLOUDPRESS_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `CLOUDPRESS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 300) |
| `CLOUDPRESS_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.cloudpress.com/api/dns/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/cloudpress/cloudpress.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
