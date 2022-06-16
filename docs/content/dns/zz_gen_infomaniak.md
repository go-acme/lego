---
title: "Infomaniak"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: infomaniak
dnsprovider:
  since:    "v4.1.0"
  code:     "infomaniak"
  url:      "https://www.infomaniak.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/infomaniak/infomaniak.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Infomaniak](https://www.infomaniak.com/).


<!--more-->

- Code: `infomaniak`
- Since: v4.1.0


Here is an example bash command using the Infomaniak provider:

```bash
INFOMANIAK_ACCESS_TOKEN=1234567898765432 \
lego --email you@example.com --dns infomaniak --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `INFOMANIAK_ACCESS_TOKEN` | Access token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `INFOMANIAK_ENDPOINT` | https://api.infomaniak.com |
| `INFOMANIAK_HTTP_TIMEOUT` | API request timeout |
| `INFOMANIAK_POLLING_INTERVAL` | Time between DNS propagation check |
| `INFOMANIAK_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `INFOMANIAK_TTL` | The TTL of the TXT record used for the DNS challenge in seconds |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).

## Access token

Access token can be created at the url https://manager.infomaniak.com/v3/infomaniak-api.
You will need domain scope.



## More information

- [API documentation](https://api.infomaniak.com/doc)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/infomaniak/infomaniak.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
