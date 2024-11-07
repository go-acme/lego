---
title: "http.net"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: httpnet
dnsprovider:
  since:    "v4.15.0"
  code:     "httpnet"
  url:      "https://www.http.net/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/httpnet/httpnet.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [http.net](https://www.http.net/).


<!--more-->

- Code: `httpnet`
- Since: v4.15.0


Here is an example bash command using the http.net provider:

```bash
HTTPNET_API_KEY=xxxxxxxx \
lego --email you@example.com --dns httpnet -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `HTTPNET_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `HTTPNET_HTTP_TIMEOUT` | API request timeout |
| `HTTPNET_POLLING_INTERVAL` | Time between DNS propagation check |
| `HTTPNET_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `HTTPNET_TTL` | The TTL of the TXT record used for the DNS challenge |
| `HTTPNET_ZONE_NAME` | Zone name in ACE format |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.http.net/docs/api/#dns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/httpnet/httpnet.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
