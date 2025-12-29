---
title: "Lima-City"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: limacity
dnsprovider:
  since:    "v4.18.0"
  code:     "limacity"
  url:      "https://www.lima-city.de"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/limacity/limacity.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Lima-City](https://www.lima-city.de).


<!--more-->

- Code: `limacity`
- Since: v4.18.0


Here is an example bash command using the Lima-City provider:

```bash
LIMACITY_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego --dns limacity -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `LIMACITY_API_KEY` | The API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `LIMACITY_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `LIMACITY_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 80) |
| `LIMACITY_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 480) |
| `LIMACITY_SEQUENCE_INTERVAL` | Time between sequential requests in seconds (Default: 90) |
| `LIMACITY_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.lima-city.de/hilfe/lima-city-api)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/limacity/limacity.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
