package vinyldns

const GetZoneResponse = `{
  "zone": {
      "accessLevel": "Delete",
      "account": "system",
      "acl": {
          "rules": []
      },
      "adminGroupId": "40000000-0000-0000-0000-000000000000",
      "adminGroupName": "OpsTeam",
      "created": "2020-07-15T21:15:36Z",
      "email": "Ops@company.invalid",
      "id": "00000000-0000-0000-0000-000000000000",
      "latestSync": "2020-07-15T21:15:36Z",
      "name": "vinyldns.invalid.",
      "shared": false,
      "status": "Active",
      "updated": "2021-03-03T18:02:47Z"
  }
}`

const FindEmptyRRSetResponse = `{
   "maxItems": 100,
   "nameSort": "ASC",
   "recordNameFilter": "_acme-challenge.host",
   "recordSets": []
}`

const FindRRSetResponse = `{
   "maxItems": 100,
   "nameSort": "ASC",
   "recordNameFilter": "_acme-challenge.host",
   "recordSets": [
       {
           "accessLevel": "Delete",
           "account": "",
           "created": "2021-03-04T00:51:43Z",
           "fqdn": "_acme-challenge.host.vinyldns.invalid.",
           "id": "30000000-0000-0000-0000-000000000000",
           "name": "_acme-challenge.host",
           "records": [
               {
                   "text": "O2UTPYgIzRNt5N27EVcNKDxv6goSF7ru3zi3chZXKUw"
               }
           ],
           "status": "Active",
           "ttl": 30,
           "type": "TXT",
           "updated": "2021-03-04T00:51:43Z",
           "zoneId": "00000000-0000-0000-0000-000000000000"
       }
   ]
}`

const CreateRRSetResponse = `{
 "changeType": "Create",
 "created": "2021-03-04T16:21:54Z",
 "id": "20000000-0000-0000-0000-000000000000",
 "recordSet": {
     "account": "",
     "created": "2021-03-04T16:21:54Z",
     "id": "11000000-0000-0000-0000-000000000000",
     "name": "_acme-challenge.host",
     "records": [
         {
             "text": "O2UTPYgIzRNt5N27EVcNKDxv6goSF7ru3zi3chZXKUw"
         }
     ],
     "status": "Pending",
     "ttl": 30,
     "type": "TXT",
     "zoneId": "00000000-0000-0000-0000-000000000000"
 },
 "singleBatchChangeIds": [],
 "status": "Pending",
 "userId": "50000000-0000-0000-0000-000000000000",
 "zone": {
     "account": "system",
     "acl": {
         "rules": []
     },
     "adminGroupId": "40000000-0000-0000-0000-000000000000",
     "created": "2020-07-15T21:15:36Z",
     "email": "Ops@company.invalid",
     "id": "00000000-0000-0000-0000-000000000000",
     "isTest": false,
     "latestSync": "2020-07-15T21:15:36Z",
     "name": "vinyldns.invalid.",
     "shared": false,
     "status": "Active",
     "updated": "2021-03-03T18:02:47Z"
     }
 }`

const GetCreateRRSetStatusResponse = `{
   "changeType": "Create",
   "created": "2021-03-04T00:49:00Z",
   "id": "27ba5c17-a217-4e8d-b662-b1dc8bee588f",
   "recordSet": {
       "account": "",
       "created": "2021-03-04T00:49:00Z",
       "id": "10000000-0000-0000-0000-000000000000",
       "name": "_acme-challenge.host",
       "records": [
           {
               "text": "O2UTPYgIzRNt5N27EVcNKDxv6goSF7ru3zi3chZXKUw"
           }
       ],
       "status": "Active",
       "ttl": 30,
       "type": "TXT",
       "updated": "2021-03-04T00:49:00Z",
       "zoneId": "00000000-0000-0000-0000-000000000000"
   },
   "singleBatchChangeIds": [],
   "status": "Complete",
   "userId": "50000000-0000-0000-0000-000000000000",
   "zone": {
       "account": "system",
       "acl": {
           "rules": []
       },
       "adminGroupId": "40000000-0000-0000-0000-000000000000",
       "created": "2020-07-15T21:15:36Z",
       "email": "Ops@company.invalid",
       "id": "00000000-0000-0000-0000-000000000000",
       "isTest": false,
       "latestSync": "2020-07-15T21:15:36Z",
       "name": "vinyldns.invalid.",
       "shared": false,
       "status": "Active",
       "updated": "2021-03-03T18:02:47Z"
   }
}`

const DeleteRRSetResponse = `{
 "changeType": "Delete",
 "created": "2021-03-04T16:21:54Z",
 "id": "20000000-0000-0000-0000-000000000000",
 "recordSet": {
     "account": "",
     "created": "2021-03-04T16:21:54Z",
     "id": "11000000-0000-0000-0000-000000000000",
     "name": "_acme-challenge.host",
     "records": [
         {
             "text": "O2UTPYgIzRNt5N27EVcNKDxv6goSF7ru3zi3chZXKUw"
         }
     ],
     "status": "Pending",
     "ttl": 30,
     "type": "TXT",
     "zoneId": "00000000-0000-0000-0000-000000000000"
 },
 "singleBatchChangeIds": [],
 "status": "Pending",
 "userId": "50000000-0000-0000-0000-000000000000",
 "zone": {
     "account": "system",
     "acl": {
         "rules": []
     },
     "adminGroupId": "40000000-0000-0000-0000-000000000000",
     "created": "2020-07-15T21:15:36Z",
     "email": "Ops@company.invalid",
     "id": "00000000-0000-0000-0000-000000000000",
     "isTest": false,
     "latestSync": "2020-07-15T21:15:36Z",
     "name": "vinyldns.invalid.",
     "shared": false,
     "status": "Active",
     "updated": "2021-03-03T18:02:47Z"
     }
 }`

const GetDeleteRRSetStatusResponse = `{
   "changeType": "Delete",
   "created": "2021-03-04T00:49:00Z",
   "id": "27ba5c17-a217-4e8d-b662-b1dc8bee588f",
   "recordSet": {
       "account": "",
       "created": "2021-03-04T00:49:00Z",
       "id": "10000000-0000-0000-0000-000000000000",
       "name": "_acme-challenge.host",
       "records": [
           {
               "text": "O2UTPYgIzRNt5N27EVcNKDxv6goSF7ru3zi3chZXKUw"
           }
       ],
       "status": "Active",
       "ttl": 30,
       "type": "TXT",
       "updated": "2021-03-04T00:49:00Z",
       "zoneId": "00000000-0000-0000-0000-000000000000"
   },
   "singleBatchChangeIds": [],
   "status": "Complete",
   "userId": "50000000-0000-0000-0000-000000000000",
   "zone": {
       "account": "system",
       "acl": {
           "rules": []
       },
       "adminGroupId": "40000000-0000-0000-0000-000000000000",
       "created": "2020-07-15T21:15:36Z",
       "email": "Ops@company.invalid",
       "id": "00000000-0000-0000-0000-000000000000",
       "isTest": false,
       "latestSync": "2020-07-15T21:15:36Z",
       "name": "vinyldns.invalid.",
       "shared": false,
       "status": "Active",
       "updated": "2021-03-03T18:02:47Z"
   }
}`
