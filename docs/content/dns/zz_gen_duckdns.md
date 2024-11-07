---
title: "Duck DNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: duckdns
dnsprovider:
  since:    "v0.5.0"
  code:     "duckdns"
  url:      "https://www.duckdns.org/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/duckdns/duckdns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Duck DNS](https://www.duckdns.org/).


<!--more-->

- Code: `duckdns`
- Since: v0.5.0


Here is an example bash command using the Duck DNS provider:

```bash
DUCKDNS_TOKEN=xxxxxx \
lego --email you@example.com --dns duckdns -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DUCKDNS_TOKEN` | Account token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DUCKDNS_HTTP_TIMEOUT` | API request timeout |
| `DUCKDNS_POLLING_INTERVAL` | Time between DNS propagation check |
| `DUCKDNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DUCKDNS_SEQUENCE_INTERVAL` | Time between sequential requests |
| `DUCKDNS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.duckdns.org/spec.jsp)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/duckdns/duckdns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
