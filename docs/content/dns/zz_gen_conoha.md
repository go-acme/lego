---
title: "ConoHa"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: conoha
dnsprovider:
  since:    "v1.2.0"
  code:     "conoha"
  url:      "https://www.conoha.jp/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/conoha/conoha.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [ConoHa](https://www.conoha.jp/).


<!--more-->

- Code: `conoha`
- Since: v1.2.0


Here is an example bash command using the ConoHa provider:

```bash
CONOHA_TENANT_ID=487727e3921d44e3bfe7ebb337bf085e \
CONOHA_API_USERNAME=xxxx \
CONOHA_API_PASSWORD=yyyy \
lego --email you@example.com --dns conoha --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `CONOHA_API_PASSWORD` | The API password |
| `CONOHA_API_USERNAME` | The API username |
| `CONOHA_TENANT_ID` | Tenant ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `CONOHA_HTTP_TIMEOUT` | API request timeout |
| `CONOHA_POLLING_INTERVAL` | Time between DNS propagation check |
| `CONOHA_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `CONOHA_REGION` | The region |
| `CONOHA_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://www.conoha.jp/docs/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/conoha/conoha.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
