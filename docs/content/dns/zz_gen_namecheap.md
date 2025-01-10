---
title: "Namecheap"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: namecheap
dnsprovider:
  since:    "v0.3.0"
  code:     "namecheap"
  url:      "https://www.namecheap.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/namecheap/namecheap.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Namecheap](https://www.namecheap.com).

**To enable API access on the Namecheap production environment, some opaque requirements must be met.**
More information in the section [Enabling API Access](https://www.namecheap.com/support/api/intro/) of the Namecheap documentation.
(2020-08: Account balance of $50+, 20+ domains in your account, or purchases totaling $50+ within the last 2 years.)



<!--more-->

- Code: `namecheap`
- Since: v0.3.0


Here is an example bash command using the Namecheap provider:

```bash
NAMECHEAP_API_USER=user \
NAMECHEAP_API_KEY=key \
lego --email you@example.com --dns namecheap -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NAMECHEAP_API_KEY` | API key |
| `NAMECHEAP_API_USER` | API user |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NAMECHEAP_HTTP_TIMEOUT` | API request timeout in seconds (Default: 60) |
| `NAMECHEAP_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 15) |
| `NAMECHEAP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 3600) |
| `NAMECHEAP_SANDBOX` | Activate the sandbox (boolean) |
| `NAMECHEAP_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.namecheap.com/support/api/methods.aspx)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/namecheap/namecheap.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
