---
title: "Variomedia"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: variomedia
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/variomedia/variomedia.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v4.8.0

Configuration for [Variomedia](https://www.variomedia.de/).


<!--more-->

- Code: `variomedia`

Here is an example bash command using the Variomedia provider:

```bash
VARIOMEDIA_API_TOKEN=xxxx \
lego --email myemail@example.com --dns variomedia --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `VARIOMEDIA_API_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DODE_SEQUENCE_INTERVAL` | Time between sequential requests |
| `VARIOMEDIA_HTTP_TIMEOUT` | API request timeout |
| `VARIOMEDIA_POLLING_INTERVAL` | Time between DNS propagation check |
| `VARIOMEDIA_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `VARIOMEDIA_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://api.variomedia.de/docs/dns-records.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/variomedia/variomedia.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
