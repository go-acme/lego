package rackspace

const recordDeleteMock = `
{
  "status": "RUNNING",
  "verb": "DELETE",
  "jobId": "00000000-0000-0000-0000-0000000000",
  "callbackUrl": "https://dns.api.rackspacecloud.com/v1.0/123456/status/00000000-0000-0000-0000-0000000000",
  "requestUrl": "https://dns.api.rackspacecloud.com/v1.0/123456/domains/112233/recordsid=TXT-654321"
}
`

const recordDetailsMock = `
{
  "records": [
    {
      "name": "_acme-challenge.example.com",
      "id": "TXT-654321",
      "type": "TXT",
      "data": "pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM",
      "ttl": 300,
      "updated": "1970-01-01T00:00:00.000+0000",
      "created": "1970-01-01T00:00:00.000+0000"
    }
  ]
}
`

const zoneDetailsMock = `
{
  "domains": [
    {
      "name": "example.com",
      "id": "112233",
      "emailAddress": "hostmaster@example.com",
      "updated": "1970-01-01T00:00:00.000+0000",
      "created": "1970-01-01T00:00:00.000+0000"
    }
  ],
  "totalEntries": 1
}
`

const identityResponseMock = `
{
  "access": {
    "token": {
      "id": "testToken",
      "expires": "1970-01-01T00:00:00.000Z",
      "tenant": {
        "id": "123456",
        "name": "123456"
      },
      "RAX-AUTH:authenticatedBy": [
        "APIKEY"
      ]
    },
    "serviceCatalog": [
      {
        "type": "rax:dns",
        "endpoints": [
          {
            "publicURL": "https://dns.api.rackspacecloud.com/v1.0/123456",
            "tenantId": "123456"
          }
        ],
        "name": "cloudDNS"
      }
    ],
    "user": {
      "id": "fakeUseID",
      "name": "testUser"
    }
  }
}
`

const recordResponseMock = `
{
  "request": "{\"records\":[{\"name\":\"_acme-challenge.example.com\",\"type\":\"TXT\",\"data\":\"pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM\",\"ttl\":300}]}",
  "status": "RUNNING",
  "verb": "POST",
  "jobId": "00000000-0000-0000-0000-0000000000",
  "callbackUrl": "https://dns.api.rackspacecloud.com/v1.0/123456/status/00000000-0000-0000-0000-0000000000",
  "requestUrl": "https://dns.api.rackspacecloud.com/v1.0/123456/domains/112233/records"
}
`
