---
title: "Digital Ocean"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: digitalocean
dnsprovider:
  since:    "v0.3.0"
  code:     "digitalocean"
  url:      "https://www.digitalocean.com/docs/networking/dns/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/digitalocean/digitalocean.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Digital Ocean](https://www.digitalocean.com/docs/networking/dns/).


<!--more-->

- Code: `digitalocean`
- Since: v0.3.0


Here is an example bash command using the Digital Ocean provider:

```bash
DO_AUTH_TOKEN=xxxxxx \
lego --email you@example.com --dns digitalocean --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DO_AUTH_TOKEN` | Authentication token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DO_API_URL` | The URL of the API |
| `DO_HTTP_TIMEOUT` | API request timeout |
| `DO_POLLING_INTERVAL` | Time between DNS propagation check |
| `DO_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DO_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://developers.digitalocean.com/documentation/v2/#domain-records)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/digitalocean/digitalocean.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
