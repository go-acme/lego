package versio

const tokenResponseMock = `
{
  "access_token":"699dd4ff-e381-46b8-8bf8-5de49dd56c1f",
  "token_type":"bearer",
  "expires_in":3600
}
`

const tokenFailToFindZoneMock = `{"error":{"code":401,"message":"ObjectDoesNotExist|Domain not found"}}`

const tokenFailToCreateTXTMock = `{"error":{"code":400,"message":"ProcessError|DNS record invalid type _acme-challenge.example.eu. TST"}}`
