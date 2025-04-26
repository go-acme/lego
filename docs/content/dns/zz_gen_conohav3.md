---
title: "ConoHa v3"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: conohav3
dnsprovider:
  since:    "v4.24.0"
  code:     "conohav3"
  url:      "https://www.conoha.jp/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/conohav3/conohav3.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [ConoHa v3](https://www.conoha.jp/).


<!--more-->

- Code: `conohav3`
- Since: v4.24.0


Here is an example bash command using the ConoHa v3 provider:

```bash
CONOHAV3_TENANT_ID=487727e3921d44e3bfe7ebb337bf085e \
CONOHAV3_API_USER_ID=xxxx \
CONOHAV3_API_PASSWORD=yyyy \
lego --email you@example.com --dns conohav3 -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `CONOHAV3_API_PASSWORD` | The API password |
| `CONOHAV3_API_USER_ID` | The API user ID |
| `CONOHAV3_TENANT_ID` | Tenant ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `CONOHAV3_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `CONOHAV3_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `CONOHAV3_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `CONOHAV3_REGION` | The region (Default: c3j1) |
| `CONOHAV3_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://doc.conoha.jp/reference/api-vps3/api-dns-vps3/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/conohav3/conohav3.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
