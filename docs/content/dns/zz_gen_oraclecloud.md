---
title: "Oracle Cloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: oraclecloud
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/oraclecloud/oraclecloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v2.3.0

Configuration for [Oracle Cloud](https://cloud.oracle.com/home).


<!--more-->

- Code: `oraclecloud`

Here is an example bash command using the Oracle Cloud provider:

```bash
OCI_PRIVKEY_FILE="~/.oci/oci_api_key.pem" \
OCI_PRIVKEY_PASS="secret" \
OCI_TENANCY_OCID="ocid1.tenancy.oc1..secret" \
OCI_USER_OCID="ocid1.user.oc1..secret" \
OCI_PUBKEY_FINGERPRINT="00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00" \
OCI_REGION="us-phoenix-1" \
OCI_COMPARTMENT_OCID="ocid1.tenancy.oc1..secret" \
lego --dns oraclecloud --domains my.domain.com --email my@email.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `OCI_COMPARTMENT_OCID` | Compartment OCID |
| `OCI_PRIVKEY_FILE` | Private key file |
| `OCI_PRIVKEY_PASS` | Private key password |
| `OCI_PUBKEY_FINGERPRINT` | Public key fingerprint |
| `OCI_REGION` | Region |
| `OCI_TENANCY_OCID` | Tenanct OCID |
| `OCI_USER_OCID` | User OCID |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `OCI_POLLING_INTERVAL` | Time between DNS propagation check |
| `OCI_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `OCI_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://docs.cloud.oracle.com/iaas/Content/DNS/Concepts/dnszonemanagement.htm)
- [Go client](https://github.com/oracle/oci-go-sdk)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/oraclecloud/oraclecloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
