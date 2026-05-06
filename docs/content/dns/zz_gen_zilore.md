---
title: "Zilore"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: zilore
dnsprovider:
  since:    "v5.0.0"
  code:     "zilore"
  url:      "https://zilore.com/en"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/zilore/zilore.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Zilore](https://zilore.com/en).


<!--more-->

- Code: `zilore`
- Since: v5.0.0


Here is an example bash command using the Zilore provider:

```bash
ZILORE_ACCESS_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego run --dns zilore -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ZILORE_ACCESS_KEY` | Access key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ZILORE_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `ZILORE_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `ZILORE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `ZILORE_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://zilore.com/en/help/api)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/zilore/zilore.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
