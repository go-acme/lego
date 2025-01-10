---
title: "freemyip.com"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: freemyip
dnsprovider:
  since:    "v4.5.0"
  code:     "freemyip"
  url:      "https://freemyip.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/freemyip/freemyip.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [freemyip.com](https://freemyip.com/).


<!--more-->

- Code: `freemyip`
- Since: v4.5.0


Here is an example bash command using the freemyip.com provider:

```bash
FREEMYIP_TOKEN=xxxxxx \
lego --email you@example.com --dns freemyip -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `FREEMYIP_TOKEN` | Account token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `FREEMYIP_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `FREEMYIP_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `FREEMYIP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `FREEMYIP_SEQUENCE_INTERVAL` | Time between sequential requests in seconds (Default: 60) |
| `FREEMYIP_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 3600) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://freemyip.com/help)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/freemyip/freemyip.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
