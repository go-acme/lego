---
title: "TodayNIC/时代互联"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: todaynic
dnsprovider:
  since:    "v4.32.0"
  code:     "todaynic"
  url:      "https://www.todaynic.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/todaynic/todaynic.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [TodayNIC/时代互联](https://www.todaynic.com/).


<!--more-->

- Code: `todaynic`
- Since: v4.32.0


Here is an example bash command using the TodayNIC/时代互联 provider:

```bash
TODAYNIC_AUTH_USER_ID="xxx" \
TODAYNIC_API_KEY="yyy" \
lego --dns todaynic -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `TODAYNIC_API_KEY` | API key |
| `TODAYNIC_AUTH_USER_ID` | account ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `TODAYNIC_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `TODAYNIC_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `TODAYNIC_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `TODAYNIC_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 600) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://www.todaynic.com/partner/mode_Http_Api_detail.php)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/todaynic/todaynic.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
