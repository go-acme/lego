---
title: "Spaceship"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: spaceship
dnsprovider:
  since:    "v4.22.0"
  code:     "spaceship"
  url:      "https://www.spaceship.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/spaceship/spaceship.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Spaceship](https://www.spaceship.com/).


<!--more-->

- Code: `spaceship`
- Since: v4.22.0


Here is an example bash command using the Spaceship provider:

```bash
SPACESHIP_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
SPACESHIP_API_SECRET="xxxxxxxxxxxxxxxxxxxxx" \
lego --dns spaceship -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SPACESHIP_API_KEY` | API key |
| `SPACESHIP_API_SECRET` | API secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SPACESHIP_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `SPACESHIP_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `SPACESHIP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `SPACESHIP_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.spaceship.dev/#tag/DNS-records)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/spaceship/spaceship.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
