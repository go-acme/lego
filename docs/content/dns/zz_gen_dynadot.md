---
title: "Dynadot"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dynadot
dnsprovider:
  since:    "v5.1.0"
  code:     "dynadot"
  url:      "https://www.dynadot.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dynadot/dynadot.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Dynadot](https://www.dynadot.com/).


<!--more-->

- Code: `dynadot`
- Since: v5.1.0


Here is an example bash command using the Dynadot provider:

```bash
DYNADOT_API_KEY="xxx" \
DYNADOT_API_SECRET="yyy" \
lego run --dns dynadot -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DYNADOT_API_KEY` | API key |
| `DYNADOT_API_SECRET` | API secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DYNADOT_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `DYNADOT_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `DYNADOT_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `DYNADOT_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.dynadot.com/domain/api-document)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dynadot/dynadot.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
