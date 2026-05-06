---
title: "Online.net"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: onlinenet
dnsprovider:
  since:    "v4.34.0"
  code:     "onlinenet"
  url:      "https://online.net/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/onlinenet/onlinenet.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Online.net](https://online.net/).


<!--more-->

- Code: `onlinenet`
- Since: v4.34.0


Here is an example bash command using the Online.net provider:

```bash
ONLINENET_API_TOKEN="xxx" \
lego run --dns onlinenet -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ONLINENET_API_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ONLINENET_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `ONLINENET_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 15) |
| `ONLINENET_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 240) |
| `ONLINENET_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://console.online.net/en/api/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/onlinenet/onlinenet.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
