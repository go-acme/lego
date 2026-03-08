---
title: "EuroDNS"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: eurodns
dnsprovider:
  since:    "v4.33.0"
  code:     "eurodns"
  url:      "https://www.eurodns.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/eurodns/eurodns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [EuroDNS](https://www.eurodns.com/).


<!--more-->

- Code: `eurodns`
- Since: v4.33.0


Here is an example bash command using the EuroDNS provider:

```bash
EURODNS_APP_ID="xxx" \
EURODNS_API_KEY="yyy" \
lego --dns eurodns -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `EURODNS_API_KEY` | API key |
| `EURODNS_APP_ID` | Application ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `EURODNS_HTTP_TIMEOUT` | API request timeout |
| `EURODNS_POLLING_INTERVAL` | Time between DNS propagation check |
| `EURODNS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `EURODNS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docapi.eurodns.com/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/eurodns/eurodns.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
