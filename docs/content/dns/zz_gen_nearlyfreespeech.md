---
title: "NearlyFreeSpeech.NET"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: nearlyfreespeech
dnsprovider:
  since:    "v4.8.0"
  code:     "nearlyfreespeech"
  url:      "https://nearlyfreespeech.net/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nearlyfreespeech/nearlyfreespeech.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [NearlyFreeSpeech.NET](https://nearlyfreespeech.net/).


<!--more-->

- Code: `nearlyfreespeech`
- Since: v4.8.0


Here is an example bash command using the NearlyFreeSpeech.NET provider:

```bash
NEARLYFREESPEECH_API_KEY=xxxxxx \
NEARLYFREESPEECH_LOGIN=xxxx \
lego --email you@example.com --dns nearlyfreespeech -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NEARLYFREESPEECH_API_KEY` | API Key for API requests |
| `NEARLYFREESPEECH_LOGIN` | Username for API requests |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NEARLYFREESPEECH_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `NEARLYFREESPEECH_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `NEARLYFREESPEECH_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `NEARLYFREESPEECH_SEQUENCE_INTERVAL` | Time between sequential requests in seconds (Default: 60) |
| `NEARLYFREESPEECH_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 3600) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://members.nearlyfreespeech.net/wiki/API/Reference)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nearlyfreespeech/nearlyfreespeech.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
