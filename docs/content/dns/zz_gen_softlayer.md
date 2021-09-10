---
title: "softlayer(IBM Cloud Domain Name Registration)"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: softlayer
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/softlayer/softlayer.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v1.0.0

Configuration for [softlayer(IBM Cloud Domain Name Registration)](https://www.ibm.com/cloud/).


<!--more-->

- Code: `softlayer`

Here is an example bash command using the softlayer(IBM Cloud Domain Name Registration) provider:

```bash
SOFTLAYER_USERNAME=xxxxx \
SOFTLAYER_API_KEY=yyyyy \
lego --email myemail@example.com --dns softlayer --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SOFTLAYER_API_KEY` | Classic Infrastructure API key |
| `SOFTLAYER_USERNAME` | User name(IBM Cloud is <account_ID>_<email_address>) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SOFTLAYER_HTTP_TIMEOUT` | API request timeout |
| `SOFTLAYER_POLLING_INTERVAL` | Time between DNS propagation check |
| `SOFTLAYER_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `SOFTLAYER_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://cloud.ibm.com/docs/dns?topic=dns-getting-started-with-the-dns-api)
- [Go client](https://github.com/softlayer/softlayer-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/softlayer/softlayer.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
