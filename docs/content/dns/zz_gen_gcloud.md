---
title: "Google Cloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: gcloud
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gcloud/gcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.3.0

Configuration for [Google Cloud](https://cloud.google.com).


<!--more-->

- Code: `gcloud`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `Application Default Credentials` | [Documentation](https://cloud.google.com/docs/authentication/production#providing_credentials_to_your_application) |
| `GCE_PROJECT` | Project name |
| `GCE_SERVICE_ACCOUNT` | Account |
| `GCE_SERVICE_ACCOUNT_FILE` | Account file path |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `GCE_POLLING_INTERVAL` | Time between DNS propagation check |
| `GCE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `GCE_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://community.exoscale.com/documentation/dns/api/)
- [Go client](https://github.com/googleapis/google-api-go-client)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/gcloud/gcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
