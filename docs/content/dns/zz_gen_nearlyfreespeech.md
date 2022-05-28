---
title: "NearlyFreeSpeech.NET"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: nearlyfreespeech
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nearlyfreespeech/nearlyfreespeech.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v4.8.0

Configuration for [NearlyFreeSpeech.NET](https://nearlyfreespeech.net/).


<!--more-->

- Code: `nearlyfreespeech`

Here is an example bash command using the NearlyFreeSpeech.NET provider:

```bash
NEARLYFREESPEECH_API_KEY=xxxxxx \
NEARLYFREESPEECH_LOGIN=xxxx \
lego --email myemail@example.com --dns nearlyfreespeech --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NEARLYFREESPEECH_API_KEY` | API Key for API requests |
| `NEARLYFREESPEECH_LOGIN` | Username for API requests |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NEARLYFREESPEECH_HTTP_TIMEOUT` | API request timeout |
| `NEARLYFREESPEECH_POLLING_INTERVAL` | Time between DNS propagation check |
| `NEARLYFREESPEECH_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `NEARLYFREESPEECH_SEQUENCE_INTERVAL` | Time between sequential requests |
| `NEARLYFREESPEECH_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://members.nearlyfreespeech.net/wiki/API/Reference)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nearlyfreespeech/nearlyfreespeech.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
