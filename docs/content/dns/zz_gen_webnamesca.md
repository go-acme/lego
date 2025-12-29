---
title: "webnames.ca"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: webnamesca
dnsprovider:
  since:    "v4.28.0"
  code:     "webnamesca"
  url:      "https://www.webnames.ca/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/webnamesca/webnamesca.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [webnames.ca](https://www.webnames.ca/).


<!--more-->

- Code: `webnamesca`
- Since: v4.28.0


Here is an example bash command using the webnames.ca provider:

```bash
WEBNAMESCA_API_USER="xxx" \
WEBNAMESCA_API_KEY="yyy" \
lego --dns webnamesca -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `WEBNAMESCA_API_KEY` | API key |
| `WEBNAMESCA_API_USER` | API username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `WEBNAMESCA_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `WEBNAMESCA_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `WEBNAMESCA_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `WEBNAMESCA_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.webnames.ca/_/swagger/index.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/webnamesca/webnamesca.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
