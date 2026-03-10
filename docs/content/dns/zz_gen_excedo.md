---
title: "Excedo DNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: excedo
dnsprovider:
  since:    "v4.0.0"
  code:     "excedo"
  url:      "your-base-url"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/excedo/excedo.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Excedo DNS](your-base-url).


<!--more-->

- Code: `excedo`
- Since: v4.0.0


Here is an example bash command using the Excedo DNS provider:

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
| `EXCEDO_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `EXCEDO_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 300) |
| `EXCEDO_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

## Description

The provider logs in with the API key (`GET /authenticate/login/`) and uses the
returned token for subsequent DNS record requests.



## More information



<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/excedo/excedo.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
