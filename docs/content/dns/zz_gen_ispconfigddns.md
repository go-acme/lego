---
title: "ISPConfig 3 - Dynamic DNS (DDNS) Module"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: ispconfigddns
dnsprovider:
  since:    "v4.31.0"
  code:     "ispconfigddns"
  url:      "https://www.ispconfig.org/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ispconfigddns/ispconfigddns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [ISPConfig 3 - Dynamic DNS (DDNS) Module](https://www.ispconfig.org/).


<!--more-->

- Code: `ispconfigddns`
- Since: v4.31.0


Here is an example bash command using the ISPConfig 3 - Dynamic DNS (DDNS) Module provider:

```bash
ISPCONFIG_DDNS_SERVER_URL="https://panel.example.com:8080" \
ISPCONFIG_DDNS_TOKEN=xxxxxx \
lego --dns ispconfigddns -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ISPCONFIG_DDNS_SERVER_URL` | API server URL (ex: https://panel.example.com:8080) |
| `ISPCONFIG_DDNS_TOKEN` | DDNS API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ISPCONFIG_DDNS_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `ISPCONFIG_DDNS_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `ISPCONFIG_DDNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `ISPCONFIG_DDNS_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 3600) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

ISPConfig DNS provider supports leveraging the [ISPConfig 3 Dynamic DNS (DDNS) Module](https://github.com/mhofer117/ispconfig-ddns-module).

Requires the DDNS module described at https://www.ispconfig.org/ispconfig/download/

See https://www.howtoforge.com/community/threads/ispconfig-3-danymic-dns-ddns-module.87967/ for additional details.



## More information

- [API documentation](https://github.com/mhofer117/ispconfig-ddns-module/tree/master/lib/updater)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ispconfigddns/ispconfigddns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
