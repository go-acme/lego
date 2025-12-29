---
title: "ZoneEdit"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: zoneedit
dnsprovider:
  since:    "v4.25.0"
  code:     "zoneedit"
  url:      "https://www.zoneedit.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/zoneedit/zoneedit.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [ZoneEdit](https://www.zoneedit.com).


<!--more-->

- Code: `zoneedit`
- Since: v4.25.0


Here is an example bash command using the ZoneEdit provider:

```bash
ZONEEDIT_USER="xxxxxxxxxxxxxxxxxxxxx" \
ZONEEDIT_AUTH_TOKEN="xxxxxxxxxxxxxxxxxxxxx" \
lego --dns zoneedit -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `ZONEEDIT_AUTH_TOKEN` | Authentication token |
| `ZONEEDIT_USER` | User ID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `ZONEEDIT_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `ZONEEDIT_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `ZONEEDIT_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://support.zoneedit.com/en/knowledgebase/article/changes-to-dynamic-dns)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/zoneedit/zoneedit.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
