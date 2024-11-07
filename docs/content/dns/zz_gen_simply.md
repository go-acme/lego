---
title: "Simply.com"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: simply
dnsprovider:
  since:    "v4.4.0"
  code:     "simply"
  url:      "https://www.simply.com/en/domains/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/simply/simply.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Simply.com](https://www.simply.com/en/domains/).


<!--more-->

- Code: `simply`
- Since: v4.4.0


Here is an example bash command using the Simply.com provider:

```bash
SIMPLY_ACCOUNT_NAME=xxxxxx \
SIMPLY_API_KEY=yyyyyy \
lego --email you@example.com --dns simply -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SIMPLY_ACCOUNT_NAME` | Account name |
| `SIMPLY_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SIMPLY_HTTP_TIMEOUT` | API request timeout |
| `SIMPLY_POLLING_INTERVAL` | Time between DNS propagation check |
| `SIMPLY_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `SIMPLY_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.simply.com/en/docs/api/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/simply/simply.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
