---
title: "Domain Chief"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: domainchief
dnsprovider:
  since:    "v5.7.0"
  code:     "domainchief"
  url:      "https://domain.chief.app/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/domainchief/domainchief.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Domain Chief](https://domain.chief.app/).


<!--more-->

- Code: `domainchief`
- Since: v5.7.0


Here is an example bash command using the Domain Chief provider:

```bash
DOMAINCHIEF_API_TOKEN="xxxxxxxxxxxxxxxxxxxxx" \
lego run --dns domainchief -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DOMAINCHIEF_API_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DOMAINCHIEF_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `DOMAINCHIEF_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `DOMAINCHIEF_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `DOMAINCHIEF_TEAM_SLUG` | Team slug |
| `DOMAINCHIEF_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.chief.tools/api/domainchief)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/domainchief/domainchief.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
