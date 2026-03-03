---
title: "Com Laude"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: comlaude
dnsprovider:
  since:    "v4.33.0"
  code:     "comlaude"
  url:      "https://comlaude.com/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/comlaude/comlaude.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Com Laude](https://comlaude.com/).


<!--more-->

- Code: `comlaude`
- Since: v4.33.0


Here is an example bash command using the Com Laude provider:

```bash
COMLAUDE_USERNAME="xxx" \
COMLAUDE_PASSWORD="yyy" \
COMLAUDE_API_KEY="zzz" \
COMLAUDE_GROUP_ID="abc" \
lego --dns comlaude -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `COMLAUDE_API_KEY` | API ley |
| `COMLAUDE_GROUP_ID` | Group ID |
| `COMLAUDE_PASSWORD` | password |
| `COMLAUDE_USERNAME` | username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `COMLAUDE_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `COMLAUDE_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `COMLAUDE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `COMLAUDE_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://api.comlaude.com/docs#/)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/comlaude/comlaude.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
