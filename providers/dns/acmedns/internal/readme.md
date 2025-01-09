# HTTP Storage

## Fetch

### Request

Endpoint: `GET <BaseURL>/<domain>`

### Response

Response status code 200.

Response body (account):

```json
{
  "fulldomain": "foo.example.com",
  "subdomain": "foo",
  "username": "user",
  "password": "secret",
  "server_url": "https://example.com"
}
```

## Fetch All

### Request

Endpoint: `GET <BaseURL>`

### Response

Response status code 200.

Response body (domain/account mapping):

```json
{
  "foo.example.com": {
    "fulldomain": "foo.example.com",
    "subdomain": "foo",
    "username": "user",
    "password": "secret",
    "server_url": "https://example.com"
  },
  "bar.example.com": {
    "fulldomain": "bar.example.com",
    "subdomain": "bar",
    "username": "user",
    "password": "secret",
    "server_url": "https://example.com"
  }
}
```

## Put

### Request

Endpoint: `POST <BaseURL>/<domain>`

### Response

Response status code 200.

No expected body.

## Save

No dedicated endpoint.
