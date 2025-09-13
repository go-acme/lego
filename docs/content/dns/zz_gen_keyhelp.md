---
title: "KeyHelp"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: keyhelp
dnsprovider:
  since:    "v4.26.0"
  code:     "keyhelp"
  url:      "https://www.keyweb.de/en/keyhelp/keyhelp/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/keyhelp/keyhelp.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [KeyHelp](https://www.keyweb.de/en/keyhelp/keyhelp/).


<!--more-->

- Code: `keyhelp`
- Since: v4.26.0


Here is an example bash command using the KeyHelp provider:

```bash
KEYHELP_BASE_URL="https://keyhelp.example.com" \
KEYHELP_API_KEY="xxx" \
lego --email you@example.com --dns keyhelp -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `KEYHELP_API_KEY` | API key |
| `KEYHELP_BASE_URL` | Server URL |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `KEYHELP_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `KEYHELP_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `KEYHELP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `KEYHELP_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://app.swaggerhub.com/apis-docs/keyhelp/api/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/keyhelp/keyhelp.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
