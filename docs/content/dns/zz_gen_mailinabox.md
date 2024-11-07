---
title: "Mail-in-a-Box"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: mailinabox
dnsprovider:
  since:    "v4.16.0"
  code:     "mailinabox"
  url:      "https://mailinabox.email"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/mailinabox/mailinabox.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Mail-in-a-Box](https://mailinabox.email).


<!--more-->

- Code: `mailinabox`
- Since: v4.16.0


Here is an example bash command using the Mail-in-a-Box provider:

```bash
MAILINABOX_EMAIL=user@example.com \
MAILINABOX_PASSWORD=yyyy \
MAILINABOX_BASE_URL=https://box.example.com \
lego --email you@example.com --dns mailinabox -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `MAILINABOX_BASE_URL` | Base API URL (ex: https://box.example.com) |
| `MAILINABOX_EMAIL` | User email |
| `MAILINABOX_PASSWORD` | User password |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `MAILINABOX_POLLING_INTERVAL` | Time between DNS propagation check |
| `MAILINABOX_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://mailinabox.email/api-docs.html)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/mailinabox/mailinabox.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
