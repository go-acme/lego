---
title: "Brandit"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: brandit
dnsprovider:
  since:    "v4.11.0"
  code:     "brandit"
  url:      "https://www.brandit.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/brandit/brandit.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Brandit](https://www.brandit.com/).


<!--more-->

- Code: `brandit`
- Since: v4.11.0


Here is an example bash command using the Brandit provider:

```bash
BRANDIT_API_KEY=xxxxxxxxxxxxxxxxxxxxx \
BRANDIT_API_USERNAME=yyyyyyyyyyyyyyyyyyyy \
lego --email myemail@example.com --dns brandit --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `BRANDIT_API_KEY` | The API key |
| `BRANDIT_API_USERNAME` | The API username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `BRANDIT_HTTP_TIMEOUT` | API request timeout |
| `BRANDIT_POLLING_INTERVAL` | Time between DNS propagation check |
| `BRANDIT_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `BRANDIT_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://portal.brandit.com/apidocv3)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/brandit/brandit.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
