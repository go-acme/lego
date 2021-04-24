---
title: "Sonic"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: sonic
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/sonic/sonic.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v4.3.1
Support DNS provided by sonic.net


<!--more-->

- Code: `sonic`

Here is an example bash command using the Sonic provider:

```bash
SONIC_USERID=12345 \
SONIC_APIKEY=4d6fbf2f9ab0fa11697470918d37625851fc0c51 \
lego --email myemail@example.com --dns sonic --domains my.example.org run

```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SONIC_APIKEY` | API Key |
| `SONIC_USERID` | API USERID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SONIC_HTTP_TIMEOUT` | API request timeout |
| `SONIC_POLLING_INTERVAL` | Time between DNS propagation check |
| `SONIC_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `SONIC_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).

## Description

You must use `SONIC_USERID` and `SONIC_APIKEY` to authenticate.

### API keys

The API keys (`SONIC_USERID` and `SONIC_APIKEY`), are generated based on an authenticated request to dyndns/api_key.

See https://public-api.sonic.net/dyndns/#requesting_an_api_key for additional details.

This UserID and APIKey combo allow modifications to any DNS entries connected to the managed domain (hostname).

Hostname should be the toplevel domain managed e.g example.com not www.example.com

example:
curl -X POST -H "Content-Type: application/json" --data '{"username":"notarealuser","password":"notarealpassword","hostname":"example.com"}' https://public-api.sonic.net/dyndns/api_key

{"userid":"12345","apikey":"4d6fbf2f9ab0fa11697470918d37625851fc0c51","result":200,"message":"OK"}



## More information

- [API documentation](https://public-api.sonic.net/dyndns/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/sonic/sonic.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
