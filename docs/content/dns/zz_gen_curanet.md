---
title: "Curanet"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: curanet
dnsprovider:
  since:    "v4.32.0"
  code:     "curanet"
  url:      "https://curanet.dk/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/curanet/curanet.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Curanet](https://curanet.dk/).


<!--more-->

- Code: `curanet`
- Since: v4.32.0


Here is an example bash command using the Curanet provider:

```bash
CURANET_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego --dns curanet -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `CURANET_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `CURANET_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `CURANET_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `CURANET_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `CURANET_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.curanet.dk/dns/swagger/index.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/curanet/curanet.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
