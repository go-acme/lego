---
title: "Bunny"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: bunny
dnsprovider:
  since:    "v4.11.0"
  code:     "bunny"
  url:      "https://bunny.net"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/bunny/bunny.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Bunny](https://bunny.net).


<!--more-->

- Code: `bunny`
- Since: v4.11.0


Here is an example bash command using the Bunny provider:

```bash
BUNNY_API_KEY=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxxxxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
lego --email you@example.com --dns bunny -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `BUNNY_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `BUNNY_POLLING_INTERVAL` | Time between DNS propagation check |
| `BUNNY_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `BUNNY_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.bunny.net/reference/dnszonepublic_index)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/bunny/bunny.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
