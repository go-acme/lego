---
title: "Openprovider"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: openprovider
dnsprovider:
  since:    "v5.3.0"
  code:     "openprovider"
  url:      "https://www.openprovider.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/openprovider/openprovider.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Openprovider](https://www.openprovider.com/).


<!--more-->

- Code: `openprovider`
- Since: v5.3.0


Here is an example bash command using the Openprovider provider:

```bash
OPENPROVIDER_USERNAME="xxx" \
OPENPROVIDER_PASSWORD="yyy" \
lego run --dns openprovider -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `OPENPROVIDER_PASSWORD` | The user's password |
| `OPENPROVIDER_USERNAME` | The user's name |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `OPENPROVIDER_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `OPENPROVIDER_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 10) |
| `OPENPROVIDER_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 600) |
| `OPENPROVIDER_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 600) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

{{% notice style="warning" %}}

The provider is only available for resellers because the Openprovider API is only available for resellers.

{{% /notice %}}



## More information

- [API documentation](https://docs.openprovider.com/doc/all)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/openprovider/openprovider.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
