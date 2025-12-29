---
title: "Mittwald"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: mittwald
dnsprovider:
  since:    "v1.48.0"
  code:     "mittwald"
  url:      "https://www.mittwald.de/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/mittwald/mittwald.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Mittwald](https://www.mittwald.de/).


<!--more-->

- Code: `mittwald`
- Since: v1.48.0


Here is an example bash command using the Mittwald provider:

```bash
MITTWALD_TOKEN=my-token \
lego --dns mittwald -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `MITTWALD_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `MITTWALD_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `MITTWALD_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `MITTWALD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `MITTWALD_SEQUENCE_INTERVAL` | Time between sequential requests in seconds (Default: 120) |
| `MITTWALD_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.mittwald.de/v2/docs/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/mittwald/mittwald.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
