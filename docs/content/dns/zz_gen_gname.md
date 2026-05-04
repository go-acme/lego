---
title: "Gname"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: gname
dnsprovider:
  since:    "v5.0.0"
  code:     "gname"
  url:      "https://www.gname.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gname/gname.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Gname](https://www.gname.com/).


<!--more-->

- Code: `gname`
- Since: v5.0.0


Here is an example bash command using the Gname provider:

```bash
GNAME_APP_ID="xxx" \
GNAME_APP_KEY="yyy" \
lego run --dns gname -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `GNAME_APP_ID` | App ID |
| `GNAME_APP_KEY` | App key |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GNAME_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `GNAME_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `GNAME_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `GNAME_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.gname.com/domain/api/jiexi/add)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gname/gname.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
