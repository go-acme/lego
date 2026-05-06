# Memcached HTTP provider

Publishes challenges into memcached where they can be retrieved by nginx. Allows
specifying multiple memcached servers and the responses will be published to all
of them, making it easier to verify when your domain is hosted on a cluster of
servers.

Example nginx config:

```
    location /.well-known/acme-challenge/ {
        set $memcached_key "$uri";
        memcached_pass 127.0.0.1:11211;
    }
```

## Local Development

```bash
docker run -d --rm -p 11211:11211 memcached:alpine
```

```bash
MEMCACHED_HOSTS=localhost:11211
```

```go
	os.Setenv("MEMCACHED_HOSTS", "localhost:11211")
```
