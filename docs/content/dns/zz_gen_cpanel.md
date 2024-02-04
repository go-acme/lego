---
title: "CPanel/WHM"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: cpanel
dnsprovider:
  since:    "v4.16.0"
  code:     "cpanel"
  url:      "https://cpanel.net/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/cpanel/cpanel.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [CPanel/WHM](https://cpanel.net/).


<!--more-->

- Code: `cpanel`
- Since: v4.16.0


Here is an example bash command using the CPanel/WHM provider:

```bash
### CPANEL (default)

CPANEL_USERNAME = "yyyy"
CPANEL_TOKEN = "xxxx"
CPANEL_BASE_URL = "https://example.com:2083" \
CPANEL_NAMESERVER = "ns1.example.com:53" \
lego --email you@example.com --dns cpanel --domains my.example.org run

## WHM

CPANEL_MODE = whm
CPANEL_USERNAME = "yyyy"
CPANEL_TOKEN = "xxxx"
CPANEL_BASE_URL = "https://example.com:2087" \
CPANEL_NAMESERVER = "ns1.example.com:53" \
lego --email you@example.com --dns cpanel --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `CPANEL_BASE_URL` | API server URL |
| `CPANEL_NAMESERVER` | Nameserver |
| `CPANEL_TOKEN` | API token |
| `CPANEL_USERNAME` | username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `CPANEL_HTTP_TIMEOUT` | API request timeout |
| `CPANEL_MODE` | use cpanel API or WHM API (Default: cpanel) |
| `CPANEL_POLLING_INTERVAL` | Time between DNS propagation check |
| `CPANEL_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `CPANEL_REGION` | The region |
| `CPANEL_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information



<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/cpanel/cpanel.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
