---
title: "Domain Offensive (do.de)"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dode
dnsprovider:
  since:    "v2.4.0"
  code:     "dode"
  url:      "https://www.do.de/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dode/dode.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Domain Offensive (do.de)](https://www.do.de/).


<!--more-->

- Code: `dode`
- Since: v2.4.0


Here is an example bash command using the Domain Offensive (do.de) provider:

```bash
DODE_TOKEN=xxxxxx \
lego --email you@example.com --dns dode --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DODE_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DODE_HTTP_TIMEOUT` | API request timeout |
| `DODE_POLLING_INTERVAL` | Time between DNS propagation check |
| `DODE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DODE_SEQUENCE_INTERVAL` | Time between sequential requests |
| `DODE_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.do.de/wiki/freie-ssl-tls-zertifikate-ueber-acme/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dode/dode.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
