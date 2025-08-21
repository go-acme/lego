---
title: "Binary Lane"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: binarylane
dnsprovider:
  since:    "v4.26.0"
  code:     "binarylane"
  url:      "https://www.binarylane.com.au/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/binarylane/binarylane.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Binary Lane](https://www.binarylane.com.au/).


<!--more-->

- Code: `binarylane`
- Since: v4.26.0


Here is an example bash command using the Binary Lane provider:

```bash
BINARYLANE_API_TOKEN="xxxxxxxxxxxxxxxxxxxxx" \
lego --email you@example.com --dns binarylane -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `BINARYLANE_API_TOKEN` | API token |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `BINARYLANE_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `BINARYLANE_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `BINARYLANE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `BINARYLANE_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.binarylane.com.au/reference/#tag/Domains)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/binarylane/binarylane.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
