---
title: "SelfHost.(de|eu)"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: selfhostde
dnsprovider:
  since:    "v4.19.0"
  code:     "selfhostde"
  url:      "https://www.selfhost.de"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/selfhostde/selfhostde.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [SelfHost.(de|eu)](https://www.selfhost.de).


<!--more-->

- Code: `selfhostde`
- Since: v4.19.0


Here is an example bash command using the SelfHost.(de|eu) provider:

```bash
SELFHOSTDE_USERNAME=xxx \
SELFHOSTDE_PASSWORD=yyy \
SELFHOSTDE_RECORDS_MAPPING=my.example.com:123 \
lego --email you@example.com --dns selfhostde -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SELFHOSTDE_PASSWORD` | Password |
| `SELFHOSTDE_RECORDS_MAPPING` | Record IDs mapping with domains (ex: example.com:123:456,example.org:789,foo.example.com:147) |
| `SELFHOSTDE_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SELFHOSTDE_HTTP_TIMEOUT` | API request timeout |
| `SELFHOSTDE_POLLING_INTERVAL` | Time between DNS propagation check |
| `SELFHOSTDE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `SELFHOSTDE_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

SelfHost.de doesn't have an API to create or delete TXT records,
there is only an "unofficial" and undocumented endpoint to update an existing TXT record.

So, before using lego to request a certificate for a given domain or wildcard (such as `my.example.org` or `*.my.example.org`),
you must create:

- one TXT record named `_acme-challenge.my.example.org` if you are **not** using wildcard for this domain.
- two TXT records named `_acme-challenge.my.example.org` if you are using wildcard for this domain.

After that you must edit the TXT record(s) to get the ID(s).

You then must prepare the `SELFHOSTDE_RECORDS_MAPPING` environment variable with the following format:

```
<domain_A>:<record_id_A1>:<record_id_A2>,<domain_B>:<record_id_B1>:<record_id_B2>,<domain_C>:<record_id_C1>:<record_id_C2>
```

where each group of domain + record ID(s) is separated with a comma (`,`),
and the domain and record ID(s) are separated with a colon (`:`).

For example, if you want to create or renew a certificate for `my.example.org`, `*.my.example.org`, and `other.example.org`,
you would need:

- two separate records for `_acme-challenge.my.example.org`
- and another separate record for `_acme-challenge.other.example.org`

The resulting environment variable would then be: `SELFHOSTDE_RECORDS_MAPPING=my.example.com:123:456,other.example.com:789`





<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/selfhostde/selfhostde.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
