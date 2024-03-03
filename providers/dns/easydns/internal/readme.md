The API doc is mainly wrong on the response schema:

ex:

- the doc for `/zones/records/all/{domain}`

```json
{
  "msg": "string",
  "status": 200,
  "tm": 1709190001,
  "data": {
    "id": 60898922,
    "domain": "example.com",
    "host": "hosta",
    "ttl": 300,
    "prio": 0,
    "geozone_id": 0,
    "type": "A",
    "rdata": "1.2.3.4",
    "last_mod": "2019-08-28 19:09:50"
  },
  "count": 0,
  "total": 0,
  "start": 0,
  "max": 0
}
```

- The reality:

```json
{
  "tm": 1709190001,
  "data": [
    {
      "id": "60898922",
      "domain": "example.com",
      "host": "hosta",
      "ttl": "300",
      "prio": "0",
      "geozone_id": "0",
      "type": "A",
      "rdata": "1.2.3.4",
      "last_mod": "2019-08-28 19:09:50"
    }
  ],
  "count": 0,
  "total": 0,
  "start": 0,
  "max": 0,
  "status": 200
}
```

`data` is an array.
`id`, `ttl`, `geozone_id` are strings.
