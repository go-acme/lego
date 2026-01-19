---
title: "Bluecat v2"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: bluecatv2
dnsprovider:
  since:    "v4.32.0"
  code:     "bluecatv2"
  url:      "https://www.bluecatnetworks.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/bluecatv2/bluecatv2.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Bluecat v2](https://www.bluecatnetworks.com).


<!--more-->

- Code: `bluecatv2`
- Since: v4.32.0


Here is an example bash command using the Bluecat v2 provider:

```bash
BLUECATV2_SERVER_URL="https://example.com" \
BLUECATV2_USERNAME="xxx" \
BLUECATV2_PASSWORD="yyy" \
BLUECATV2_CONFIG_NAME="myConfiguration" \
BLUECATV2_VIEW_NAME="myView" \
lego --dns bluecatv2 -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `BLUECATV2_CONFIG_NAME` | Configuration name |
| `BLUECATV2_PASSWORD` | API password |
| `BLUECATV2_USERNAME` | API username |
| `BLUECATV2_VIEW_NAME` | DNS View Name |
| `BLUECAT_SERVER_URL` | The server URL: it should have a scheme, hostname, and port (if required) of the authoritative Bluecat BAM serve |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `BLUECATV2_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `BLUECATV2_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `BLUECATV2_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `BLUECATV2_SKIP_DEPLOY` | Skip quick deployements |
| `BLUECATV2_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.bluecatnetworks.com/r/Address-Manager-RESTful-v2-API-Guide/Introduction/9.6.0)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/bluecatv2/bluecatv2.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
