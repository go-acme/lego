---
title: "Axelname"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: axelname
dnsprovider:
  since:    "v4.23.0"
  code:     "axelname"
  url:      "https://axelname.ru"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/axelname/axelname.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Axelname](https://axelname.ru).


<!--more-->

- Code: `axelname`
- Since: v4.23.0


Here is an example bash command using the Axelname provider:

```bash
AXELNAME_NICKNAME="yyy" \
AXELNAME_TOKEN="xxx" \
lego --email you@example.com --dns axelname -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `AXELNAME_NICKNAME` | Account nickname |
| `AXELNAME_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `AXELNAME_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `AXELNAME_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `AXELNAME_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `AXELNAME_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://axelname.ru/static/content/files/axelname_api_rest_lite.pdf)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/axelname/axelname.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
