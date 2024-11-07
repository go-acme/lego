---
title: "Bindman"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: bindman
dnsprovider:
  since:    "v2.6.0"
  code:     "bindman"
  url:      "https://github.com/labbsr0x/bindman-dns-webhook"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/bindman/bindman.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Bindman](https://github.com/labbsr0x/bindman-dns-webhook).


<!--more-->

- Code: `bindman`
- Since: v2.6.0


Here is an example bash command using the Bindman provider:

```bash
BINDMAN_MANAGER_ADDRESS=<your bindman manager address> \
lego --email you@example.com --dns bindman -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `BINDMAN_MANAGER_ADDRESS` | The server URL, should have scheme, hostname, and port (if required) of the Bindman-DNS Manager server |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `BINDMAN_HTTP_TIMEOUT` | API request timeout |
| `BINDMAN_POLLING_INTERVAL` | Time between DNS propagation check |
| `BINDMAN_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://gitlab.isc.org/isc-projects/bind9)
- [Go client](https://github.com/labbsr0x/bindman-dns-webhook)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/bindman/bindman.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
