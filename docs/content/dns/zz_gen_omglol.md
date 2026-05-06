---
title: "omg.lol"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: omglol
dnsprovider:
  since:    "v5.0.0"
  code:     "omglol"
  url:      "https://home.omg.lol/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/omglol/omglol.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [omg.lol](https://home.omg.lol/).


<!--more-->

- Code: `omglol`
- Since: v5.0.0


Here is an example bash command using the omg.lol provider:

```bash
OMGLOL_API_KEY="xx" \
lego run --dns omglol -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `OMGLOL_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `OMGLOL_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `OMGLOL_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `OMGLOL_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `OMGLOL_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.omg.lol/#dns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/omglol/omglol.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
