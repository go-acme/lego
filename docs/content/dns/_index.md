---
title: "DNS Providers"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 3
---

{{% notice important %}}
lego is an independent, free, and open-source project, if you value it, consider [supporting it](https://donate.ldez.dev)! ❤️

This project is not owned by a company. I'm not an employee of a company.

I don't have gifted domains/accounts from DNS companies.

I've been maintaining it for about 10 years.
{{% /notice %}}

## Configuration and Credentials

Credentials and DNS configuration for DNS providers must be passed through environment variables.

### Environment Variables: Value

The environment variables can reference a value.

Here is an example bash command using the Cloudflare DNS provider:

```bash
$ CLOUDFLARE_EMAIL=you@example.com \
  CLOUDFLARE_API_KEY=b9841238feb177a84330febba8a83208921177bffe733 \
  lego --dns cloudflare --domains www.example.com --email you@example.com run
```

### Environment Variables: File

The environment variables can reference a path to file.

In this case the name of environment variable must be suffixed by `_FILE`.

{{% notice note %}}
The file must contain only the value.
{{% /notice %}}

Here is an example bash command using the CloudFlare DNS provider:

```bash
$ cat /the/path/to/my/key
b9841238feb177a84330febba8a83208921177bffe733

$ cat /the/path/to/my/email
you@example.com

$ CLOUDFLARE_EMAIL_FILE=/the/path/to/my/email \
  CLOUDFLARE_API_KEY_FILE=/the/path/to/my/key \
  lego --dns cloudflare --domains www.example.com --email you@example.com run
```

## DNS Providers

{{% tableofdnsproviders %}}
