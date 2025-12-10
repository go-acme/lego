---
title: "Neodigit"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: neodigit
dnsprovider:
  since:    "v4.30.0"
  code:     "neodigit"
  url:      "https://www.neodigit.net"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/neodigit/neodigit.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Neodigit](https://www.neodigit.net).


<!--more-->

- Code: `neodigit`
- Since: v4.30.0


Here is an example bash command using the Neodigit provider:

```bash
NEODIGIT_TOKEN=xxxxxx \
lego --email you@example.com --dns neodigit -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NEODIGIT_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NEODIGIT_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `NEODIGIT_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `NEODIGIT_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 300) |
| `NEODIGIT_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://developers.neodigit.net/#dns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/neodigit/neodigit.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
