## Error Response

```json
{
  "code":  "<TODO>",
  "message":  "<TODO>",
  "response": false
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
  "response": "abc"
}
```

- `response`: is the `sessionID`

## Get Client ID Endpoint

* URL: `<server>?client_get_id`
* HTTP Method: `POST`

- function `client_get_id`: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/interface/lib/classes/remote.d/client.inc.php#L97
- TABLE `sys_user`: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/install/sql/ispconfig3.sql?ref_type=heads#L1852
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
  "response": 123
}
```

## DNS Zone Get ID Endpoint

* URL: `<server>?dns_zone_get_id`
* HTTP Method: `POST`

- function `dns_zone_get_id`: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/interface/lib/classes/remote.d/dns.inc.php#L142
- TABLE `dns_soa`: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/install/sql/ispconfig3.sql?ref_type=heads#L615

### Request Body (JSON)

```json
{
  "session_id": "<session_id>",
  "origin": "<zone_name>"
}
```

### Response Body (JSON)

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

- function `dns_zone_get`: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/interface/lib/classes/remote.d/dns.inc.php#L87
- function `getDataRecord`: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/interface/lib/classes/remoting_lib.inc.php#L248
- TABLE `dns_soa`: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/install/sql/ispconfig3.sql?ref_type=heads#L615
- Depending on the request, the response may be an array or an object (`primary_id` can be a string, an array or an object).
- https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/remoting_client/API-docs/dns_zone_get.html
- https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/remoting_client/examples/dns_zone_get.php

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

- function `dns_txt_get`: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/interface/lib/classes/remote.d/dns.inc.php#L640
- function `dns_rr_get`: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/interface/lib/classes/remote.d/dns.inc.php#L195
- form: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/interface/web/dns/form/dns_txt.tform.php
- TABLE `dns_rr`: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/install/sql/ispconfig3.sql?ref_type=heads#L490
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
  "response": {
    "id": 123
  }
}
```

## DNS TXT Add Endpoint

* URL: `<server>?dns_txt_add`
* HTTP Method: `POST`

- function `dns_txt_add`: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/interface/lib/classes/remote.d/dns.inc.php#L645
- function `dns_rr_add` https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/interface/lib/classes/remote.d/dns.inc.php#L212
- form: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/interface/web/dns/form/dns_txt.tform.php
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
    "update_serial": true
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
  "response": "123"
}
```

## DNS TXT Delete Endpoint

* URL: `<server>?dns_txt_delete`
* HTTP Method: `POST`

- function `dns_txt_delete`: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/interface/lib/classes/remote.d/dns.inc.php#L655
- function `dns_rr_delete`: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/interface/lib/classes/remote.d/dns.inc.php#L247
- form: https://git.ispconfig.org/ispconfig/ispconfig3/-/blob/develop/interface/web/dns/form/dns_txt.tform.php
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
  "response": 1
}
```

---

https://www.ispconfig.org/
https://git.ispconfig.org/ispconfig/ispconfig3
https://forum.howtoforge.com/#ispconfig-3.23
