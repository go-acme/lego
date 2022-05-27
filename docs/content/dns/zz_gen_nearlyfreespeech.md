---
title: "nearlyfreespeech.net"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: nearlyfreespeech
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nearlyfreespeech/nearlyfreespeech.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.1.0

Configuration for [nearlyfreespeech.net](https://nearlyfreespeech.net/).


<!--more-->

- Code: `nearlyfreespeech`

Here is an example bash command using the nearlyfreespeech.net provider:

```bash
NFS_API_KEY=xxxxxx \
NFS_LOGIN=xxxx \
lego --email myemail@example.com --dns nearlyfreespeech --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NFS_API_KEY` | API Key for API requests |
| `NFS_LOGIN` | Username for API requests |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NFS_HTTP_TIMEOUT` | API request timeout |
| `NFS_POLLING_INTERVAL` | Time between DNS propagation check |
| `NFS_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `NFS_SEQUENCE_INTERVAL` | Time between sequential requests |
| `NFS_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://api.nearlyfreespeech.net)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nearlyfreespeech/nearlyfreespeech.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
