---
title: "Servercow"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: servercow
dnsprovider:
  since:    "v3.4.0"
  code:     "servercow"
  url:      "https://servercow.de/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/servercow/servercow.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Servercow](https://servercow.de/).


<!--more-->

- Code: `servercow`
- Since: v3.4.0


Here is an example bash command using the Servercow provider:

```bash
SERVERCOW_USERNAME=xxxxxxxx \
SERVERCOW_PASSWORD=xxxxxxxx \
lego --email you@example.com --dns servercow --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SERVERCOW_PASSWORD` | API password |
| `SERVERCOW_USERNAME` | API username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SERVERCOW_HTTP_TIMEOUT` | API request timeout |
| `SERVERCOW_POLLING_INTERVAL` | Time between DNS propagation check |
| `SERVERCOW_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `SERVERCOW_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://cp.servercow.de/client/plugin/support_manager/knowledgebase/view/34/dns-api-v1/7/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/servercow/servercow.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
