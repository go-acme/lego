---
title: "DNSPod"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dnspod
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnspod/dnspod.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [DNSPod](http://www.dnspod.com/).


<!--more-->

- Code: `dnspod`

{{% notice note %}}
_Please contribute by adding a CLI example._
{{% /notice %}}




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DNSPOD_API_KEY` | The user token |


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DNSPOD_HTTP_TIMEOUT` | API request timeout |
| `DNSPOD_POLLING_INTERVAL` | Time between DNS propagation check |
| `DNSPOD_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation |
| `DNSPOD_TTL` | The TTL of the TXT record used for the DNS challenge |




## More information

- [API documentation](https://www.dnspod.com/docs/index.html)
- [Go client](https://github.com/decker502/dnspod-go)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnspod/dnspod.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
