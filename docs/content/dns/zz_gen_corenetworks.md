---
title: "Core-Networks"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: corenetworks
dnsprovider:
  since:    "v4.20.0"
  code:     "corenetworks"
  url:      "https://www.core-networks.de/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/corenetworks/corenetworks.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Core-Networks](https://www.core-networks.de/).


<!--more-->

- Code: `corenetworks`
- Since: v4.20.0


Here is an example bash command using the Core-Networks provider:

```bash
CORENETWORKS_LOGIN="xxxx" \
CORENETWORKS_PASSWORD="yyyy" \
lego --email you@example.com --dns corenetworks -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `CORENETWORKS_LOGIN` | The username of the API account |
| `CORENETWORKS_PASSWORD` | The password |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `CORENETWORKS_HTTP_TIMEOUT` | API request timeout |
| `CORENETWORKS_POLLING_INTERVAL` | Time between DNS propagation check |
| `CORENETWORKS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `CORENETWORKS_SEQUENCE_INTERVAL` | Time between sequential requests |
| `CORENETWORKS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://beta.api.core-networks.de/doc/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/corenetworks/corenetworks.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
