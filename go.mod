module github.com/go-acme/lego/v4

go 1.17

// github.com/exoscale/egoscale v1.19.0 => It is an error, please don't use it.
// github.com/linode/linodego v1.0.0 => It is an error, please don't use it.
require (
	cloud.google.com/go v0.54.0
	github.com/Azure/azure-sdk-for-go v32.4.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.19
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.8
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/BurntSushi/toml v0.3.1
	github.com/OpenDNS/vegadns2client v0.0.0-20180418235048-a3fa4a771d87
	github.com/akamai/AkamaiOPEN-edgegrid-golang v1.1.1
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.1183
	github.com/aws/aws-sdk-go v1.39.0
	github.com/cenkalti/backoff/v4 v4.1.1
	github.com/cloudflare/cloudflare-go v0.20.0
	github.com/cpu/goacmedns v0.1.1
	github.com/dnsimple/dnsimple-go v0.70.1
	github.com/exoscale/egoscale v0.67.0
	github.com/google/go-querystring v1.1.0
	github.com/gophercloud/gophercloud v0.16.0
	github.com/gophercloud/utils v0.0.0-20210216074907-f6de111f2eae
	github.com/iij/doapi v0.0.0-20190504054126-0bbf12d6d7df
	github.com/infobloxopen/infoblox-go-client v1.1.1
	github.com/labbsr0x/bindman-dns-webhook v1.0.2
	github.com/linode/linodego v0.31.1
	github.com/liquidweb/liquidweb-go v1.6.3
	github.com/miekg/dns v1.1.47
	github.com/mimuret/golang-iij-dpf v0.7.1
	github.com/mitchellh/mapstructure v1.4.1
	github.com/namedotcom/go v0.0.0-20180403034216-08470befbe04
	github.com/nrdcg/auroradns v1.0.1
	github.com/nrdcg/desec v0.6.0
	github.com/nrdcg/dnspod-go v0.4.0
	github.com/nrdcg/freemyip v0.2.0
	github.com/nrdcg/goinwx v0.8.1
	github.com/nrdcg/namesilo v0.2.1
	github.com/nrdcg/porkbun v0.1.1
	github.com/oracle/oci-go-sdk v24.3.0+incompatible
	github.com/ovh/go-ovh v1.1.0
	github.com/pquerna/otp v1.3.0
	github.com/rainycape/memcache v0.0.0-20150622160815-1031fa0ce2f2
	github.com/sacloud/libsacloud v1.36.2
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.7.0.20210127161313-bd30bebeac4f
	github.com/softlayer/softlayer-go v1.0.3
	github.com/stretchr/testify v1.7.1
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.287
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod v1.0.287
	github.com/transip/gotransip/v6 v6.6.1
	github.com/urfave/cli/v2 v2.3.0
	github.com/vinyldns/go-vinyldns v0.9.16
	github.com/vultr/govultr/v2 v2.16.0
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/time v0.0.0-20210611083556-38a9dc6acbc6
	google.golang.org/api v0.20.0
	gopkg.in/ns1/ns1-go.v2 v2.6.2
	gopkg.in/square/go-jose.v2 v2.6.0
	gopkg.in/yaml.v2 v2.4.0
	software.sslmate.com/src/go-pkcs12 v0.0.0-20210415151418-c5206de65a78
)

require (
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.13 // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.2 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deepmap/oapi-codegen v1.6.1 // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible // indirect
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/go-resty/resty/v2 v2.1.1-0.20191201195748-d7b97669fe48 // indirect
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.1 // indirect
	github.com/jarcoal/httpmock v1.0.8 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/k0kubun/go-ansi v0.0.0-20180517002512-3bf9e2903213 // indirect
	github.com/kolo/xmlrpc v0.0.0-20200310150728-e0350524596b // indirect
	github.com/labbsr0x/goh v1.0.1 // indirect
	github.com/liquidweb/go-lwApi v0.0.5 // indirect
	github.com/liquidweb/liquidweb-cli v0.6.9 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/smartystreets/assertions v1.0.1 // indirect
	github.com/smartystreets/go-aws-auth v0.0.0-20180515143844-0c1422d1fdb9 // indirect
	github.com/softlayer/xmlrpc v0.0.0-20200409220501-5f089df7cb7e // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	go.opencensus.io v0.22.3 // indirect
	go.uber.org/ratelimit v0.0.0-20180316092928-c15da0234277 // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/sys v0.0.0-20220209214540-3681064d5158 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.6-0.20210726203631-07bc1bf47fb2 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20200305110556-506484158171 // indirect
	google.golang.org/grpc v1.27.1 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
