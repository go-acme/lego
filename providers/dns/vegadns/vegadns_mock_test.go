package vegadns

const tokenResponseMock = `
{
  "access_token":"699dd4ff-e381-46b8-8bf8-5de49dd56c1f",
  "token_type":"bearer",
  "expires_in":3600
}
`

const domainsResponseMock = `
{
  "domains":[
    {
      "domain_id":1,
      "domain":"example.com",
      "status":"active",
      "owner_id":0
    }
  ]
}
`

const recordsResponseMock = `
{
  "status":"ok",
  "total_records":2,
  "domain":{
    "status":"active",
    "domain":"example.com",
    "owner_id":0,
    "domain_id":1
  },
  "records":[
    {
      "retry":"2048",
      "minimum":"2560",
      "refresh":"16384",
      "email":"hostmaster.example.com",
      "record_type":"SOA",
      "expire":"1048576",
      "ttl":86400,
      "record_id":1,
      "nameserver":"ns1.example.com",
      "domain_id":1,
      "serial":""
    },
    {
      "name":"example.com",
      "value":"ns1.example.com",
      "record_type":"NS",
      "ttl":3600,
      "record_id":2,
      "location_id":null,
      "domain_id":1
    },
    {
      "name":"_acme-challenge.example.com",
      "value":"my_challenge",
      "record_type":"TXT",
      "ttl":3600,
      "record_id":3,
      "location_id":null,
      "domain_id":1
    }
  ]
}
`

const recordCreatedResponseMock = `
{
  "status":"ok",
  "record":{
    "name":"_acme-challenge.example.com",
    "value":"my_challenge",
    "record_type":"TXT",
    "ttl":3600,
    "record_id":3,
    "location_id":null,
    "domain_id":1
  }
}
`

const recordDeletedResponseMock = `{"status": "ok"}`
