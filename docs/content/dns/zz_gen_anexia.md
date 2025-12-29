---
title: "Anexia CloudDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: anexia
dnsprovider:
  since:    "v4.28.0"
  code:     "anexia"
  url:      "https://www.anexia-it.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/anexia/anexia.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Anexia CloudDNS](https://www.anexia-it.com/).


<!--more-->

- Code: `anexia`
- Since: v4.28.0


Here is an example bash command using the Anexia CloudDNS provider:

```bash
ANEXIA_TOKEN=xxx \
lego --dns anexia -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ANEXIA_TOKEN` | API token for Anexia Engine |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ANEXIA_API_URL` | API endpoint URL (default: https://engine.anexia-it.com) |
| `ANEXIA_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `ANEXIA_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `ANEXIA_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 300) |
| `ANEXIA_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

## Description

You need to create an API token in the [Anexia Engine](https://engine.anexia-it.com/).

The token must have permissions to manage DNS zones and records.



## More information

- [API documentation](https://engine.anexia-it.com/docs/en/module/clouddns/api)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/anexia/anexia.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
