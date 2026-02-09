---
title: "Combell"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: combell
dnsprovider:
  since:    "v4.32.0"
  code:     "combell"
  url:      "https://www.combell.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/combell/combell.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Combell](https://www.combell.com/).


<!--more-->

- Code: `combell`
- Since: v4.32.0


Here is an example bash command using the Combell provider:

```bash
COMBELL_API_KEY=xxxxxxxxxxxxxxxxxxxxx \
COMBELL_API_SECRET=yyyyyyyyyyyyyyyyyyyy \
lego --dns combell -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `COMBELL_API_KEY` | The API key |
| `COMBELL_API_SECRET` | The API secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `COMBELL_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `COMBELL_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `COMBELL_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `COMBELL_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 3600) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.combell.com/v2/documentation)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/combell/combell.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
