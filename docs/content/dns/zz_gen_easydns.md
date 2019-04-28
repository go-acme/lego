---
title: "EasyDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: easydns
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/easydns/easydns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v2.6.0

Configuration for [EasyDNS](https://easydns.com/).


<!--more-->

- Code: `easydns`

Here is an example bash command using the EasyDNS provider:

```bash
EASYDNS_TOKEN=<your token> \
EASYDNS_KEY=<your key> \
lego --dns easydns --domains my.domain.com --email my@email.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `EASYDNS_KEY` | API Key |
| `EASYDNS_TOKEN` | API Token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `EASYDNS_ENDPOINT` | The endpoint URL of the API Server |
| `EASYDNS_HTTP_TIMEOUT` | API request timeout |
| `EASYDNS_POLLING_INTERVAL` | Time between DNS propagation check |
| `EASYDNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `EASYDNS_SEQUENCE_INTERVAL` | Time between sequential requests |
| `EASYDNS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).

To test with the sandbox environment set ```EASYDNS_ENDPOINT=https://sandbox.rest.easydns.net```



## More information

- [API documentation](http://docs.sandbox.rest.easydns.net)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/easydns/easydns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
