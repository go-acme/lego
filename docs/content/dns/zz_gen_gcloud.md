---
title: "Google Cloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: gcloud
dnsprovider:
  since:    "v0.3.0"
  code:     "gcloud"
  url:      "https://cloud.google.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gcloud/gcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Google Cloud](https://cloud.google.com).


<!--more-->

- Code: `gcloud`
- Since: v0.3.0


Here is an example bash command using the Google Cloud provider:

```bash
GCE_PROJECT="gc-project-id" GCE_SERVICE_ACCOUNT_FILE="/path/to/svc/account/file.json" lego \
    --email="abc@email.com" \
    --domains="example.com" \
    --dns="gcloud" \
    --path="${HOME}/.lego" \
    run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `Application Default Credentials` | [Documentation](https://cloud.google.com/docs/authentication/production#providing_credentials_to_your_application) |
| `GCE_PROJECT` | Project name (by default, the project name is auto-detected by using the metadata service) |
| `GCE_SERVICE_ACCOUNT` | Account |
| `GCE_SERVICE_ACCOUNT_FILE` | Account file path |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GCE_ALLOW_PRIVATE_ZONE` | Allows requested domain to be in private DNS zone, works only with a private ACME server (by default: false) |
| `GCE_POLLING_INTERVAL` | Time between DNS propagation check |
| `GCE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `GCE_TTL` | The TTL of the TXT record used for the DNS challenge |
| `GCE_ZONE_ID` | Allows to skip the automatic detection of the zone |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://cloud.google.com/dns/api/v1/)
- [Go client](https://github.com/googleapis/google-api-go-client)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gcloud/gcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
