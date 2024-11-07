---
title: "Hosttech"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: hosttech
dnsprovider:
  since:    "v4.5.0"
  code:     "hosttech"
  url:      "https://www.hosttech.eu/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hosttech/hosttech.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Hosttech](https://www.hosttech.eu/).


<!--more-->

- Code: `hosttech`
- Since: v4.5.0


Here is an example bash command using the Hosttech provider:

```bash
HOSTTECH_API_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxx \
lego --email you@example.com --dns hosttech -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `HOSTTECH_API_KEY` | API login |
| `HOSTTECH_PASSWORD` | API password |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `HOSTTECH_HTTP_TIMEOUT` | API request timeout |
| `HOSTTECH_POLLING_INTERVAL` | Time between DNS propagation check |
| `HOSTTECH_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `HOSTTECH_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.ns1.hosttech.eu/api/documentation)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hosttech/hosttech.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
