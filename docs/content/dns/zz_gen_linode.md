---
title: "Linode (v4)"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: linode
dnsprovider:
  since:    "v1.1.0"
  code:     "linode"
  url:      "https://www.linode.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/linode/linode.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Linode (v4)](https://www.linode.com/).


<!--more-->

- Code: `linode`
- Since: v1.1.0


Here is an example bash command using the Linode (v4) provider:

```bash
LINODE_TOKEN=xxxxx \
lego --dns linode -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `LINODE_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `LINODE_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `LINODE_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 15) |
| `LINODE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 120) |
| `LINODE_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 300) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://developers.linode.com/api/v4)
- [Go client](https://github.com/linode/linodego)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/linode/linode.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
