---
title: "NameOcean"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: nameocean
dnsprovider:
  since:    "v5.7.0"
  code:     "nameocean"
  url:      "https://nameocean.net/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nameocean/nameocean.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [NameOcean](https://nameocean.net/).


<!--more-->

- Code: `nameocean`
- Since: v5.7.0


Here is an example bash command using the NameOcean provider:

```bash
NAMEOCEAN_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego run --dns nameocean -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NAMEOCEAN_API_KEY` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NAMEOCEAN_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `NAMEOCEAN_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `NAMEOCEAN_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `NAMEOCEAN_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://nameocean.net/developers/api/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nameocean/nameocean.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
