# Netnod Primary DNS API - API Documentation

## Overview

The following documentation describes the Netnod Primary DNS API. This API is used to manage DNS zones and resource record sets.

**Base URL**: `https://primarydnsapi.netnod.se`

---

## Authentication

### Token-Based Authentication

This API uses the same token-based authentication mechanism as the Netnod Secondary DNS API. Use the customer portal to generate or obtain a token for your account and configure valid IP ranges for your account.

**Header Format**:

Authorization: Token <YOUR_TOKEN>

---

## API Endpoints

### Zone Management Endpoints (Auth Required)

All `/api/v1/*` endpoints require authentication.

#### `GET /api/v1/zones`

Return a paginated list of available zones.

**Query Parameters**:

| Parameter     | Type    | Description                       | Default | Validation |
|:--------------|:--------|:----------------------------------|:--------|:-----------|
| `offset`      | integer | Pagination offset                 | `0`     | >= 0       |
| `limit`       | integer | Results per page                  | `100`   | 1-1000     |
| `endcustomer` | string  | Filter zones by endcustomer value | -       | -          |

**Request Example**:

```sh
curl -X GET "https://primarydnsapi.netnod.se/api/v1/zones?offset=0&limit=50" \
  -H "Authorization: Token <YOUR_TOKEN>"
```

**Request Example with end customer filter**:


```sh
curl -X GET "https://primarydnsapi.netnod.se/api/v1/zones?endcustomer=customer123" \
  -H "Authorization: Token <YOUR_TOKEN>"
```

**Response**: `200 OK`

```json
{
  "data": [
    {
      "id": "example.com.",
      "name": "example.com.",
      "notified_serial": 2025110401
    },
    {
      "id": "test.com.",
      "name": "test.com.",
      "notified_serial": 2025110402
    }
  ],
  "offset": 0,
  "limit": 50,
  "total": 2
}
```

**Error Responses**:

- `400 Bad Request`: Invalid offset/limit parameters  
- `401 Unauthorized`: Missing or invalid token  
- `500 Internal Server Error`: Request processing failed

---

#### `POST /api/v1/zones`

Create new DNS zone.

**Allowed Request Body Fields** (JSON):

- `name`: Zone name (required)  
- `rrsets`: Array of resource record sets (optional)  
- `zone`: BIND style zone file (optional)  
- `also_notify`: Array of IP addresses to send DNS NOTIFY to (optional). Netnod secondary service is always notified, no need to explicitly notify it here.  
- `allow_transfer_keys`: Array of TSIG key IDs that are allowed to perform zone transfers from the distribution servers (optional). This list should include key IDs for all keys used for DNS secondary replication (including the Netnod DNS secondary service)  
- `endcustomer`: Customer identifier string for internal organization or tracking purposes (optional)

| Distribution servers                   | IPV4           | IPV6               |
|:---------------------------------------|:---------------|:-------------------|
| primary-distmaster-stb.sth.dnsnode.net | 192.36.144.195 | 2a01:3f0:0:17::195 |
| primary-distmaster-stc.sth.dnsnode.net | 192.36.144.198 | 2a01:3f0:0:17::198 |

**Request Example with RRsets in JSON format**:

```sh
curl -X POST "https://primarydnsapi.netnod.se/api/v1/zones" \
  -H "Authorization: Token <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "example.com.",
    "also_notify": ["1.2.3.4", "1.2.3.5"],
    "allow_transfer_keys": ["netnod-something-key1."],
    "rrsets": [
      {
        "name": "example.com.",
        "type": "SOA",
        "records": [
          {
            "content": "ns1.example.com. hostmaster.example.com. 2025110401 10800 3600 604800 3600"
          }
        ]
      },
      {
        "name": "example.com.",
        "type": "NS",
        "records": [
          {
            "content": "ns1.example.com."
          }
        ]
      }
    ]
  }'
```

**Response**: `201 Created`

```json
{
  "id": "example.com.",
  "name": "example.com.",
  "notified_serial": 1,
  "also_notify": ["1.2.3.4", "1.2.3.5"],
  "allow_transfer_keys": ["netnod-something-key1."],
  "rrsets": [
    {
      "name": "example.com.",
      "type": "SOA",
      "records": [
        {
          "content": "ns1.example.com. hostmaster.example.com. 2025110401 10800 3600 604800 3600",
          "disabled": false
        }
      ]
    },
    {
      "name": "example.com.",
      "type": "NS",
      "ttl": 3600,
      "records": [
        {
          "content": "ns1.example.com",
          "disabled": false
        }
      ]
    }
  ]
}
```

**Request Example with RRsets in BIND format**:

```sh
curl -X POST "https://primarydnsapi.netnod.se/api/v1/zones" \
  -H "Authorization: Token <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "example.com.",
    "also_notify": ["1.2.3.4", "1.2.3.5"],
    "allow_transfer_keys": ["netnod-something-key1."],
    "zone": "example.com.	3600	IN	SOA	ns1.example.com. hostmaster.example.com. 2025110401 10800 3600 604800 3600\nexample.com. 17800 IN NS ns1.example.com."
  }'
```

**Response**: `201 Created`

```json
{
  "id": "example.com.",
  "name": "example.com.",
  "notified_serial": 1,
  "also_notify": ["1.2.3.4", "1.2.3.5"],
  "allow_transfer_keys": ["netnod-something-key1."],
  "rrsets": [
    {
      "name": "example.com.",
      "type": "SOA",
      "records": [
        {
          "content": "ns1.example.com. hostmaster.example.com. 2025110401 10800 3600 604800 3600",
          "disabled": false
        }
      ]
    },
    {
      "name": "example.com.",
      "type": "NS",
      "ttl": 17800,
      "records": [
        {
          "content": "ns1.example.com.",
          "disabled": false
        }
      ]
    }
  ]
}
```

**Default SOA and NS resource records**

The following default SOA and NS resource records will be added to the zone if no records of the corresponding types are included in the request body:

- SOA nna.netnod.se. noc.netnod.se. 0 7200 1800 604800 7200  
- NS nsu.dnsnode.net.  
- NS nsp.dnsnode.net.

**Error Responses**:

- `400 Bad Request`: Invalid JSON or disallowed fields  
- `401 Unauthorized`: Missing or invalid token  
- `500 Internal Server Error`: Request processing failed

---

#### `GET /api/v1/zones/{zoneId}`

Get detailed information about a specific zone including all resource record sets.

**Path Parameters**:

- `zoneId`: Zone name (e.g., `example.com.`)

**Query Parameters**: None

**Request Example**:

```sh
curl -X GET "https://primarydnsapi.netnod.se/api/v1/zones/example.com." \
  -H "Authorization: Token <YOUR_TOKEN>"

```
**Response**: `200 OK`

```json
{
  "id": "example.com.",
  "name": "example.com.",
  "notified_serial": 2025110401,
  "also_notify": ["1.2.3.4"],
  "allow_transfer_keys": ["netnod-something-key1.", "netnod-something-key2."],
  "endcustomer": "customer123",
  "rrsets": [
    {
      "name": "example.com.",
      "type": "SOA",
      "ttl": 3600,
      "records": [
        {
          "content": "ns1.example.com. hostmaster.example.com. 2025110401 10800 3600 604800 3600",
          "disabled": false
        }
      ]
    },
    {
      "name": "example.com.",
      "type": "NS",
      "ttl": 3600,
      "records": [
        {
          "content": "ns1.example.com.",
          "disabled": false
        },
        {
          "content": "ns2.example.com.",
          "disabled": false
        }
      ]
    }
  ]
}
```

**Response Fields**:

| Field                 | Type    | Description                              |
|:----------------------|:--------|:-----------------------------------------|
| `id`                  | string  | Zone identifier (same as name)           |
| `name`                | string  | Zone name with trailing dot              |
| `notified_serial`     | integer | Last notified SOA serial number          |
| `also_notify`         | array   | IP addresses to notify on zone changes   |
| `allow_transfer_keys` | array   | TSIG key IDs allowed for zone transfers  |
| `endcustomer`         | string  | Customer identifier (omitted if not set) |
| `rrsets`              | array   | All resource record sets in the zone     |

**RRset Fields**:

| Field     | Type    | Description                                 |
|:----------|:--------|:--------------------------------------------|
| `name`    | string  | Record name with trailing dot               |
| `type`    | string  | Record type (A, AAAA, CNAME, MX, TXT, etc.) |
| `ttl`     | integer | Time-to-live in seconds (nullable)          |
| `records` | array   | Individual records in this RRset            |

**Record Fields**:

| Field      | Type    | Description                                     |
|:-----------|:--------|:------------------------------------------------|
| `content`  | string  | Record data (for MX records, includes priority) |
| `disabled` | boolean | Whether record is disabled                      |

**Notes**:

- MX records include priority in content field (e.g., "10 mx1.example.com.")

**Error Responses**:

- `400 Bad Request`: Missing zone ID  
- `401 Unauthorized`: Missing or invalid token  
- `403 Forbidden`: User doesn't have access to this zone  
- `404 Not Found`: Zone doesn't exist  
- `500 Internal Server Error`: Request processing failed

---

#### `GET /api/v1/zones/{zoneId}/export`

Export zone in BIND zone file format.

**Path Parameters**:

- `zoneId`: Zone name (e.g., `example.com.`)

**Query Parameters**: None

**Request Example**:

```sh
curl -X GET "https://primarydnsapi.netnod.se/api/v1/zones/example.com./export" \
  -H "Authorization: Token <YOUR_TOKEN>"
```

**Response**: `200 OK`

```
example.com.	3600	IN	SOA	ns1.example.com. hostmaster.example.com. 2025110401 10800 3600 604800 3600
example.com.	3600	IN	NS	ns1.example.com.
example.com.	3600	IN	NS	ns2.example.com.
mail.example.com.	3600	IN	MX	10 mx1.example.com.
```

**Content-Type**: `text/plain`

**Error Responses**:

- `401 Unauthorized`: Missing or invalid token  
- `403 Forbidden`: User doesn't have access to this zone  
- `404 Not Found`: Zone doesn't exist  
- `500 Internal Server Error`: Request processing failed

---

#### `PUT /api/v1/zones/{zoneId}/notify`

The Netnod secondary DNS service as well as configured secondary nameservers are notified about modified zones every 15 seconds. This endpoint can be used to trigger an immediate DNS NOTIFY.

**Path Parameters**:

- `zoneId`: Zone name (e.g., `example.com.`)

**Request Example**:

```sh
curl -X PUT "https://primarydnsapi.netnod.se/api/v1/zones/example.com./notify" \
  -H "Authorization: Token <YOUR_TOKEN>"
```

**Response**: `200 OK`

```json
{
  "result": "Notification queued"
}
```

**Error Responses**:

- `401 Unauthorized`: Missing or invalid token  
- `403 Forbidden`: User doesn't have access to this zone  
- `404 Not Found`: Zone doesn't exist  
- `500 Internal Server Error`: Request processing failed

---

#### `PUT /api/v1/zones/{zoneId}`

Update zone configuration. Fields included in the request body override existing values.

**Path Parameters**:

- `zoneId`: Zone name (e.g., `example.com.`)

**Allowed Request Body Fields** (JSON):

- `rrsets`: Array of resource record sets (optional)  
- `also_notify`: Array of IP addresses to notify (optional)  
- `allow_transfer_keys`: Array of TSIG key IDs for zone transfers (optional)  
- `endcustomer`: Customer identifier string for internal organization or tracking purposes (optional)

**Request Example**:

```sh
curl -X PUT "https://primarydnsapi.netnod.se/api/v1/zones/example.com." \
  -H "Authorization: Token <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "also_notify": ["1.2.3.4", "1.2.3.5"],
    "allow_transfer_keys": ["netnod-something-key1."],
    "rrsets": [
      {
        "name": "www.example.com.",
        "type": "A",
        "changetype": "REPLACE",
        "records": [
          {
            "content": "1.2.3.4",
            "disabled": false
          }
        ]
      }
    ]
  }'
```

**Response**: `204 No Content` (success)

**Error Responses**:

- `400 Bad Request`: Invalid JSON, disallowed fields, or invalid query parameters  
- `401 Unauthorized`: Missing or invalid token  
- `403 Forbidden`: User doesn't have access to this zone  
- `404 Not Found`: Zone doesn't exist  
- `500 Internal Server Error`: Request processing failed

---

#### `PATCH /api/v1/zones/{zoneId}`

Partially update a zones resource record sets. Individual records can be added, removed or updated. `changetype` must be one of the following:

| Value     | Description                                                     |
|:----------|:----------------------------------------------------------------|
| `REPLACE` | Replace all records in this RRset with the provided ones        |
| `DELETE`  | Remove all records in this RRset                                |
| `EXTEND`  | Add new records to the end of this RRset if not already present |
| `PRUNE`   | Remove the specified record from the RRset if present           |

**Path Parameters**:

- `zoneId`: Zone name (e.g., `example.com.`)

**Allowed Request Body Fields** (JSON):

- `rrsets`: Array of resource record sets with changetype

**Request Example**:

```sh
curl -X PATCH "https://primarydnsapi.netnod.se/api/v1/zones/example.com." \
  -H "Authorization: Token <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "rrsets": [
      {
        "name": "subdomain.example.com.",
        "type": "A",
        "changetype": "DELETE"
      },
      {
        "name": "newhost.example.com.",
        "type": "A",
        "changetype": "REPLACE",
        "records": [
          {
            "content": "1.2.3.5",
            "disabled": false
          }
        ]
      }
    ]
  }'
```

**Response**: `204 No Content` (success)

**Error Responses**:

- `400 Bad Request`: Invalid JSON, disallowed fields, or invalid query parameters  
- `401 Unauthorized`: Missing or invalid token  
- `403 Forbidden`: User doesn't have access to this zone  
- `404 Not Found`: Zone doesn't exist  
- `500 Internal Server Error`: Request processing failed

---

#### `DELETE /api/v1/zones/{zoneId}`

Delete a zone entirely.

**Path Parameters**:

- `zoneId`: Zone name (e.g., `example.com.`)

**Request Example**:

```sh
curl -X DELETE "https://primarydnsapi.netnod.se/api/v1/zones/example.com." \
  -H "Authorization: Token <YOUR_TOKEN>"
```

**Response**: `204 No Content` (success)

**Error Responses**:

- `400 Bad Request`: Request body not empty or invalid query parameters  
- `401 Unauthorized`: Missing or invalid token  
- `403 Forbidden`: User doesn't have access to this zone  
- `404 Not Found`: Zone doesn't exist  
- `500 Internal Server Error`: Request processing failed

---

### DynDNS Management Endpoints (Auth Required)

These endpoints manage Dynamic DNS for a zone. Once enabled for a label, clients can update the DNS record using the public `/dyndns/update` endpoint.

#### `GET /api/v1/zones/{zoneId}/dyndns`

List all DynDNS-enabled labels for a zone.

**Path Parameters**:

- `zoneId`: Zone name (e.g., `example.com.`)

**Request Example**:

```sh
curl -X GET "https://primarydnsapi.netnod.se/api/v1/zones/example.com./dyndns" \
  -H "Authorization: Token <YOUR_TOKEN>"
```

**Response**: `200 OK`

```json
{
  "labels": [
    {
      "label": "home",
      "hostname": "home.example.com."
    },
    {
      "label": "office",
      "hostname": "office.example.com."
    }
  ]
}
```

---

#### `POST /api/v1/zones/{zoneId}/dyndns/{label}`

Enable DynDNS for a label. Returns a token that must be used for updates.

**Path Parameters**:

- `zoneId`: Zone name (e.g., `example.com.`)  
- `label`: DNS label (e.g., `home`)

**Request Example**:

```sh
curl -X POST "https://primarydnsapi.netnod.se/api/v1/zones/example.com./dyndns/home" \
  -H "Authorization: Token <YOUR_TOKEN>"
```

**Response**: `201 Created`

```json
{
  "hostname": "home.example.com.",
  "token": "a1b2c3d4e5f6..."
}
```

**Important**: Save the token securely. It cannot be retrieved later.

**Error Responses**:

- `400 Bad Request`: Invalid label format  
- `409 Conflict`: DynDNS already enabled for this label

---

#### `DELETE /api/v1/zones/{zoneId}/dyndns/{label}`

Disable DynDNS for a label. The token is invalidated.

**Path Parameters**:

- `zoneId`: Zone name (e.g., `example.com.`)  
- `label`: DNS label (e.g., `home`)

**Request Example**:

```sh
curl -X DELETE "https://primarydnsapi.netnod.se/api/v1/zones/example.com./dyndns/home" \
  -H "Authorization: Token <YOUR_TOKEN>"
```

**Response**: `204 No Content`

**Error Responses**:

- `404 Not Found`: DynDNS not enabled for this label

---

### DynDNS Public Update Endpoint

#### `GET /dyndns/update`

Update DNS A/AAAA records for a DynDNS-enabled hostname.

**Authentication**:

Authentication is via the `Authorization` header using either Basic auth or Bearer token:

| Method     | Header Format                             | Description                                |
|:-----------|:------------------------------------------|:-------------------------------------------|
| Basic auth | `Authorization: Basic base64(user:token)` | Username is ignored, password is the token |
| Bearer     | `Authorization: Bearer <token>`           | Token directly in header                   |

Basic auth is recommended as it's supported by most DynDNS clients (ddclient, Fritz\!Box, routers, etc.).

**Query Parameters**:

| Parameter  | Type   | Required | Description                                    |
|:-----------|:-------|:---------|:-----------------------------------------------|
| `hostname` | string | Yes      | Full hostname (e.g., `home.example.com`)       |
| `myip`     | string | No       | IP address(es), comma-separated for dual-stack |

If `myip` is omitted, the client's IP address is auto-detected.

**Request Examples**:

* Basic auth (recommended for DynDNS clients)

```sh
curl -u "dyndns:YOUR_TOKEN" "https://primarydnsapi.netnod.se/dyndns/update?hostname=home.example.com"
```

* Bearer token

```sh
curl -H "Authorization: Bearer YOUR_TOKEN" "https://primarydnsapi.netnod.se/dyndns/update?hostname=home.example.com"
```

* With explicit IPv4

```sh
curl -u "dyndns:YOUR_TOKEN" "https://primarydnsapi.netnod.se/dyndns/update?hostname=home.example.com&myip=1.2.3.4"
```

* Dual-stack (IPv4 \+ IPv6)

```sh
curl -u "dyndns:YOUR_TOKEN" "https://primarydnsapi.netnod.se/dyndns/update?hostname=home.example.com&myip=1.2.3.4,2001:db8::1"
```

**Response**: `200 OK`

```json
{
  "status": "updated",
  "ipv4": "1.2.3.4",
  "ipv6": "2001:db8::1"
}
```

**Error Responses**:

| Code | Error                     | Description                     |
|:-----|:--------------------------|:--------------------------------|
| 400  | `missing_hostname`        | hostname parameter required     |
| 400  | `invalid_ip_address`      | myip contains invalid IP        |
| 401  | `missing_credentials`     | Authorization header missing    |
| 401  | `invalid_token`           | Token doesn't match             |
| 404  | `hostname_not_configured` | DynDNS not enabled for hostname |

---

### ACME DNS-01 Challenge Management Endpoints (Auth Required)

These endpoints manage ACME DNS-01 challenge automation for a zone. Once enabled for a label, ACME clients can set challenge TXT records using the public `/acme/update` endpoint.

#### `GET /api/v1/zones/{zoneId}/acme`

List all ACME-enabled labels for a zone.

**Path Parameters**:

- `zoneId`: Zone name (e.g., `example.com.`)

**Request Example**:

```sh
curl -X GET "https://primarydnsapi.netnod.se/api/v1/zones/example.com./acme" \
  -H "Authorization: Token <YOUR_TOKEN>"
```

**Response**: `200 OK`

```json
{
  "labels": [
    {
      "label": "www",
      "hostname": "www.example.com.",
      "challenge_hostname": "_acme-challenge.www.example.com."
    },
    {
      "label": "@",
      "hostname": "example.com.",
      "challenge_hostname": "_acme-challenge.example.com."
    }
  ]
}
```

---

#### `POST /api/v1/zones/{zoneId}/acme/{label}`

Enable ACME DNS-01 challenges for a label. Use `@` for the zone apex.

**Path Parameters**:

- `zoneId`: Zone name (e.g., `example.com.`)  
- `label`: DNS label (e.g., `www` or `@` for apex)

**Request Example**:

```sh
curl -X POST "https://primarydnsapi.netnod.se/api/v1/zones/example.com./acme/www" \
  -H "Authorization: Token <YOUR_TOKEN>"
```

**Response**: `201 Created`

```json
{
  "hostname": "www.example.com.",
  "challenge_hostname": "_acme-challenge.www.example.com.",
  "token": "a1b2c3d4e5f6..."
}
```

**Important**: Save the token securely. It cannot be retrieved later.

**Error Responses**:

- `400 Bad Request`: Invalid label format  
- `409 Conflict`: ACME already enabled for this label

---

#### `DELETE /api/v1/zones/{zoneId}/acme/{label}`

Disable ACME for a label. Also removes any existing challenge TXT record.

**Path Parameters**:

- `zoneId`: Zone name (e.g., `example.com.`)  
- `label`: DNS label (e.g., `www` or `@`)

**Request Example**:

```sh
curl -X DELETE "https://primarydnsapi.netnod.se/api/v1/zones/example.com./acme/www" \
  -H "Authorization: Token <YOUR_TOKEN>"
```

**Response**: `204 No Content`

**Error Responses**:

- `404 Not Found`: ACME not enabled for this label

---

### ACME Public Update Endpoints

**Authentication** (applies to both endpoints):

Authentication is via the `Authorization` header using either Basic auth or Bearer token:

| Method     | Header Format                             | Description                                |
|:-----------|:------------------------------------------|:-------------------------------------------|
| Basic auth | `Authorization: Basic base64(user:token)` | Username is ignored, password is the token |
| Bearer     | `Authorization: Bearer <token>`           | Token directly in header                   |

#### `GET /acme/update`

Set ACME DNS-01 challenge TXT record.

**Query Parameters**:

| Parameter  | Type   | Required | Description                             |
|:-----------|:-------|:---------|:----------------------------------------|
| `hostname` | string | Yes      | Full hostname (e.g., `www.example.com`) |
| `txt`      | string | Yes      | ACME challenge value                    |

**Request Examples**:

* Basic auth

```sh
curl -u "acme:YOUR_TOKEN" "https://primarydnsapi.netnod.se/acme/update?hostname=www.example.com&txt=challenge_value"
```

* Bearer token

```sh
curl -H "Authorization: Bearer YOUR_TOKEN" "https://primarydnsapi.netnod.se/acme/update?hostname=www.example.com&txt=challenge_value"
```

**Response**: `200 OK`

```json
{
  "status": "updated",
  "txt": "challenge_value"
}
```

---

#### `DELETE /acme/update`

Remove ACME DNS-01 challenge TXT record after validation completes.

**Query Parameters**:

| Parameter  | Type   | Required | Description                             |
|:-----------|:-------|:---------|:----------------------------------------|
| `hostname` | string | Yes      | Full hostname (e.g., `www.example.com`) |

**Request Examples**:

* Basic auth

```sh
curl -u "acme:YOUR_TOKEN" -X DELETE "https://primarydnsapi.netnod.se/acme/update?hostname=www.example.com"
```

* Bearer token

```sh
curl -H "Authorization: Bearer YOUR_TOKEN" -X DELETE "https://primarydnsapi.netnod.se/acme/update?hostname=www.example.com"
```

**Response**: `200 OK`

```json
{
  "status": "deleted"
}
```

**Error Responses** (both endpoints):

| Code | Error                     | Description                       |
|:-----|:--------------------------|:----------------------------------|
| 400  | `missing_hostname`        | hostname parameter required       |
| 400  | `missing_txt`             | txt parameter required (GET only) |
| 401  | `missing_credentials`     | Authorization header missing      |
| 401  | `invalid_token`           | Token doesn't match               |
| 404  | `hostname_not_configured` | ACME not enabled for hostname     |

---

## Rate Limiting

- Implementation: IP-based rate limiting using `httprate`
- Default Limit: 100 requests per minute per IP address
- Configuration: `RATE_LIMIT_RPM` environment variable
- Scope: Applied globally to all endpoints

Rate Limit Exceeded Response: `429 Too Many Requests`

---

## Error Response Format

All error responses follow a consistent JSON format:

```json
{
  "error": "Error message description"
}
```

### Common HTTP Status Codes

| Code  | Description                              |
|:------|:-----------------------------------------|
| `200` | Successful request with response body    |
| `204` | Successful request without response body |
| `400` | Invalid request parameters or body       |
| `401` | Missing or invalid authentication        |
| `403` | Valid auth but insufficient permissions  |
| `404` | Resource doesn't exist                   |
| `429` | Rate limit exceeded                      |
| `500` | Server-side error                        |

### Supported record types

- A  
- AAAA  
- ALIAS  
- CAA  
- CERT  
- CNAME  
- CSYNC  
- HTTPS  
- LOC  
- PTR  
- SOA  
- SPF  
- SRV  
- NS  
- MX  
- TLSA  
- TXT

## Example Workflows

### 1. List Available Zones

```sh
curl -X GET "https://primarydnsapi.netnod.se/api/v1/zones" \
  -H "Authorization: Token YOUR_TOKEN"
```

### 2. Create New Zone

```sh
curl -X POST "https://primarydnsapi.netnod.se/api/v1/zones" \
  -H "Authorization: Token YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "example.com.",
    "also_notify": ["1.2.3.4"],
    "allow_transfer_keys": ["netnod-something-key1."]
  }'
```

### 3. View Zone Details

```sh
curl -X GET "https://primarydnsapi.netnod.se/api/v1/zones/example.com." \
  -H "Authorization: Token YOUR_TOKEN"
```

### 4. Add DNS Record

```sh
curl -X PATCH "https://primarydnsapi.netnod.se/api/v1/zones/example.com." \
  -H "Authorization: Token YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "rrsets": [
      {
        "name": "api.example.com.",
        "type": "A",
        "changetype": "REPLACE",
        "ttl": 3600,
        "records": [
          {"content": "127.0.0.1", "disabled": false}
        ]
      }
    ]
  }'
```

### 5. Delete DNS Record

```sh
curl -X PATCH "https://primarydnsapi.netnod.se/api/v1/zones/example.com." \
  -H "Authorization: Token YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "rrsets": [
      {
        "name": "old.example.com.",
        "type": "A",
        "changetype": "DELETE"
      }
    ]
  }'
```

### 6. Update Zone Metadata

```sh
curl -X PUT "https://primarydnsapi.netnod.se/api/v1/zones/example.com." \
  -H "Authorization: Token YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "also_notify": ["1.2.3.4", "1.2.3.5"],
    "allow_transfer_keys": ["netnod-something-key1.", "netnod-something-key2."]
  }'
```

### 7. Export Zone

```sh
curl -X GET "https://primarydnsapi.netnod.se/api/v1/zones/example.com./export" \
  -H "Authorization: Token YOUR_TOKEN"
```

### 8. Trigger Zone Notify

```sh
curl -X PUT "https://primarydnsapi.netnod.se/api/v1/zones/example.com./notify" \
  -H "Authorization: Token YOUR_TOKEN"
```

---

## Troubleshooting

### Common Issues

**401 Unauthorized**

- Verify "Authorization: Token <YOUR_TOKEN>" header format  
- Verify that the NDS token is valid using the customer portal  
- Check that the request originates from Check client IP is within customer's configured prefix range  
- Ensure that your organisation has been onboarded to the primary dns service

**403 Forbidden**

- Verify that the referenced ZoneId belongs to your organisation

**429 Too Many Requests**

- Reduce request rate or contact support\`  
- Implement client-side rate limiting/backoff

