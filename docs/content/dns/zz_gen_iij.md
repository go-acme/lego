---
title: "Internet Initiative Japan"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: iij
dnsprovider:
  since:    "v1.1.0"
  code:     "iij"
  url:      "https://www.iij.ad.jp/en/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/iij/iij.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Internet Initiative Japan](https://www.iij.ad.jp/en/).


<!--more-->

- Code: `iij`
- Since: v1.1.0


Here is an example bash command using the Internet Initiative Japan provider:

```bash
IIJ_API_ACCESS_KEY=xxxxxxxx \
IIJ_API_SECRET_KEY=yyyyyy \
IIJ_DO_SERVICE_CODE=zzzzzz \
lego --email you@example.com --dns iij -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `IIJ_API_ACCESS_KEY` | API access key |
| `IIJ_API_SECRET_KEY` | API secret key |
| `IIJ_DO_SERVICE_CODE` | DO service code |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `IIJ_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 4) |
| `IIJ_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 240) |
| `IIJ_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://manual.iij.jp/p2/pubapi/)
- [Go client](https://github.com/iij/doapi)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/iij/iij.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
