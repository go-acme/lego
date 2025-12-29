---
title: "Hosting.de"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: hostingde
dnsprovider:
  since:    "v1.1.0"
  code:     "hostingde"
  url:      "https://www.hosting.de/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hostingde/hostingde.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Hosting.de](https://www.hosting.de/).


<!--more-->

- Code: `hostingde`
- Since: v1.1.0


Here is an example bash command using the Hosting.de provider:

```bash
HOSTINGDE_API_KEY=xxxxxxxx \
lego --dns hostingde -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `HOSTINGDE_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `HOSTINGDE_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `HOSTINGDE_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `HOSTINGDE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `HOSTINGDE_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |
| `HOSTINGDE_ZONE_NAME` | Zone name in ACE format |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.hosting.de/api/#dns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hostingde/hostingde.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
