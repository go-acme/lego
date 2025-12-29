---
title: "Alwaysdata"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: alwaysdata
dnsprovider:
  since:    "v4.31.0"
  code:     "alwaysdata"
  url:      "https://alwaysdata.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/alwaysdata/alwaysdata.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Alwaysdata](https://alwaysdata.com/).


<!--more-->

- Code: `alwaysdata`
- Since: v4.31.0


Here is an example bash command using the Alwaysdata provider:

```bash
ALWAYSDATA_API_KEY="xxxxxxxxxxxxxxxxxxxxx" \
lego --dns alwaysdata -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ALWAYSDATA_API_KEY` | API Key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ALWAYSDATA_ACCOUNT` | Account name |
| `ALWAYSDATA_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `ALWAYSDATA_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `ALWAYSDATA_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `ALWAYSDATA_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://help.alwaysdata.com/en/api/resources/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/alwaysdata/alwaysdata.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
