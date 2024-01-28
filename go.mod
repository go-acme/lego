module github.com/go-acme/lego/v4

go 1.20

// github.com/exoscale/egoscale v1.19.0 => It is an error, please don't use it.

require (
	cloud.google.com/go/compute/metadata v0.2.3
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.6.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.3.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns v1.1.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns v1.1.0
	github.com/Azure/go-autorest/autorest v0.11.29
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.12
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/BurntSushi/toml v1.3.2
	github.com/OpenDNS/vegadns2client v0.0.0-20180418235048-a3fa4a771d87
	github.com/akamai/AkamaiOPEN-edgegrid-golang v1.2.2
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.1755
	github.com/aws/aws-sdk-go-v2 v1.24.1
	github.com/aws/aws-sdk-go-v2/config v1.26.6
	github.com/aws/aws-sdk-go-v2/credentials v1.16.16
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.34.0
	github.com/aws/aws-sdk-go-v2/service/route53 v1.37.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.48.1
	github.com/aws/aws-sdk-go-v2/service/sts v1.26.7
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/civo/civogo v0.3.11
	github.com/cloudflare/cloudflare-go v0.86.0
	github.com/cpu/goacmedns v0.1.1
	github.com/dnsimple/dnsimple-go v1.2.0
	github.com/exoscale/egoscale v0.102.3
	github.com/go-jose/go-jose/v3 v3.0.1
	github.com/google/go-querystring v1.1.0
	github.com/gophercloud/gophercloud v1.0.0
	github.com/gophercloud/utils v0.0.0-20210216074907-f6de111f2eae
	github.com/hashicorp/go-retryablehttp v0.7.5
	github.com/iij/doapi v0.0.0-20190504054126-0bbf12d6d7df
	github.com/infobloxopen/infoblox-go-client v1.1.1
	github.com/labbsr0x/bindman-dns-webhook v1.0.2
	github.com/linode/linodego v1.28.0
	github.com/liquidweb/liquidweb-go v1.6.4
	github.com/mattn/go-isatty v0.0.20
	github.com/miekg/dns v1.1.58
	github.com/mimuret/golang-iij-dpf v0.9.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/namedotcom/go v0.0.0-20180403034216-08470befbe04
	github.com/nrdcg/auroradns v1.1.0
	github.com/nrdcg/bunny-go v0.0.0-20230728143221-c9dda82568d9
	github.com/nrdcg/desec v0.7.0
	github.com/nrdcg/dnspod-go v0.4.0
	github.com/nrdcg/freemyip v0.2.0
	github.com/nrdcg/goinwx v0.10.0
	github.com/nrdcg/namesilo v0.2.1
	github.com/nrdcg/nodion v0.1.0
	github.com/nrdcg/porkbun v0.3.0
	github.com/nzdjb/go-metaname v1.0.0
	github.com/oracle/oci-go-sdk v24.3.0+incompatible
	github.com/ovh/go-ovh v1.4.3
	github.com/pquerna/otp v1.4.0
	github.com/rainycape/memcache v0.0.0-20150622160815-1031fa0ce2f2
	github.com/sacloud/api-client-go v0.2.8
	github.com/sacloud/iaas-api-go v1.11.1
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.22
	github.com/softlayer/softlayer-go v1.1.3
	github.com/stretchr/testify v1.8.4
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.490
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod v1.0.490
	github.com/transip/gotransip/v6 v6.23.0
	github.com/ultradns/ultradns-go-sdk v1.6.1-20231103022937-8589b6a
	github.com/urfave/cli/v2 v2.27.1
	github.com/vinyldns/go-vinyldns v0.9.16
	github.com/vultr/govultr/v2 v2.17.2
	github.com/yandex-cloud/go-genproto v0.0.0-20220805142335-27b56ddae16f
	github.com/yandex-cloud/go-sdk v0.0.0-20220805164847-cf028e604997
	golang.org/x/crypto v0.18.0
	golang.org/x/net v0.20.0
	golang.org/x/oauth2 v0.16.0
	golang.org/x/time v0.5.0
	google.golang.org/api v0.126.0
	gopkg.in/ns1/ns1-go.v2 v2.7.13
	gopkg.in/yaml.v2 v2.4.0
	software.sslmate.com/src/go-pkcs12 v0.4.0
)

require (
	cloud.google.com/go/compute v1.20.1 // indirect
	github.com/AdamSLevy/jsonrpc2/v14 v14.1.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.3.0 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.22 // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.5 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.0.0 // indirect
	github.com/andres-erbsen/clock v0.0.0-20160526145045-9e14626cd129 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.5.4 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.14.11 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.5.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.7.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.10.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.10.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.16.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.18.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.21.7 // indirect
	github.com/aws/smithy-go v1.19.0 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deepmap/oapi-codegen v1.9.1 // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/go-resty/resty/v2 v2.11.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/s2a-go v0.1.4 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/googleapis/gax-go/v2 v2.11.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/k0kubun/go-ansi v0.0.0-20180517002512-3bf9e2903213 // indirect
	github.com/kolo/xmlrpc v0.0.0-20220921171641-a4b6fa1dd06b // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/labbsr0x/goh v1.0.1 // indirect
	github.com/liquidweb/liquidweb-cli v0.6.9 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sacloud/go-http v0.1.6 // indirect
	github.com/sacloud/packages-go v0.0.9 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/smartystreets/assertions v1.0.1 // indirect
	github.com/smartystreets/go-aws-auth v0.0.0-20180515143844-0c1422d1fdb9 // indirect
	github.com/softlayer/xmlrpc v0.0.0-20200409220501-5f089df7cb7e // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/stretchr/objx v0.5.1 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/ratelimit v0.2.0 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.17.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/grpc v1.55.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
