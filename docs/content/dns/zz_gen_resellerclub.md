---
title: "ResellerClub"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: resellerclub
dnsprovider:
  since:    "v5.7.0"
  code:     "resellerclub"
  url:      "https://www.resellerclub.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/resellerclub/resellerclub.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [ResellerClub](https://www.resellerclub.com/).


<!--more-->

- Code: `resellerclub`
- Since: v5.7.0


Here is an example bash command using the ResellerClub provider:

```bash
RESELLERCLUB_API_KEY="xxx" \
RESELLERCLUB_USER_ID="yyy" \
lego run --dns resellerclub -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `RESELLERCLUB_API_KEY` | API key |
| `RESELLERCLUB_USER_ID` | User ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `RESELLERCLUB_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `RESELLERCLUB_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `RESELLERCLUB_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `RESELLERCLUB_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://developer.resellerclub.com/documentation)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/resellerclub/resellerclub.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
