---
title: "Akamai EdgeDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: edgedns
dnsprovider:
  since:    "v3.9.0"
  code:     "edgedns"
  url:      "https://www.akamai.com/us/en/products/security/edge-dns.jsp"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/edgedns/edgedns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Akamai edgedns supersedes FastDNS; implementing a DNS provider for solving the DNS-01 challenge using Akamai EdgeDNS



<!--more-->

- Code: `edgedns`
- Since: v3.9.0


Here is an example bash command using the Akamai EdgeDNS provider:

```bash
AKAMAI_CLIENT_SECRET=abcdefghijklmnopqrstuvwxyz1234567890ABCDEFG= \
AKAMAI_CLIENT_TOKEN=akab-mnbvcxzlkjhgfdsapoiuytrewq1234567 \
AKAMAI_HOST=akab-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.luna.akamaiapis.net \
AKAMAI_ACCESS_TOKEN=akab-1234567890qwerty-asdfghjklzxcvtnu \
lego --email you@example.com --dns edgedns --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `AKAMAI_ACCESS_TOKEN` | Access token, managed by the Akamai EdgeGrid client |
| `AKAMAI_CLIENT_SECRET` | Client secret, managed by the Akamai EdgeGrid client |
| `AKAMAI_CLIENT_TOKEN` | Client token, managed by the Akamai EdgeGrid client |
| `AKAMAI_EDGERC` | Path to the .edgerc file, managed by the Akamai EdgeGrid client |
| `AKAMAI_EDGERC_SECTION` | Configuration section, managed by the Akamai EdgeGrid client |
| `AKAMAI_HOST` | API host, managed by the Akamai EdgeGrid client |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `AKAMAI_POLLING_INTERVAL` | Time between DNS propagation check. Default: 15 seconds |
| `AKAMAI_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation. Default: 3 minutes |
| `AKAMAI_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

Akamai's credentials are automatically detected in the following locations and prioritized in the following order:

1. Section-specific environment variables (where `{SECTION}` is specified using `AKAMAI_EDGERC_SECTION`):
  - `AKAMAI_{SECTION}_HOST`
  - `AKAMAI_{SECTION}_ACCESS_TOKEN`
  - `AKAMAI_{SECTION}_CLIENT_TOKEN`
  - `AKAMAI_{SECTION}_CLIENT_SECRET`
2. If `AKAMAI_EDGERC_SECTION` is not defined or is set to `default`, environment variables:
  - `AKAMAI_HOST`
  - `AKAMAI_ACCESS_TOKEN`
  - `AKAMAI_CLIENT_TOKEN`
  - `AKAMAI_CLIENT_SECRET`
3. `.edgerc` file located at `AKAMAI_EDGERC`
  - defaults to `~/.edgerc`, sections can be specified using `AKAMAI_EDGERC_SECTION`
4. Default environment variables:
  - `AKAMAI_HOST`
  - `AKAMAI_ACCESS_TOKEN`
  - `AKAMAI_CLIENT_TOKEN`
  - `AKAMAI_CLIENT_SECRET`

See also:

- [Setting up Akamai credentials](https://developer.akamai.com/api/getting-started)
- [.edgerc Format](https://developer.akamai.com/legacy/introduction/Conf_Client.html#edgercformat)
- [API Client Authentication](https://developer.akamai.com/legacy/introduction/Client_Auth.html)
- [Config from Env](https://github.com/akamai/AkamaiOPEN-edgegrid-golang/blob/master/pkg/edgegrid/config.go#L118)



## More information

- [API documentation](https://developer.akamai.com/api/cloud_security/edge_dns_zone_management/v2.html)
- [Go client](https://github.com/akamai/AkamaiOPEN-edgegrid-golang)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/edgedns/edgedns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
