---
title: "Regfish"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: regfish
dnsprovider:
  since:    "v4.20.0"
  code:     "regfish"
  url:      "https://regfish.de/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/regfish/regfish.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Regfish](https://regfish.de/).


<!--more-->

- Code: `regfish`
- Since: v4.20.0


Here is an example bash command using the Regfish provider:

```bash
REGFISH_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego --email you@example.com --dns regfish -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `REGFISH_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `REGFISH_HTTP_TIMEOUT` | API request timeout |
| `REGFISH_POLLING_INTERVAL` | Time between DNS propagation check |
| `REGFISH_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `REGFISH_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://regfish.readme.io/)
- [Go client](https://github.com/regfish/regfish-dnsapi-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/regfish/regfish.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
