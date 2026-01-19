---
title: "No-IP"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: noip
dnsprovider:
  since:    "v4.32.0"
  code:     "noip"
  url:      "https://www.noip.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/noip/noip.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [No-IP](https://www.noip.com/).


<!--more-->

- Code: `noip`
- Since: v4.32.0


Here is an example bash command using the No-IP provider:

```bash
NOIP_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego --dns noip -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NOIP_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NOIP_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `NOIP_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `NOIP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `NOIP_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://developer.noip.com/reference/v1-dns-records-list-names)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/noip/noip.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
