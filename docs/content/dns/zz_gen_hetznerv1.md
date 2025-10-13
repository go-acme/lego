---
title: "Hetzner"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: hetznerv1
dnsprovider:
  since:    "v4.27.0"
  code:     "hetznerv1"
  url:      "https://hetzner.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hetznerv1/hetznerv1.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Hetzner](https://hetzner.com).


<!--more-->

- Code: `hetznerv1`
- Since: v4.27.0


Here is an example bash command using the Hetzner provider:

```bash
HETZNER_API_TOKEN="xxxxxxxxxxxxxxxxxxxxx" \
lego --email you@example.com --dns hetznerv1 -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `HETZNER_API_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `HETZNER_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `HETZNER_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `HETZNER_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `HETZNER_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.hetzner.cloud/reference/cloud#dns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hetznerv1/hetznerv1.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
