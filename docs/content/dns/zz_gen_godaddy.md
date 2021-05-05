---
title: "Go Daddy"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: godaddy
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/godaddy/godaddy.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.5.0

Configuration for [Go Daddy](https://godaddy.com).


<!--more-->

- Code: `godaddy`

Here is an example bash command using the Go Daddy provider:

```bash
GODADDY_API_KEY=xxxxxxxx \
GODADDY_API_SECRET=yyyyyyyy \
lego --email myemail@example.com --dns godaddy --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `GODADDY_API_KEY` | API key |
| `GODADDY_API_SECRET` | API secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GODADDY_HTTP_TIMEOUT` | API request timeout |
| `GODADDY_POLLING_INTERVAL` | Time between DNS propagation check |
| `GODADDY_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `GODADDY_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://developer.godaddy.com/doc/endpoint/domains)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/godaddy/godaddy.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
