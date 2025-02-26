---
title: "Infoblox"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: infoblox
dnsprovider:
  since:    "v4.4.0"
  code:     "infoblox"
  url:      "https://www.infoblox.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/infoblox/infoblox.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Infoblox](https://www.infoblox.com/).


<!--more-->

- Code: `infoblox`
- Since: v4.4.0


Here is an example bash command using the Infoblox provider:

```bash
INFOBLOX_USERNAME=api-user-529 \
INFOBLOX_PASSWORD=b9841238feb177a84330febba8a83208921177bffe733 \
INFOBLOX_HOST=infoblox.example.org
lego --email you@example.com --dns infoblox -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `INFOBLOX_HOST` | Host URI |
| `INFOBLOX_PASSWORD` | Account Password |
| `INFOBLOX_USERNAME` | Account Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `INFOBLOX_CA_CERTIFICATE` | The path to the CA certificate (PEM encoded) |
| `INFOBLOX_DNS_VIEW` | The view for the TXT records (Default: External) |
| `INFOBLOX_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `INFOBLOX_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `INFOBLOX_PORT` | The port for the infoblox grid manager  (Default: 443) |
| `INFOBLOX_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `INFOBLOX_SSL_VERIFY` | Whether or not to verify the TLS certificate  (Default: true) |
| `INFOBLOX_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |
| `INFOBLOX_WAPI_VERSION` | The version of WAPI being used  (Default: 2.11) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

When creating an API's user ensure it has the proper permissions for the view you are working with.



## More information

- [API documentation](https://your.infoblox.server/wapidoc/)
- [Go client](https://github.com/infobloxopen/infoblox-go-client)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/infoblox/infoblox.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
