---
title: "RcodeZero"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: rcodezero
dnsprovider:
  since:    "v4.13"
  code:     "rcodezero"
  url:      "https://www.rcodezero.at/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rcodezero/rcodezero.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [RcodeZero](https://www.rcodezero.at/).


<!--more-->

- Code: `rcodezero`
- Since: v4.13


Here is an example bash command using the RcodeZero provider:

```bash
RCODEZERO_API_TOKEN=<mytoken> \
lego --email you@example.com --dns rcodezero -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `RCODEZERO_API_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `RCODEZERO_HTTP_TIMEOUT` | API request timeout |
| `RCODEZERO_POLLING_INTERVAL` | Time between DNS propagation check |
| `RCODEZERO_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `RCODEZERO_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

## Description

Generate your API Token via https://my.rcodezero.at with the `ACME` permissions.
These are special tokens with limited access for ACME requests only.

RcodeZero is an Anycast Network so the distribution of the DNS01-Challenge can take up to 2 minutes.




## More information

- [API documentation](https://my.rcodezero.at/openapi)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rcodezero/rcodezero.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
