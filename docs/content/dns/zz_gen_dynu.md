---
title: "Dynu"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dynu
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dynu/dynu.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v3.5.0

Configuration for [Dynu](https://www.dynu.com/).


<!--more-->

- Code: `dynu`

Here is an example bash command using the Dynu provider:

```bash
DYNU_API_KEY=1234567890abcdefghijklmnopqrstuvwxyz \
lego --email myemail@example.com --dns dynu --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DYNU_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DYNU_HTTP_TIMEOUT` | API request timeout |
| `DYNU_POLLING_INTERVAL` | Time between DNS propagation check |
| `DYNU_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DYNU_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://www.dynu.com/en-US/Support/API)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dynu/dynu.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
