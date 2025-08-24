---
title: "Oracle Cloud"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: oraclecloud
dnsprovider:
  since:    "v2.3.0"
  code:     "oraclecloud"
  url:      "https://cloud.oracle.com/home"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/oraclecloud/oraclecloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [Oracle Cloud](https://cloud.oracle.com/home).


<!--more-->

- Code: `oraclecloud`
- Since: v2.3.0


Here is an example bash command using the Oracle Cloud provider:

```bash
# Using API Key authentication:
OCI_PRIVATE_KEY_PATH="~/.oci/oci_api_key.pem" \
OCI_PRIVATE_KEY_PASSWORD="secret" \
OCI_TENANCY_OCID="ocid1.tenancy.oc1..secret" \
OCI_USER_OCID="ocid1.user.oc1..secret" \
OCI_FINGERPRINT="00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00" \
OCI_REGION="us-phoenix-1" \
OCI_COMPARTMENT_OCID="ocid1.tenancy.oc1..secret" \
lego --email you@example.com --dns oraclecloud -d '*.example.com' -d example.com run

# Using Instance Principal authentication (when running on OCI compute instances):
# https://docs.oracle.com/en-us/iaas/Content/Identity/Tasks/callingservicesfrominstances.htm
OCI_AUTH_TYPE="instance_principal" \
OCI_COMPARTMENT_OCID="ocid1.tenancy.oc1..secret" \
lego --email you@example.com --dns oraclecloud -d '*.example.com' -d example.com run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `OCI_COMPARTMENT_OCID` | Compartment OCID |
| `OCI_FINGERPRINT` | Public key fingerprint (ignored if `OCI_AUTH_TYPE=instance_principal`) |
| `OCI_PRIVATE_KEY_PASSWORD` | Private key password (ignored if `OCI_AUTH_TYPE=instance_principal`) |
| `OCI_PRIVATE_KEY_PATH` | Private key file (ignored if `OCI_AUTH_TYPE=instance_principal`) |
| `OCI_REGION` | Region (it can be empty if `OCI_AUTH_TYPE=instance_principal`). |
| `OCI_TENANCY_OCID` | Tenancy OCID (ignored if `OCI_AUTH_TYPE=instance_principal`) |
| `OCI_USER_OCID` | User OCID (ignored if `OCI_AUTH_TYPE=instance_principal`) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `OCI_AUTH_TYPE` | Authorization type. Possible values: 'instance_principal', ''  (Default: '') |
| `OCI_HTTP_TIMEOUT` | API request timeout in seconds (Default: 60) |
| `OCI_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `OCI_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |
| `OCI_TTL` | The TTL of the TXT record used for the DNS challenge in seconds (Default: 120) |
| `TF_VAR_fingerprint` | Alias on `OCI_FINGERPRINT` |
| `TF_VAR_private_key_path` | Alias on `OCI_PRIVATE_KEY_PATH` |
| `TF_VAR_region` | Alias on `OCI_REGION` |
| `TF_VAR_tenancy_ocid` | Alias on `OCI_TENANCY_OCID` |
| `TF_VAR_user_ocid` | Alias on `OCI_USER_OCID` |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).




## More information

- [API documentation](https://docs.cloud.oracle.com/iaas/Content/DNS/Concepts/dnszonemanagement.htm)
- [Go client](https://github.com/oracle/oci-go-sdk)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/oraclecloud/oraclecloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
