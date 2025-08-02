---
title: "Tencent EdgeOne"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: edgeone
dnsprovider:
  since:    "v4.26.0"
  code:     "edgeone"
  url:      "https://edgeone.ai"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/edgeone/edgeone.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Tencent EdgeOne](https://edgeone.ai).


<!--more-->

- Code: `edgeone`
- Since: v4.26.0


Here is an example bash command using the Tencent EdgeOne provider:

```bash
EDGEONE_SECRET_ID=abcdefghijklmnopqrstuvwx \
EDGEONE_SECRET_KEY=your-secret-key \
lego --email you@example.com --dns edgeone -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `EDGEONE_SECRET_ID` | Access key ID |
| `EDGEONE_SECRET_KEY` | Access Key secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `EDGEONE_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `EDGEONE_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 30) |
| `EDGEONE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 1200) |
| `EDGEONE_REGION` | Region |
| `EDGEONE_SESSION_TOKEN` | Access Key token |
| `EDGEONE_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://edgeone.ai/document/50454#dns-record-apis)
- [Go client](https://github.com/tencentcloud/tencentcloud-sdk-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/edgeone/edgeone.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
