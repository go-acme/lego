---
title: "Hosting.nl"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: hostingnl
dnsprovider:
  since:    "v4.21.0"
  code:     "hostingnl"
  url:      "https://hosting.nl"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hostingnl/hostingnl.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Hosting.nl](https://hosting.nl).


<!--more-->

- Code: `hostingnl`
- Since: v4.21.0


Here is an example bash command using the Hosting.nl provider:

```bash
HOSTINGNL_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego --email you@example.com --dns hostingnl -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `HOSTINGNL_API_KEY` | The API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `HOSTINGNL_HTTP_TIMEOUT` | API request timeout |
| `HOSTINGNL_POLLING_INTERVAL` | Time between DNS propagation check |
| `HOSTINGNL_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `HOSTINGNL_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.hosting.nl/api/documentation)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hostingnl/hostingnl.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
