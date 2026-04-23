---
title: "51DNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dns51
dnsprovider:
  since:    "v5.0.0"
  code:     "dns51"
  url:      "https://www.51dns.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dns51/dns51.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [51DNS](https://www.51dns.com).


<!--more-->

- Code: `dns51`
- Since: v5.0.0


Here is an example bash command using the 51DNS provider:

```bash
DNS51_API_KEY="xxx" \
DNS51_API_SECRET="yyy" \
lego --dns dns51 -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DNS51_API_KEY` | API key |
| `DNS51_API_SECRET` | API secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DNS51_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `DNS51_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `DNS51_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `DNS51_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.51dns.com/document/api/4/81.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dns51/dns51.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
