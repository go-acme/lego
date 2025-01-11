---
title: "deSEC.io"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: desec
dnsprovider:
  since:    "v3.7.0"
  code:     "desec"
  url:      "https://desec.io"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/desec/desec.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [deSEC.io](https://desec.io).


<!--more-->

- Code: `desec`
- Since: v3.7.0


Here is an example bash command using the deSEC.io provider:

```bash
DESEC_TOKEN=x-xxxxxxxxxxxxxxxxxxxxxxxxxx \
lego --email you@example.com --dns desec -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DESEC_TOKEN` | Domain token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DESEC_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `DESEC_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 4) |
| `DESEC_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `DESEC_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 3600) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://desec.readthedocs.io/en/latest/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/desec/desec.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
