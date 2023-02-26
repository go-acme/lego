---
title: "Nodion"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: nodion
dnsprovider:
  since:    "v4.11.0"
  code:     "nodion"
  url:      "https://www.nodion.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nodion/nodion.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Nodion](https://www.nodion.com).


<!--more-->

- Code: `nodion`
- Since: v4.11.0


Here is an example bash command using the Nodion provider:

```bash
NODION_API_TOKEN="xxxxxxxxxxxxxxxxxxxxx" \
lego --email myemail@example.com --dns nodion --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `NODION_API_TOKEN` | The API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `NODION_HTTP_TIMEOUT` | API request timeout |
| `NODION_POLLING_INTERVAL` | Time between DNS propagation check |
| `NODION_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `NODION_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://www.nodion.com/en/docs/dns/api/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/nodion/nodion.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
