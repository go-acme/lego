---
title: "PowerDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: pdns
dnsprovider:
  since:    "v0.4.0"
  code:     "pdns"
  url:      "https://www.powerdns.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/pdns/pdns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [PowerDNS](https://www.powerdns.com/).


<!--more-->

- Code: `pdns`
- Since: v0.4.0


Here is an example bash command using the PowerDNS provider:

```bash
PDNS_API_URL=http://pdns-server:80/ \
PDNS_API_KEY=xxxx \
lego --email you@example.com --dns pdns --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `PDNS_API_KEY` | API key |
| `PDNS_API_URL` | API URL |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `PDNS_API_VERSION` | Skip API version autodetection and use the provided version number. |
| `PDNS_HTTP_TIMEOUT` | API request timeout |
| `PDNS_POLLING_INTERVAL` | Time between DNS propagation check |
| `PDNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `PDNS_SERVER_NAME` | Name of the server in the URL, 'localhost' by default |
| `PDNS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).

## Information

Tested and confirmed to work with PowerDNS authoritative server 3.4.8 and 4.0.1. Refer to [PowerDNS documentation](https://doc.powerdns.com/md/httpapi/README/) instructions on how to enable the built-in API interface.

PowerDNS Notes:
- PowerDNS API does not currently support SSL, therefore you should take care to ensure that traffic between lego and the PowerDNS API is over a trusted network, VPN etc.
- In order to have the SOA serial automatically increment each time the `_acme-challenge` record is added/modified via the API, set `SOA-EDIT-API` to `INCEPTION-INCREMENT` for the zone in the `domainmetadata` table
- Some PowerDNS servers doesn't have root API endpoints enabled and API version autodetection will not work. In that case version number can be defined using `PDNS_API_VERSION`.



## More information

- [API documentation](https://doc.powerdns.com/md/httpapi/README/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/pdns/pdns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
