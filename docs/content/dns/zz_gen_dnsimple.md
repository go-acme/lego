---
title: "DNSimple"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dnsimple
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnsimple/dnsimple.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.3.0

Configuration for [DNSimple](https://dnsimple.com/).


<!--more-->

- Code: `dnsimple`

Here is an example bash command using the DNSimple provider:

```bash
DNSIMPLE_OAUTH_TOKEN=1234567890abcdefghijklmnopqrstuvwxyz \
lego --dns dnsimple --domains my.domain.com --email my@email.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DNSIMPLE_OAUTH_TOKEN` | OAuth token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DNSIMPLE_BASE_URL` | API endpoint URL |
| `DNSIMPLE_POLLING_INTERVAL` | Time between DNS propagation check |
| `DNSIMPLE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DNSIMPLE_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).

## Description

`DNSIMPLE_BASE_URL` is optional and must be set to production (https://api.dnsimple.com).
if `DNSIMPLE_BASE_URL` is not defined or empty, the production URL is used by default.

While you can manage DNS records in the [DNSimple Sandbox environment](https://developer.dnsimple.com/sandbox/),
DNS records will not resolve and you will not be able to satisfy the ACME DNS challenge.

To authenticate you need to provide a valid API token.
HTTP Basic Authentication is intentionally not supported.

### API tokens

You can [generate a new API token](https://support.dnsimple.com/articles/api-access-token/) from your account page.
Only Account API tokens are supported, if you try to use an User API token you will receive an error message.



## More information

- [API documentation](https://developer.dnsimple.com/v2/)
- [Go client](https://github.com/dnsimple/dnsimple-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnsimple/dnsimple.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
