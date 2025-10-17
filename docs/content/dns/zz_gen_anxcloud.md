---
title: "Anexia CloudDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: anxcloud
dnsprovider:
  since:    "v4.21.0"
  code:     "anxcloud"
  url:      "https://www.anexia-it.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/anxcloud/anxcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Anexia CloudDNS](https://www.anexia-it.com/).


<!--more-->

- Code: `anxcloud`
- Since: v4.21.0


Here is an example bash command using the Anexia CloudDNS provider:

```bash
ANEXIA_TOKEN=your-api-token \
lego --email you@example.com --dns anxcloud -d '*.example.com' -d example.com run
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
| `ANEXIA_API_URL` | Custom API endpoint URL (optional) |
| `ANEXIA_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `ANEXIA_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `ANEXIA_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 300) |
| `ANEXIA_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

## Description

You need to create an API token in the [Anexia Engine](https://engine.anexia-it.com/).

The token must have permissions to manage DNS zones and records.

## API Token

Create an API token through the Anexia Engine web interface or API.
Pass the token to Lego using the `ANEXIA_TOKEN` environment variable.



## More information

- [API documentation](https://engine.anexia-it.com/docs/)
- [Go client](https://github.com/anexia-it/go-anxcloud)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/anxcloud/anxcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
