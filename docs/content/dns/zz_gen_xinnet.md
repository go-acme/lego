---
title: "Xinnet"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: xinnet
dnsprovider:
  since:    "v5.0.0"
  code:     "xinnet"
  url:      "https://www.xinnet.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/xinnet/xinnet.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Xinnet](https://www.xinnet.com/).


<!--more-->

- Code: `xinnet`
- Since: v5.0.0


Here is an example bash command using the Xinnet provider:

```bash
XINNET_SECRET="xxx" \
XINNET_AGENT_ID="agent12345" \
lego run --dns xinnet -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `XINNET_AGENT_ID` | Agent ID |
| `XINNET_SECRET` | Application secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `XINNET_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `XINNET_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `XINNET_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `XINNET_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://apidoc.xin.cn/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/xinnet/xinnet.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
