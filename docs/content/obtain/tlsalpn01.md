---
title: "TLS-ALPN-01 Challenge"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 1
---

This guide explains how to get and renew a certificate with the TLS-ALPN-01 challenge.

<!--more-->

{{% notice note %}}
The examples require that the `lego` binary has permission to bind to ports 443.  
If your environment does not allow you to bind to these ports, please read [Running without root privileges]({{% ref "advanced/tips#running-without-root-privileges" %}}) and [Port Usage]({{% ref "advanced/tips/#port-usage" %}}).
{{% /notice %}}

{{< tabs groupid="usage-examples" >}}
{{% tab title="Classic Way" %}}

Execute the following command:

```bash
lego run -d 'example.com' --tls
```

{{% /tab %}}
{{% tab title="With a Configuration File" %}}

Create a `.lego.yml` file with the following content:

```yaml
certificates:
  foo:
    challenge: tls-alpn-01
    domains:
      - example.com
```

And execute:

```bash
lego
```

{{% /tab %}}
{{< /tabs >}}
