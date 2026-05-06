---
title: "HostUp"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: hostup
dnsprovider:
  since:    "v4.36.0"
  code:     "hostup"
  url:      "https://hostup.se/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hostup/hostup.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [HostUp](https://hostup.se/).


<!--more-->

- Code: `hostup`
- Since: v4.36.0


Here is an example bash command using the HostUp provider:

```bash
HOSTUP_API_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxx \
lego --dns hostup -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `HOSTUP_API_KEY` | API key with read:dns, write:dns, and read:domains scopes |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `HOSTUP_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `HOSTUP_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `HOSTUP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `HOSTUP_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://hostup.se/en/support/api-autentisering/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hostup/hostup.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
