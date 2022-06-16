---
title: "NS1"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: ns1
dnsprovider:
  since:    "v0.4.0"
  code:     "ns1"
  url:      "https://ns1.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ns1/ns1.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [NS1](https://ns1.com).


<!--more-->

- Code: `ns1`
- Since: v0.4.0


Here is an example bash command using the NS1 provider:

```bash
NS1_API_KEY=xxxx \
lego --email you@example.com --dns ns1 --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NS1_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NS1_HTTP_TIMEOUT` | API request timeout |
| `NS1_POLLING_INTERVAL` | Time between DNS propagation check |
| `NS1_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `NS1_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://ns1.com/api)
- [Go client](https://github.com/ns1/ns1-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ns1/ns1.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
