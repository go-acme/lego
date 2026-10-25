---
title: "NordName"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: nordname
dnsprovider:
  since:    "v5.6.0"
  code:     "nordname"
  url:      "https://nordname.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nordname/nordname.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [NordName](https://nordname.com/).


<!--more-->

- Code: `nordname`
- Since: v5.6.0


Here is an example bash command using the NordName provider:

```bash
NORDNAME_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego run --dns nordname -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NORDNAME_API_KEY` | API key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NORDNAME_CUSTOMER_UUID` | UUID of the customer to access. |
| `NORDNAME_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `NORDNAME_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `NORDNAME_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `NORDNAME_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://app.nordname.eu/en/api-docs)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nordname/nordname.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
