---
title: "iwantmyname"
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


Configuration for [iwantmyname](https://iwantmyname.com).


<!--more-->

- Code: `iwantmyname`
- Since: v4.7.0


Here is an example bash command using the iwantmyname provider:

```bash
IWANTMYNAME_USERNAME=xxxxxxxx \
IWANTMYNAME_PASSWORD=xxxxxxxx \
lego --email you@example.com --dns iwantmyname --domains my.example.org run
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
| `IWANTMYNAME_HTTP_TIMEOUT` | API request timeout |
| `IWANTMYNAME_POLLING_INTERVAL` | Time between DNS propagation check |
| `IWANTMYNAME_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `IWANTMYNAME_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://iwantmyname.com/developer/domain-dns-api)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/iwantmyname/iwantmyname.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
