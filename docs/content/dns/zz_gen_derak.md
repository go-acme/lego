---
title: "Derak Cloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: derak
dnsprovider:
  since:    "v4.12.0"
  code:     "derak"
  url:      "https://derak.cloud/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/derak/derak.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Derak Cloud](https://derak.cloud/).


<!--more-->

- Code: `derak`
- Since: v4.12.0


Here is an example bash command using the Derak Cloud provider:

```bash
DERAK_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego --email myemail@example.com --dns derak --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DERAK_API_KEY` | The API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DERAK_HTTP_TIMEOUT` | API request timeout |
| `DERAK_POLLING_INTERVAL` | Time between DNS propagation check |
| `DERAK_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DERAK_TTL` | The TTL of the TXT record used for the DNS challenge |
| `DERAK_WEBSITE_ID` | Force the zone/website ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).





<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/derak/derak.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
