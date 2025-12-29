---
title: "Stackpath"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: stackpath
dnsprovider:
  since:    "v1.1.0"
  code:     "stackpath"
  url:      "https://www.stackpath.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/stackpath/stackpath.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Stackpath](https://www.stackpath.com/).


<!--more-->

- Code: `stackpath`
- Since: v1.1.0


Here is an example bash command using the Stackpath provider:

```bash
STACKPATH_CLIENT_ID=xxxxx \
STACKPATH_CLIENT_SECRET=yyyyy \
STACKPATH_STACK_ID=zzzzz \
lego --dns stackpath -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `STACKPATH_CLIENT_ID` | Client ID |
| `STACKPATH_CLIENT_SECRET` | Client secret |
| `STACKPATH_STACK_ID` | Stack ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `STACKPATH_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `STACKPATH_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `STACKPATH_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://developer.stackpath.com/en/api/dns/#tag/Zone)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/stackpath/stackpath.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
