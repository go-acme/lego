---
title: "Hostinger"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: hostinger
dnsprovider:
  since:    "v4.27.0"
  code:     "hostinger"
  url:      "https://www.hostinger.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hostinger/hostinger.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Hostinger](https://www.hostinger.com/).


<!--more-->

- Code: `hostinger`
- Since: v4.27.0


Here is an example bash command using the Hostinger provider:

```bash
HOSTINGER_API_TOKEN="xxxxxxxxxxxxxxxxxxxxx" \
lego --dns hostinger -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `HOSTINGER_API_TOKEN` | API Token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `HOSTINGER_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `HOSTINGER_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `HOSTINGER_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `HOSTINGER_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://developers.hostinger.com/#tag/dns-zone)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hostinger/hostinger.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
