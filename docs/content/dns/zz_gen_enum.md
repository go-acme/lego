---
title: "Enum"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: enum
dnsprovider:
  since:    "v5.6.0"
  code:     "enum"
  url:      "https://enum.co"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/enum/enum.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Enum](https://enum.co).


<!--more-->

- Code: `enum`
- Since: v5.6.0


Here is an example bash command using the Enum provider:

```bash
ENUM_API_TOKEN="xxx" \
lego run --dns enum -d '*.example.com' -d example.com

# or

ENUM_CREDENTIALS="example.com:xxx,example.org:yyy" \
lego run --dns enum -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ENUM_API_TOKEN` | API token. |
| `ENUM_CREDENTIALS` | Only works if API token is not defined. Mapping of zone name and API token. |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ENUM_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `ENUM_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `ENUM_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `ENUM_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.enum.co/api/)
- [Go client](https://github.com/enumco/client-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/enum/enum.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
