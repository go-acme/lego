---
title: "Hurricane Electric DNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: hurricane
dnsprovider:
  since:    "v4.3.0"
  code:     "hurricane"
  url:      "https://dns.he.net/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hurricane/hurricane.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Hurricane Electric DNS](https://dns.he.net/).


<!--more-->

- Code: `hurricane`
- Since: v4.3.0


Here is an example bash command using the Hurricane Electric DNS provider:

```bash
HURRICANE_TOKENS=example.org:token \
lego --email you@example.com --dns hurricane --domains example.org --domains '*.example.org' run

HURRICANE_TOKENS=my.example.org:token1,demo.example.org:token2 \
lego --email you@example.com --dns hurricane --domains my.example.org --domains demo.example.org
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `HURRICANE_TOKENS` | TXT record names and tokens |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).



Before using lego to request a certificate for a given domain or wildcard (such as `my.example.org` or `*.my.example.org`),
create a TXT record named `_acme-challenge.my.example.org`, and enable dynamic updates on it.
Generate a token for each URL with Hurricane Electric's UI, and copy it down.
Stick to alphanumeric tokens for greatest reliability.

To authenticate with the Hurricane Electric API,
add each record name/token pair you want to update to the `HURRICANE_TOKENS` environment variable, as shown in the examples.
Record names (without the `_acme-challenge.` component) and their tokens are separated with colons,
while the credential pairs are concatenated into a comma-separated list, like so:

```
HURRICANE_TOKENS=my.example.org:token1,demo.example.org:token2
```

If you are issuing both a wildcard certificate and a standard certificate for a given subdomain,
you should not have repeat entries for that name, as both will use the same credential.

```
HURRICANE_TOKENS=example.org:token
```



## More information

- [API documentation](https://dns.he.net/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hurricane/hurricane.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
