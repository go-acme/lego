---
title: "DNSExit"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dnsexit
dnsprovider:
  since:    "v4.32.0"
  code:     "dnsexit"
  url:      "https://dnsexit.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnsexit/dnsexit.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [DNSExit](https://dnsexit.com).


<!--more-->

- Code: `dnsexit`
- Since: v4.32.0


Here is an example bash command using the DNSExit provider:

```bash
DNSEXIT_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego --dns dnsexit -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DNSEXIT_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DNSEXIT_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `DNSEXIT_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `DNSEXIT_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 300) |
| `DNSEXIT_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://dnsexit.com/dns/dns-api/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnsexit/dnsexit.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
