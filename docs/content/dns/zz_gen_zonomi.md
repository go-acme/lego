---
title: "Zonomi"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: zonomi
dnsprovider:
  since:    "v3.5.0"
  code:     "zonomi"
  url:      "https://zonomi.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/zonomi/zonomi.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Zonomi](https://zonomi.com).


<!--more-->

- Code: `zonomi`
- Since: v3.5.0


Here is an example bash command using the Zonomi provider:

```bash
ZONOMI_API_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
lego --dns zonomi -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ZONOMI_API_KEY` | User API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ZONOMI_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `ZONOMI_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `ZONOMI_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `ZONOMI_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 3600) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://zonomi.com/app/dns/dyndns.jsp)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/zonomi/zonomi.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
