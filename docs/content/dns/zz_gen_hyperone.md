---
title: "HyperOne"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: hyperone
dnsprovider:
  since:    "v3.9.0"
  code:     "hyperone"
  url:      "https://www.hyperone.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hyperone/hyperone.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [HyperOne](https://www.hyperone.com).


<!--more-->

- Code: `hyperone`
- Since: v3.9.0


Here is an example bash command using the HyperOne provider:

```bash
lego --email you@example.com --dns hyperone --domains my.example.org run
```






## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `HYPERONE_API_URL` | Allows to pass custom API Endpoint to be used in the challenge (default https://api.hyperone.com/v2) |
| `HYPERONE_LOCATION_ID` | Specifies location (region) to be used in API calls. (default pl-waw-1) |
| `HYPERONE_PASSPORT_LOCATION` | Allows to pass custom passport file location (default ~/.h1/passport.json) |
| `HYPERONE_POLLING_INTERVAL` | Time between DNS propagation check |
| `HYPERONE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `HYPERONE_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).

## Description

Default configuration does not require any additional environment variables,
just a passport file in `~/.h1/passport.json` location.

### Generating passport file using H1 CLI

To use this application you have to generate passport file for `sa`:

```
h1 iam project sa credential generate --name my-passport --project <project ID> --sa <sa ID> --passport-output-file ~/.h1/passport.json
```

### Required permissions

The application requires following permissions:
-  `dns/zone/list`
-  `dns/zone.recordset/list`
-  `dns/zone.recordset/create`
-  `dns/zone.recordset/delete`
-  `dns/zone.record/create`
-  `dns/zone.record/list`
-  `dns/zone.record/delete`

All required permissions are available via platform role `tool.lego`.



## More information

- [API documentation](https://api.hyperone.com/v2/docs)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/hyperone/hyperone.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
