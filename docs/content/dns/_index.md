---
title: "DNS Providers"
date: 2019-03-03T16:39:46+01:00
draft: false
weight: 3
---

> [!INFO] Important
> lego is an independent, free, and open-source project, if you value it, consider [supporting it](https://donate.ldez.dev)! ❤️

## Configuration and Credentials

Credentials and DNS configuration for DNS providers must be passed through environment variables.

### Environment Variables

The environment variables can reference a value.

```bash
CLOUDFLARE_EMAIL='you@example.com' \
CLOUDFLARE_API_KEY='yourprivatecloudflareapikey' \
...
```

#### `_FILE` suffix

The environment variables can reference a path to a file.

In this case the name of environment variable must be suffixed by `_FILE`.

{{< tabs >}}
{{% tab title="Command" %}}

```bash
CLOUDFLARE_EMAIL_FILE=/the/path/to/my/email \
CLOUDFLARE_API_KEY_FILE=/the/path/to/my/key \
lego run --dns cloudflare --domains www.example.com
```

{{% /tab %}}
{{% tab title="/the/path/to/my/key" %}}

```
yourprivatecloudflareapikey
```

{{% /tab %}}
{{% tab title="/the/path/to/my/email" %}}

```
you@example.com
```

{{% /tab %}}
{{< /tabs >}}

{{% notice note %}}
The file must contain only the value.
{{% /notice %}}

### Dotenv File

You can also use a dotenv file.

When using `lego run`, you can pass the path to the dotenv file with the `--env-file` flag.

{{< tabs >}}
{{% tab title="Command" %}}

```bash
lego run --dns cloudflare --domains 'example.org' --domains '*.example.org' --env-file .env.cf
```

{{% /tab %}}
{{% tab title=".env.cf" %}}

```ini
CLOUDFLARE_EMAIL=you@example.com
CLOUDFLARE_API_KEY=yourprivatecloudflareapikey
```

{{% /tab %}}
{{< /tabs >}}

When using `lego`, the environment variables are loaded from the file defined by `envFile` in the configuration file for the DNS provider.

{{< tabs >}}
{{% tab title=".lego.yml" %}}

```yaml
challenges:
  cf:
    dns:
      provider: cloudflare
      envFile: .env.cf

certificates:
  foo:
    domains:
      - example.com
      - '*.example.com'
```

{{% /tab %}}
{{% tab title=".env.cf" %}}

```ini
CLOUDFLARE_EMAIL=you@example.com
CLOUDFLARE_API_KEY=yourprivatecloudflareapikey
```

{{% /tab %}}
{{< /tabs >}}

## DNS Providers

{{% tableofdnsproviders %}}
