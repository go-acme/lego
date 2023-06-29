---
title: "RcodeZero"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: rcodezero
dnsprovider:
  since:    "v4.12.4"
  code:     "rcodezero"
  url:      "https://www.rcodezero.at/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rcodezero/rcodezero.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [RcodeZero](https://www.rcodezero.at/).


<!--more-->

- Code: `rcodezero`
- Since: v4.12.4


Here is an example bash command using the RcodeZero provider:

```bash
RCODEZERO_API_TOKEN=<mytoken> \
lego --email you@example.com --dns pdns --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `RCODEZERO_API_TOKEN` | API token |
| `RCODEZERO_API_URL` | alternative API URL, default: https://my.rcodezero.at/api/v1/acme |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `RCODEZERO_HTTP_TIMEOUT` | API request timeout |
| `RCODEZERO_POLLING_INTERVAL` | Time between DNS propagation check |
| `RCODEZERO_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `RCODEZERO_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://my.rcodezero.at/openapi)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rcodezero/rcodezero.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
