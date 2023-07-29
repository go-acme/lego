---
title: "IBM Cloud (SoftLayer)"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: ibmcloud
dnsprovider:
  since:    "v4.5.0"
  code:     "ibmcloud"
  url:      "https://www.ibm.com/cloud/"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ibmcloud/ibmcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [IBM Cloud (SoftLayer)](https://www.ibm.com/cloud/).


<!--more-->

- Code: `ibmcloud`
- Since: v4.5.0


Here is an example bash command using the IBM Cloud (SoftLayer) provider:

```bash
SOFTLAYER_USERNAME=xxxxx \
SOFTLAYER_API_KEY=yyyyy \
lego --email you@example.com --dns ibmcloud --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `SOFTLAYER_API_KEY` | Classic Infrastructure API key |
| `SOFTLAYER_USERNAME` | Username (IBM Cloud is <accountID>_<emailAddress>) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `SOFTLAYER_POLLING_INTERVAL` | Time between DNS propagation check |
| `SOFTLAYER_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `SOFTLAYER_TIMEOUT` | API request timeout |
| `SOFTLAYER_TTL` | The TTL of the TXT record used for the DNS challenge |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{< ref "dns#configuration-and-credentials" >}}).




## More information

- [API documentation](https://cloud.ibm.com/docs/dns?topic=dns-getting-started-with-the-dns-api)
- [Go client](https://github.com/softlayer/softlayer-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/ibmcloud/ibmcloud.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
