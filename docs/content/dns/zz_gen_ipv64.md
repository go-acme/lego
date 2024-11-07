---
title: "IPv64"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: ipv64
dnsprovider:
  since:    "v4.13.0"
  code:     "ipv64"
  url:      "https://ipv64.net/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ipv64/ipv64.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [IPv64](https://ipv64.net/).


<!--more-->

- Code: `ipv64`
- Since: v4.13.0


Here is an example bash command using the IPv64 provider:

```bash
IPV64_API_KEY=xxxxxx \
lego --email you@example.com --dns ipv64 -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `IPV64_API_KEY` | Account API Key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `IPV64_HTTP_TIMEOUT` | API request timeout |
| `IPV64_POLLING_INTERVAL` | Time between DNS propagation check |
| `IPV64_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `IPV64_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://ipv64.net/dyndns_updater_api)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ipv64/ipv64.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
