---
title: "FusionLayer NameSurfer"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: namesurfer
dnsprovider:
  since:    "v4.32.0"
  code:     "namesurfer"
  url:      "https://www.fusionlayer.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/namesurfer/namesurfer.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [FusionLayer NameSurfer](https://www.fusionlayer.com/).


<!--more-->

- Code: `namesurfer`
- Since: v4.32.0


Here is an example bash command using the FusionLayer NameSurfer provider:

```bash
NAMESURFER_API_ENDPOINT=https://your-namesurfer-server.example.com:8443/API_10/NSService_10/jsonrpc10 \
NAMESURFER_API_KEY=your_api_key \
NAMESURFER_API_SECRET=your_api_secret \
lego --dns namesurfer -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NAMESURFER_API_ENDPOINT` | NameSurfer API endpoint URL (e.g., https://namesurfer.example.com:8443/API_10/NSService_10/) |
| `NAMESURFER_API_KEY` | API key name |
| `NAMESURFER_API_SECRET` | API secret |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NAMESURFER_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `NAMESURFER_INSECURE_SKIP_VERIFY` | Whether to verify the API certificate |
| `NAMESURFER_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `NAMESURFER_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `NAMESURFER_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |
| `NAMESURFER_VIEW` | DNS view name (optional, default: empty string) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](http://95.128.3.201:8053/API/NSService_10)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/namesurfer/namesurfer.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
