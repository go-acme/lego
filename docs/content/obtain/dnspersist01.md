---
title: "DNS-PERSIST-01 Challenge"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 4
---

This guide explains how to get and renew a certificate with the DNS-PERSIST-01 challenge.

<!--more-->

{{< tabs groupid="usage-examples" >}}
{{% tab title="Classic Way" %}}

Execute the following command:

```bash
lego run -d 'example.com' --dns-persist
```

{{% /tab %}}
{{% tab title="With a Configuration File" %}}

Create a `.lego.yml` file with the following content:

```yaml
certificates:
  foo:
    challenge: dns-persist-01
    domains:
      - example.com
```

And execute:

```bash
lego
```

{{% /tab %}}
{{< /tabs >}}
