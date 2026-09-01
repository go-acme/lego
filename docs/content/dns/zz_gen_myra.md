---
title: "Myra"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: myra
dnsprovider:
  since:    "v5.5.0"
  code:     "myra"
  url:      "https://www.myrasecurity.com/en/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/myra/myra.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Myra](https://www.myrasecurity.com/en/).


<!--more-->

- Code: `myra`
- Since: v5.5.0


Here is an example bash command using the Myra provider:

```bash
MYRA_API_KEY="xxx" \
MYRA_API_SECRET="xxx" \
lego run --dns myra -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `MYRA_API_KEY` | API key |
| `MYRA_API_SECRET` | API secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `MYRA_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 20) |
| `MYRA_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 600) |
| `MYRA_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.myracloud.com/en/api/general/using-the-myra-api/)
- [Go client](https://github.com/Myra-Security-GmbH/myrasec-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/myra/myra.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
