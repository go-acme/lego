---
title: "mijn.host"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: mijnhost
dnsprovider:
  since:    "v4.18.0"
  code:     "mijnhost"
  url:      "https://mijn.host/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/mijnhost/mijnhost.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [mijn.host](https://mijn.host/).


<!--more-->

- Code: `mijnhost`
- Since: v4.18.0


Here is an example bash command using the mijn.host provider:

```bash
MIJNHOST_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego --email you@example.com --dns mijnhost -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `MIJNHOST_API_KEY` | The API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `MIJNHOST_HTTP_TIMEOUT` | API request timeout |
| `MIJNHOST_POLLING_INTERVAL` | Time between DNS propagation check |
| `MIJNHOST_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `MIJNHOST_SEQUENCE_INTERVAL` | Time between sequential requests |
| `MIJNHOST_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://mijn.host/api/doc/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/mijnhost/mijnhost.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
