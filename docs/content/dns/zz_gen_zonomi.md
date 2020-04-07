---
title: "Zonomi"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: zonomi
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/zonomi/zonomi.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.3.5

Configuration for [Zonomi](https://zonomi.com).


<!--more-->

- Code: `zonomi`

Here is an example bash command using the Zonomi provider:

```bash
ZONOMI_API_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
lego --dns zonomi --domains my.domain.com --email my@email.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ZONOMI_API_KEY` | User API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ZONOMI_HTTP_TIMEOUT` | API request timeout |
| `ZONOMI_POLLING_INTERVAL` | Time between DNS propagation check |
| `ZONOMI_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `ZONOMI_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://zonomi.com/app/dns/dyndns.jsp)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/zonomi/zonomi.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
