---
title: "Infoblox"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: infoblox
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/infoblox/infoblox.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v4.3.1

Configuration for [Infoblox](https://www.infoblox.com/).


<!--more-->

- Code: `infoblox`

Here is an example bash command using the Infoblox provider:

```bash
INFOBLOX_USER=api-user-529 \
INFOBLOX_PASSWORD=b9841238feb177a84330febba8a83208921177bffe733 \
INFOBLOX_HOST=infoblox.example.org
lego --email myemail@example.com --dns infoblox --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `INFOBLOX_HOST` | Infoblox Host URI |
| `INFOBLOX_PASSWORD` | Infoblox Account Password |
| `INFOBLOX_USER` | Infoblox Account Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `INFOBLOX_PORT` | The port for the infoblox grid manager, default: 443 |
| `INFOBLOX_TTL` | The TTL of the TXT record used for the DNS challenge |
| `INFOBLOX_VIEW` | The view for the TXT records, default: External |
| `INFOBLOX_WAPI_VERSION` | The version of wapi being used, default: 2.11 |
| `SSL_VERIFY` | Whether or not to verify the TLS certificate, default: true |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).

When creating an api user ensure it has the proper permissions for the view you are working with. 



## More information

- [API documentation](https://your.infoblox.server/wapidoc/)
- [Go client](https://github.com/infobloxopen/infoblox-go-client)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/infoblox/infoblox.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
