---
title: "BookMyName"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: bookmyname
dnsprovider:
  since:    "v4.23.0"
  code:     "bookmyname"
  url:      "https://www.bookmyname.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/bookmyname/bookmyname.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [BookMyName](https://www.bookmyname.com/).


<!--more-->

- Code: `bookmyname`
- Since: v4.23.0


Here is an example bash command using the BookMyName provider:

```bash
BOOKMYNAME_USERNAME="xxx" \
BOOKMYNAME_PASSWORD="yyy" \
lego --email you@example.com --dns bookmyname -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `BOOKMYNAME_PASSWORD` | Password |
| `BOOKMYNAME_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `BOOKMYNAME_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `BOOKMYNAME_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `BOOKMYNAME_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `BOOKMYNAME_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://fr.faqs.bookmyname.com/frfaqs/dyndns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/bookmyname/bookmyname.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
