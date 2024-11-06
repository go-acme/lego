---
title: "Selectel v2"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: selectelv2
dnsprovider:
  since:    "v4.17.0"
  code:     "selectelv2"
  url:      "https://selectel.ru"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/selectelv2/selectelv2.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Selectel v2](https://selectel.ru).


<!--more-->

- Code: `selectelv2`
- Since: v4.17.0


Here is an example bash command using the Selectel v2 provider:

```bash
SELECTELV2_USERNAME=trex \
SELECTELV2_PASSWORD=xxxxx \
SELECTELV2_ACCOUNT_ID=1234567 \
SELECTELV2_PROJECT_ID=111a11111aaa11aa1a11aaa11111aa1a \
lego --email you@example.com --dns selectelv2 --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SELECTELV2_ACCOUNT_ID` | Selectel account ID (INT) |
| `SELECTELV2_PASSWORD` | Openstack username's password |
| `SELECTELV2_PROJECT_ID` | Cloud project ID (UUID) |
| `SELECTELV2_USERNAME` | Openstack username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SELECTELV2_BASE_URL` | API endpoint URL |
| `SELECTELV2_HTTP_TIMEOUT` | API request timeout |
| `SELECTELV2_POLLING_INTERVAL` | Time between DNS propagation check |
| `SELECTELV2_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `SELECTELV2_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://developers.selectel.ru/docs/cloud-services/dns_api/dns_api_actual/)
- [Go client](https://github.com/selectel/domains-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/selectelv2/selectelv2.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
