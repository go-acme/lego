# Hetzner Cloud DNS Provider

The `hetznerhcloud` provider supports the **new** Hetzner Cloud DNS API (`api.hetzner.cloud`) for ACME DNS-01 challenges.

## Differences from `hetzner` Provider

| Feature | `hetzner` (existing) | `hetznerhcloud` (this) |
|---------|---------------------|------------------------|
| API Endpoint | dns.hetzner.com | api.hetzner.cloud |
| Token Type | HETZNER_DNS_TOKEN | HCLOUD_TOKEN |
| API Version | Legacy DNS API | Cloud API v1 |
| Integration | DNS-only | Full Hetzner Cloud |

## Configuration

### Environment Variables

```bash
# Required
export HCLOUD_TOKEN="your-hetzner-cloud-api-token"

# Optional
export HCLOUD_TTL=60                    # Record TTL in seconds
export HCLOUD_PROPAGATION_TIMEOUT=120   # Max wait time
export HCLOUD_POLLING_INTERVAL=5        # Check interval
export HCLOUD_HTTP_TIMEOUT=30           # HTTP timeout
export HCLOUD_BASE_URL="https://api.hetzner.cloud"  # API endpoint
```

### API Token

1. Go to [Hetzner Cloud Console](https://console.hetzner.cloud)
2. Select your project
3. Navigate to **Security** → **API Tokens**
4. Generate token with **Read & Write** permissions

## Usage

### CLI

```bash
# Basic usage
HCLOUD_TOKEN="xxx" lego --dns hetznerhcloud -d example.com run

# Wildcard certificate
HCLOUD_TOKEN="xxx" lego --dns hetznerhcloud -d "*.example.com" -d example.com run

# With custom settings
HCLOUD_TOKEN="xxx" \
HCLOUD_TTL=120 \
HCLOUD_PROPAGATION_TIMEOUT=300 \
lego --dns hetznerhcloud -d example.com run
```

### Docker

```bash
docker run -v $(pwd):/lego \
  -e HCLOUD_TOKEN="xxx" \
  goacme/lego:hcloud \
  --dns hetznerhcloud \
  -d example.com \
  run
```

### Library

```go
import "github.com/go-acme/lego/v4/providers/dns/hetznerhcloud"

config := hetznerhcloud.NewDefaultConfig()
config.Token = "your-token"

provider, err := hetznerhcloud.NewDNSProviderConfig(config)
if err != nil {
    log.Fatal(err)
}

// Use with Lego client
client.Challenge.SetDNS01Provider(provider)
```

## API Details

### Present Method

Creates TXT record for ACME challenge:

```http
POST /v1/zones/{zoneID}/records
{
  "name": "_acme-challenge",
  "type": "TXT",
  "value": "{challenge-value}",
  "ttl": 60
}
```

### CleanUp Method

Removes TXT record after validation:

```http
DELETE /v1/zones/{zoneID}/records/{recordID}
```

## Features

- **Thread-safe**: Concurrent operations supported
- **Retry logic**: Automatic retry with exponential backoff
- **Zone caching**: Reduces API calls
- **Rate limiting**: Respects API limits

## Troubleshooting

### Zone Not Found

```
Error: could not find zone for domain
```

**Solution**: Ensure DNS zone exists in Hetzner Cloud project

### Authentication Failed

```
Error: 401 Unauthorized
```

**Solution**: Check API token has correct permissions

### Propagation Timeout

```
Error: timeout waiting for DNS propagation
```

**Solution**: Increase `HCLOUD_PROPAGATION_TIMEOUT`

### Debug Mode

```bash
LEGO_LOG_LEVEL=debug HCLOUD_TOKEN=xxx lego --dns hetznerhcloud -d example.com run
```

## Integration Examples

### With Traefik

```yaml
# docker-compose.yml
services:
  traefik:
    image: traefik:v3.5
    environment:
      - HCLOUD_TOKEN=${HCLOUD_TOKEN}
    command:
      - --certificatesresolvers.le.acme.dnschallenge.provider=exec
      - --certificatesresolvers.le.acme.dnschallenge.exec=/lego
    volumes:
      - ./lego:/lego:ro  # Mount lego binary
```

### With Certbot

Use Lego to obtain certificates, then configure Certbot to use them:

```bash
# Get certificate with Lego
lego --dns hetznerhcloud -d example.com run

# Use with Certbot
certbot install --cert-path certificates/example.com.crt \
  --key-path certificates/example.com.key
```

### With Kubernetes cert-manager

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: hcloud-token
stringData:
  token: "your-token"
---
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt
spec:
  acme:
    solvers:
    - dns01:
        webhook:
          groupName: acme.example.com
          solverName: hetznerhcloud
          config:
            tokenSecretRef:
              name: hcloud-token
              key: token
```

## Links

- [Hetzner Cloud API Documentation](https://docs.hetzner.cloud/reference/dns)
- [Lego Documentation](https://go-acme.github.io/lego/)
- [ACME Protocol](https://datatracker.ietf.org/doc/html/rfc8555)