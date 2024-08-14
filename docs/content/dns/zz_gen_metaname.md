---
title: "Metaname"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: metaname
dnsprovider:
  since:    "v4.13.0"
  code:     "metaname"
  url:      "https://metaname.net"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/metaname/metaname.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Metaname](https://metaname.net).


<!--more-->

- Code: `metaname`
- Since: v4.13.0


Here is an example bash command using the Metaname provider:

```bash
METANAME_ACCOUNT_REFERENCE=xxxx \
METANAME_API_KEY=yyyyyyy \
lego --email you@example.com --dns metaname --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `METANAME_ACCOUNT_REFERENCE` | The four-digit reference of a Metaname account |
| `METANAME_API_KEY` | API Key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `METANAME_POLLING_INTERVAL` | Time between DNS propagation check |
| `METANAME_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `METANAME_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://metaname.net/api/1.1/doc)
- [Go client](https://github.com/nzdjb/go-metaname)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/metaname/metaname.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
