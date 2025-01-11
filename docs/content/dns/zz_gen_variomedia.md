---
title: "Variomedia"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: variomedia
dnsprovider:
  since:    "v4.8.0"
  code:     "variomedia"
  url:      "https://www.variomedia.de/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/variomedia/variomedia.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Variomedia](https://www.variomedia.de/).


<!--more-->

- Code: `variomedia`
- Since: v4.8.0


Here is an example bash command using the Variomedia provider:

```bash
VARIOMEDIA_API_TOKEN=xxxx \
lego --email you@example.com --dns variomedia -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `VARIOMEDIA_API_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `VARIOMEDIA_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `VARIOMEDIA_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `VARIOMEDIA_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `VARIOMEDIA_SEQUENCE_INTERVAL` | Time between sequential requests in seconds (Default: 60) |
| `VARIOMEDIA_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.variomedia.de/docs/dns-records.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/variomedia/variomedia.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
