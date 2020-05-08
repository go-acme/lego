---
title: "Hetzner"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: hetzner
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hetzner/hetzner.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v3.7.0

Configuration for [Hetzner](https://hetzner.com).


<!--more-->

- Code: `hetzner`

Here is an example bash command using the Hetzner provider:

```bash
HETZNER_API_KEY=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
lego --dns hetzner --domains my.domain.com --email my@email.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `HETZNER_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `HETZNER_HTTP_TIMEOUT` | API request timeout |
| `HETZNER_POLLING_INTERVAL` | Time between DNS propagation check |
| `HETZNER_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `HETZNER_SEQUENCE_INTERVAL` | Interval between iteration |
| `HETZNER_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://dns.hetzner.com/api-docs)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hetzner/hetzner.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
