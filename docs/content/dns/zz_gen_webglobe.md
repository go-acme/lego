---
title: "Webglobe"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: webglobe
dnsprovider:
  since:    "v5.5.0"
  code:     "webglobe"
  url:      "https://www.webglobe.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/webglobe/webglobe.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Webglobe](https://www.webglobe.com/).


<!--more-->

- Code: `webglobe`
- Since: v5.5.0


Here is an example bash command using the Webglobe provider:

```bash
WEBGLOBE_LOGIN="xxx" \
WEBGLOBE_TOKEN="yyy" \
lego run --dns webglobe -d '*.example.com' -d example.com
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `WEBGLOBE_API_KEY` | API token |
| `WEBGLOBE_LOGIN` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `WEBGLOBE_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `WEBGLOBE_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `WEBGLOBE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `WEBGLOBE_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.webglobe.com/doc)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/webglobe/webglobe.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
