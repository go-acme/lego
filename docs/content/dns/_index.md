---
title: "DNS Providers"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 3
---

Credentials for DNS providers must be passed through environment variables.

Here is an example bash command using the CloudFlare DNS provider:

```bash
CLOUDFLARE_EMAIL=foo@bar.com \
CLOUDFLARE_API_KEY=b9841238feb177a84330febba8a83208921177bffe733 \
lego --dns cloudflare --domains www.example.com --email me@bar.com run
```

{{%children style="h2" description="true" %}}