---
title: "Websupport"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: websupport
dnsprovider:
  since:    "v4.10.0"
  code:     "websupport"
  url:      "https://websupport.sk"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/websupport/websupport.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Websupport](https://websupport.sk).


<!--more-->

- Code: `websupport`
- Since: v4.10.0


Here is an example bash command using the Websupport provider:

```bash
WEBSUPPORT_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
WEBSUPPORT_SECRET="yyyyyyyyyyyyyyyyyyyyy" \
lego --email you@example.com --dns websupport -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `WEBSUPPORT_API_KEY` | API key |
| `WEBSUPPORT_SECRET` | API secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `WEBSUPPORT_HTTP_TIMEOUT` | API request timeout |
| `WEBSUPPORT_POLLING_INTERVAL` | Time between DNS propagation check |
| `WEBSUPPORT_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `WEBSUPPORT_SEQUENCE_INTERVAL` | Time between sequential requests |
| `WEBSUPPORT_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://rest.websupport.sk/docs/v1.zone)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/websupport/websupport.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
