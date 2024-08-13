---
title: "Joohoi's ACME-DNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: acme-dns
dnsprovider:
  since:    "v1.1.0"
  code:     "acme-dns"
  url:      "https://github.com/joohoi/acme-dns"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/acmedns/acmedns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Joohoi's ACME-DNS](https://github.com/joohoi/acme-dns).


<!--more-->

- Code: `acme-dns`
- Since: v1.1.0


Here is an example bash command using the Joohoi's ACME-DNS provider:

```bash
ACME_DNS_API_BASE=http://10.0.0.8:4443 \
ACME_DNS_STORAGE_PATH=/root/.lego-acme-dns-accounts.json \
lego --email you@example.com --dns acme-dns --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ACME_DNS_API_BASE` | The ACME-DNS API address |
| `ACME_DNS_STORAGE_PATH` | The ACME-DNS JSON account data file. A per-domain account will be registered/persisted to this file and used for TXT updates. |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).






## More information

- [API documentation](https://github.com/joohoi/acme-dns#api)
- [Go client](https://github.com/cpu/goacmedns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/acmedns/acmedns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
