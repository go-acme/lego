---
title: "Gehirn"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: gehirn
dnsprovider:
  since:    "v5.0.0"
  code:     "gehirn"
  url:      "https://www.gehirn.jp/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gehirn/gehirn.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Gehirn](https://www.gehirn.jp/).


<!--more-->

- Code: `gehirn`
- Since: v5.0.0


Here is an example bash command using the Gehirn provider:

```bash
GEHIRN_TOKEN_ID="xxx" \
GEHIRN_TOKEN_SECRET="xxx" \
lego --dns gehirn -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `GEHIRN_TOKEN_ID` | Token ID |
| `GEHIRN_TOKEN_SECRET` | Token secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GEHIRN_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `GEHIRN_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `GEHIRN_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `GEHIRN_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://support.gehirn.jp/apidocs/dns/index.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gehirn/gehirn.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
