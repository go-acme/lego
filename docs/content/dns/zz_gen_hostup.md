---
title: "HostUp"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: hostup
dnsprovider:
  since:    "v5.0.0"
  code:     "hostup"
  url:      "https://hostup.se/en/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hostup/hostup.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [HostUp](https://hostup.se/en/).


<!--more-->

- Code: `hostup`
- Since: v5.0.0


Here is an example bash command using the HostUp provider:

```bash
HOSTUP_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego run --dns hostup -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `HOSTUP_API_KEY` | API token (required scopes: read:dns, write:dns, read:domains) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `HOSTUP_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `HOSTUP_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `HOSTUP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `HOSTUP_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://developer.hostup.se/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hostup/hostup.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
