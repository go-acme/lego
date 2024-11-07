---
title: "Name.com"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: namedotcom
dnsprovider:
  since:    "v0.5.0"
  code:     "namedotcom"
  url:      "https://www.name.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/namedotcom/namedotcom.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Name.com](https://www.name.com).


<!--more-->

- Code: `namedotcom`
- Since: v0.5.0


Here is an example bash command using the Name.com provider:

```bash
NAMECOM_USERNAME=foo.bar \
NAMECOM_API_TOKEN=a379a6f6eeafb9a55e378c118034e2751e682fab \
lego --email you@example.com --dns namedotcom -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NAMECOM_API_TOKEN` | API token |
| `NAMECOM_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NAMECOM_HTTP_TIMEOUT` | API request timeout |
| `NAMECOM_POLLING_INTERVAL` | Time between DNS propagation check |
| `NAMECOM_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `NAMECOM_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.name.com/api-docs/DNS)
- [Go client](https://github.com/namedotcom/go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/namedotcom/namedotcom.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
