---
title: "Azion"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: azion
dnsprovider:
  since:    "v4.24.0"
  code:     "azion"
  url:      "https://www.azion.com/en/products/edge-dns/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/azion/azion.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Azion](https://www.azion.com/en/products/edge-dns/).


<!--more-->

- Code: `azion`
- Since: v4.24.0


Here is an example bash command using the Azion provider:

```bash
AZION_PERSONAL_TOKEN=xxxxxxxxxxxxxxxxxxxxxxxxxx \
lego --email you@example.com --dns azion -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `AZION_PERSONAL_TOKEN` | Your Azion personal token. |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `AZION_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `AZION_PAGE_SIZE` | The page size for the API request (Default: 50) |
| `AZION_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `AZION_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `AZION_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.azion.com/)
- [Go client](https://github.com/aziontech/azionapi-go-sdk)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/azion/azion.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
