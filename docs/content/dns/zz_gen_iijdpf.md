---
title: "IIJ DNS Platform Service"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: iijdpf
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/iijdpf/iijdpf.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

Since: v4.7.0

Configuration for [IIJ DNS Platform Service](https://www.iij.ad.jp/en/biz/dns-pfm/).


<!--more-->

- Code: `iijdpf`

Here is an example bash command using the IIJ DNS Platform Service provider:

```bash
IIJ_DPF_API_TOKEN=xxxxxxxx \
IIJ_DPF_DPM_SERVICE_CODE=yyyyyy \
lego -email myemail@example.com --dns iijdpf --domains my.example.org run
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `IIJ_DPF_API_TOKEN` | API token |
| `IIJ_DPF_DPM_SERVICE_CODE` | IIJ Managed DNS Service's service code |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `IIJ_DPF_API_ENDPOINT` | API endpoint URL, defaults to https://api.dns-platform.jp/dpf/v1 |
| `IIJ_DPF_POLLING_INTERVAL` | Time between DNS propagation check, defaults to 5 second |
| `IIJ_DPF_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation, defaults to 660 second |
| `IIJ_DPF_TTL` | The TTL of the TXT record used for the DNS challenge, default to 300 |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here](/lego/dns/#configuration-and-credentials).




## More information

- [API documentation](https://manual.iij.jp/dpf/dpfapi/)
- [Go client](https://github.com/mimuret/golang-iij-dpf)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/iijdpf/iijdpf.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
