---
title: "Designate DNSaaS for Openstack"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: designate
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/designate/designate.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v2.2.0

Configuration for [Designate DNSaaS for Openstack](https://docs.openstack.org/designate/latest/).


<!--more-->

- Code: `designate`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `OS_AUTH_URL` | Identity endpoint URL |
| `OS_PASSWORD` | Password |
| `OS_REGION_NAME` | Region name |
| `OS_TENANT_NAME` | Tenant name |
| `OS_USERNAME` | Username |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DESIGNATE_POLLING_INTERVAL` | Time between DNS propagation check |
| `DESIGNATE_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DESIGNATE_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://docs.openstack.org/designate/latest/)
- [Go client](https://godoc.org/github.com/gophercloud/gophercloud/openstack/dns/v2)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/designate/designate.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
