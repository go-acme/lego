---
title: "Virtualname"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: virtualname
dnsprovider:
  since:    "v4.30.0"
  code:     "virtualname"
  url:      "https://www.virtualname.es/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/virtualname/virtualname.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Virtualname](https://www.virtualname.es/).


<!--more-->

- Code: `virtualname`
- Since: v4.30.0


Here is an example bash command using the Virtualname provider:

```bash
VIRTUALNAME_TOKEN=xxxxxx \
lego --email you@example.com --dns virtualname -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `VIRTUALNAME_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `VIRTUALNAME_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `VIRTUALNAME_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `VIRTUALNAME_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 300) |
| `VIRTUALNAME_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://developers.virtualname.net/#dns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/virtualname/virtualname.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
