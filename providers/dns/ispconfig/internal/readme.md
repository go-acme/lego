## Error Response

```json
{
  "code":  <TODO>,
  "message":  <TODO>,
  "response": <TODO>
}
```

## Login Endpoint

* URL: `<server>?login`
* HTTP Method: `POST`

- https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/remoting_client/API-docs/login.html
- https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/remoting_client/examples/login.php

### Request Body (JSON)

```json
{
  "username": "<username>",
  "password": "<password>",
  "client_login": false
}
```

### Response Body (JSON)

```json
{
  "code": "ok",
  "message": "foo",
  "response": <TODO>
}
```

```json
{
  "code": "ok",
  "message": "foo",
  "response": "abc"
}
```

- `response`: is the `sessionID`

## Get Client ID Endpoint

* URL: `<server>?client_get_id`
* HTTP Method: `POST`

- https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/remoting_client/API-docs/client_get_id.html
- https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/remoting_client/examples/client_get_id.php

### Request Body (JSON)

```json
{
  "session_id": "<sessionID>",
  "sys_userid": "<sys_userid>"
}
```

### Response Body (JSON)

```json
{
  "code": "ok",
  "message": "foo",
  "response": <TODO>
}
```

```json
{
  "code": "ok",
  "message": "foo",
  "response": 123
}
```

## DNS Zone Get Endpoint

* URL: `<server>?dns_zone_get`
* HTTP Method: `POST`

- https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/remoting_client/API-docs/dns_zone_get.html
- https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/remoting_client/examples/dns_zone_get.php
- TABLE `dns_soa`: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/install/sql/ispconfig3.sql?ref_type=heads#L615

### Request Body (JSON)

```json
{
  "session_id": "<session_id>",
  "primary_id": "<zone_id>"
}
```

### Response Body (JSON)

```json
{
  "code": "ok",
  "message": "foo",
  "response": <TODO>
}
```

```json
{
  "code": "ok",
  "message": "foo",
  "response": {
    "id": 456,
    "server_id": 123,
    "sys_userid": 789
  }
}
```

## DNS TXT Get Endpoint

* URL: `<server>?dns_txt_get`
* HTTP Method: `POST`

- https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/remoting_client/API-docs/dns_txt_get.html
- https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/remoting_client/examples/dns_txt_get.php

### Request Body (JSON)

```json
{
  "session_id": "<session_id>",
  "primary_id": {
    "name": "<fulldomain>.",
    "type": "TXT"
  }
}
```

### Response Body (JSON)

```json
{
  "code": "ok",
  "message": "foo",
  "response": <TODO>
}
```

```json
{
  "code": "ok",
  "message": "foo",
  "response": {
    "id": 123
  }
}
```

## DNS TXT Add Endpoint

* URL: `<server>?dns_txt_add`
* HTTP Method: `POST`

- https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/remoting_client/API-docs/dns_txt_add.html
- https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/remoting_client/examples/dns_txt_add.php

### Request Body (JSON)

```json
{
  "session_id": "<session_id>",
  "client_id": "<client_id>",
  "params": {
    "server_id": "<server_id>",
    "zone": "<zone>",
    "name": "<fulldomain>.",
    "type": "txt",
    "data": "<txtvalue>",
    "aux": "0",
    "ttl": "3600",
    "active": "y",
    "stamp": "<curStamp>",
    "serial": "<curSerial>"
  },
  "update_serial": true
}
```

- `stamp`: (ex: `2025-12-17 23:35:58`)
- `serial`: (ex: `1766010947`)

### Response Body (JSON)

```json
{
  "code": "ok",
  "message": "foo",
  "response": <TODO>
}
```

```json
{
  "code": "ok",
  "message": "foo",
  "response": 123
}
```

## DNS TXT Delete Endpoint

* URL: `<server>?dns_txt_delete`
* HTTP Method: `POST`

- https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/remoting_client/API-docs/dns_txt_delete.html
- https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/remoting_client/examples/dns_txt_delete.php

### Request Body (JSON)

```json
{
  "session_id": "<session_id>",
  "primary_id": "<record_id>",
  "update_serial": true
}
```

### Response Body (JSON)

```json
{
  "code": "ok",
  "message": "foo",
  "response": <TODO>
}
```

```json
{
  "code": "ok",
  "message": "foo",
  "response": 1
}
```

---

https://www.ispconfig.org/
https://git.ispconfig.org/ispconfig/ispconfig3
https://forum.howtoforge.com/#ispconfig-3.23


https://github.com/tbrehm
https://github.com/ispconfig

