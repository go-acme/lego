---
title: "Aurora DNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: auroradns
dnsprovider:
  since:    "v0.4.0"
  code:     "auroradns"
  url:      "https://www.pcextreme.com/dns-health-checks"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/auroradns/auroradns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Aurora DNS](https://www.pcextreme.com/dns-health-checks).


<!--more-->

- Code: `auroradns`
- Since: v0.4.0


Here is an example bash command using the Aurora DNS provider:

```bash
AURORA_API_KEY=xxxxx \
AURORA_SECRET=yyyyyy \
lego --email you@example.com --dns auroradns --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `AURORA_API_KEY` | API key or username to used |
| `AURORA_SECRET` | Secret password to be used |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `AURORA_ENDPOINT` | API endpoint URL |
| `AURORA_POLLING_INTERVAL` | Time between DNS propagation check |
| `AURORA_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `AURORA_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://libcloud.readthedocs.io/en/latest/dns/drivers/auroradns.html#api-docs)
- [Go client](https://github.com/nrdcg/auroradns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/auroradns/auroradns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
