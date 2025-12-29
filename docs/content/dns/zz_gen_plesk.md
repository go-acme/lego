---
title: "plesk.com"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: plesk
dnsprovider:
  since:    "v4.11.0"
  code:     "plesk"
  url:      "https://www.plesk.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/plesk/plesk.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [plesk.com](https://www.plesk.com/).


<!--more-->

- Code: `plesk`
- Since: v4.11.0


Here is an example bash command using the plesk.com provider:

```bash
PLESK_SERVER_BASE_URL="https://plesk.myserver.com:8443" \
PLESK_USERNAME=xxxxxx \
PLESK_PASSWORD=yyyyyy \
lego --dns plesk -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `PLESK_PASSWORD` | API password |
| `PLESK_SERVER_BASE_URL` | Base URL of the server (ex: https://plesk.myserver.com:8443) |
| `PLESK_USERNAME` | API username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `PLESK_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `PLESK_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `PLESK_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `PLESK_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.plesk.com/en-US/obsidian/api-rpc/about-xml-api/reference.28784/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/plesk/plesk.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
