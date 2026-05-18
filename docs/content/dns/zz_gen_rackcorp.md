---
title: "Rackcorp"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: rackcorp
dnsprovider:
  since:    "v5.1.0"
  code:     "rackcorp"
  url:      "https://rackcorp.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rackcorp/rackcorp.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Rackcorp](https://rackcorp.com).


<!--more-->

- Code: `rackcorp`
- Since: v5.1.0


Here is an example bash command using the Rackcorp provider:

```bash
RACKCORP_API_UUID=xxxx \
RACKCORP_API_SECRET=xxxx \
lego run --dns rackcorp -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `RACKCORP_API_SECRET` | API key |
| `RACKCORP_API_UUID` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `RACKCORP_HTTP_TIMEOUT` | API request timeout in seconds (Default: 10) |
| `RACKCORP_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `RACKCORP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `RACKCORP_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

## Description

To authenticate you need to provide a valid API key UUID and Secret.
The API key must have permissions for: dns.domain.getall, dns.domain.get,
  dns.record.create, dns.record.get, dns.record.update, and dns.record.delete.



## More information

- [API documentation](https://github.com/RackCorpCloud/rackcorp-api/wiki/RACKCORP-REST-API)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rackcorp/rackcorp.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
