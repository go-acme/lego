---
title: "iwantmyname (Deprecated)"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: iwantmyname
dnsprovider:
  since:    "v4.7.0"
  code:     "iwantmyname"
  url:      "https://iwantmyname.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/iwantmyname/iwantmyname.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

The iwantmyname API has shut down.

https://github.com/go-acme/lego/issues/2563



<!--more-->

- Code: `iwantmyname`
- Since: v4.7.0


Here is an example bash command using the iwantmyname (Deprecated) provider:

```bash
IWANTMYNAME_USERNAME=xxxxxxxx \
IWANTMYNAME_PASSWORD=xxxxxxxx \
lego --dns iwantmyname -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `IWANTMYNAME_PASSWORD` | API password |
| `IWANTMYNAME_USERNAME` | API username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `IWANTMYNAME_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `IWANTMYNAME_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `IWANTMYNAME_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `IWANTMYNAME_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://iwantmyname.com/developer/domain-dns-api)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/iwantmyname/iwantmyname.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
