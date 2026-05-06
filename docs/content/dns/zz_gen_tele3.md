---
title: "Tele3"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: tele3
dnsprovider:
  since:    "v5.0.0"
  code:     "tele3"
  url:      "https://www.tele3.cz"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/tele3/tele3.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Tele3](https://www.tele3.cz).


<!--more-->

- Code: `tele3`
- Since: v5.0.0


Here is an example bash command using the Tele3 provider:

```bash
TELE3_KEY="xxx" \
TELE3_SECRET="yyy" \
lego run --dns tele3 -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `TELE3_KEY` | Key |
| `TELE3_SECRET` | Secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `TELE3_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `TELE3_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `TELE3_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `TELE3_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.tele3.cz/system-acme-api.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/tele3/tele3.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
