---
title: "Azure"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: azure
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/azure/azure.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v0.4.0

Configuration for [Azure](https://azure.microsoft.com/services/dns/).


<!--more-->

- Code: `azure`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `AZURE_CLIENT_ID` | Client ID |
| `AZURE_CLIENT_SECRET` | Client secret |
| `AZURE_RESOURCE_GROUP` | Resource group |
| `AZURE_SUBSCRIPTION_ID` | Subscription ID |
| `AZURE_TENANT_ID` | Tenant ID |
| `instance metadata service` | If the credentials are **not** set via the environment, then it will attempt to get a bearer token via the [instance metadata service](https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service). |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `AZURE_METADATA_ENDPOINT` | Metadata Service endpoint URL |
| `AZURE_POLLING_INTERVAL` | Time between DNS propagation check |
| `AZURE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `AZURE_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://docs.microsoft.com/en-us/go/azure/)
- [Go client](https://github.com/Azure/azure-sdk-for-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/azure/azure.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
