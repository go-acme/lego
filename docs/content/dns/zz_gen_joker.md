---
title: "Joker"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: joker
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/joker/joker.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v2.6.0

Configuration for [Joker](https://joker.com).


<!--more-->

- Code: `joker`

Here is an example bash command using the Joker provider:

```bash
JOKER_USERNAME=<your email> \
JOKER_PASSWORD=<your password> \
lego --dns joker --domains my.domain.com --email my@email.com run
# or
JOKER_API_KEY=<your API key> \
lego --dns joker --domains my.domain.com --email my@email.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `JOKER_API_KEY` | API key |
| `JOKER_PASSWORD` | Joker.com password |
| `JOKER_USERNAME` | Joker.com username (email address) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `JOKER_HTTP_TIMEOUT` | API request timeout |
| `JOKER_POLLING_INTERVAL` | Time between DNS propagation check |
| `JOKER_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `JOKER_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://joker.com/faq/category/39/22-dmapi.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/joker/joker.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
