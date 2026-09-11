---
title: "RackCorp"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: rackcorp
dnsprovider:
  since:    "v5.0.0"
  code:     "rackcorp"
  url:      "https://www.rackcorp.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rackcorp/rackcorp.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [RackCorp](https://www.rackcorp.com/).


<!--more-->

- Code: `rackcorp`
- Since: v5.0.0


Here is an example bash command using the RackCorp provider:

```bash
RACKCORP_API_UUID="xxx" \
RACKCORP_API_SECRET="yyy" \
lego --dns rackcorp -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `RACKCORP_API_SECRET` | API secret |
| `RACKCORP_API_UUID` | API UUID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `RACKCORP_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `RACKCORP_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `RACKCORP_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `RACKCORP_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://wiki.rackcorp.com/books/help-and-support-en/chapter/rackcorp-rest-api)
- [Go client](https://github.com/RackCorpCloud/rackcorp-api-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/rackcorp/rackcorp.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
