---
title: "Excedo"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: excedo
dnsprovider:
  since:    "v4.33.0"
  code:     "excedo"
  url:      "https://excedo.se/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/excedo/excedo.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Excedo](https://excedo.se/).


<!--more-->

- Code: `excedo`
- Since: v4.33.0


Here is an example bash command using the Excedo provider:

```bash
EXCEDO_API_KEY=your-api-key \
EXCEDO_API_URL=your-base-url \
lego --dns excedo -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `EXCEDO_API_KEY` | API key |
| `EXCEDO_API_URL` | API base URL |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `EXCEDO_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `EXCEDO_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `EXCEDO_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 300) |
| `EXCEDO_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](none)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/excedo/excedo.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
