---
title: "MyDNS.jp"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: mydnsjp
dnsprovider:
  since:    "v1.2.0"
  code:     "mydnsjp"
  url:      "https://www.mydns.jp"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/mydnsjp/mydnsjp.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [MyDNS.jp](https://www.mydns.jp).


<!--more-->

- Code: `mydnsjp`
- Since: v1.2.0


Here is an example bash command using the MyDNS.jp provider:

```bash
MYDNSJP_MASTER_ID=xxxxx \
MYDNSJP_PASSWORD=xxxxx \
lego --email you@example.com --dns mydnsjp -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `MYDNSJP_MASTER_ID` | Master ID |
| `MYDNSJP_PASSWORD` | Password |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `MYDNSJP_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `MYDNSJP_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `MYDNSJP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.mydns.jp/?MENU=030)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/mydnsjp/mydnsjp.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
