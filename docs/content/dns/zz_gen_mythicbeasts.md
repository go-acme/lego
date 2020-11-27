---
title: "MythicBeasts"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: mythicbeasts
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/mythicbeasts/mythicbeasts.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.3.7

Configuration for [MythicBeasts](https://www.mythic-beasts.com/).


<!--more-->

- Code: `mythicbeasts`

Here is an example bash command using the MythicBeasts provider:

```bash
MYTHICBEASTS_USER_NAME=myuser \
MYTHICBEASTS_PASSWORD=mypass \
lego --email myemail@example.com --dns mythicbeasts --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `MYTHICBEASTS_PASSWORD` | Password |
| `MYTHICBEASTS_USERNAME` | User name |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `MYTHICBEASTS_API_ENDPOINT` | The endpoint for the API (must implement v2) |
| `MYTHICBEASTS_HTTP_TIMEOUT` | API request timeout |
| `MYTHICBEASTS_POLLING_INTERVAL` | Time between DNS propagation check |
| `MYTHICBEASTS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `MYTHICBEASTS_TTL` | The TTL of the TXT record used for the DNS challenge |
| `MYTHICBEASYS_AUTH_API_ENDPOINT` | The endpoint for Mythic Beasts' Authentication |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).

If you are using specific API keys, then the username is the API ID for your API key, and the password is the API secret.

Your API key name is not needed to operate lego.



## More information

- [API documentation](https://www.mythic-beasts.com/support/api/dnsv2)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/mythicbeasts/mythicbeasts.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
