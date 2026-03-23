---
title: "1cloud.ru"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: onecloudru
dnsprovider:
  since:    "v4.34.0"
  code:     "onecloudru"
  url:      "https://1cloud.ru/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/onecloudru/onecloudru.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [1cloud.ru](https://1cloud.ru/).


<!--more-->

- Code: `onecloudru`
- Since: v4.34.0


Here is an example bash command using the 1cloud.ru provider:

```bash
ONECLOUDRU_TOKEN="xxxxxxxxxxxxxxxxxxxxx" \
lego --dns onecloudru -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ONECLOUDRU_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ONECLOUDRU_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `ONECLOUDRU_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `ONECLOUDRU_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `ONECLOUDRU_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://1cloud.ru/api/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/onecloudru/onecloudru.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
