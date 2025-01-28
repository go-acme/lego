---
title: "myaddr.{tools,dev,io}"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: myaddr
dnsprovider:
  since:    "v4.22.0"
  code:     "myaddr"
  url:      "https://myaddr.tools/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/myaddr/myaddr.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [myaddr.{tools,dev,io}](https://myaddr.tools/).


<!--more-->

- Code: `myaddr`
- Since: v4.22.0


Here is an example bash command using the myaddr.{tools,dev,io} provider:

```bash
MYADDR_PRIVATE_KEYS_MAPPING="example:123,test:456" \
lego --email you@example.com --dns myaddr -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `MYADDR_PRIVATE_KEYS_MAPPING` | Mapping between subdomains and private keys. The format is: `<subdomain1>:<private_key1>,<subdomain2>:<private_key2>,<subdomain3>:<private_key3>` |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `MYADDR_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `MYADDR_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `MYADDR_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `MYADDR_SEQUENCE_INTERVAL` | Time between sequential requests in seconds (Default: 2) |
| `MYADDR_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://myaddr.tools/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/myaddr/myaddr.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
