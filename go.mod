module github.com/go-acme/lego/v4

go 1.22

// github.com/exoscale/egoscale v1.19.0 => It is an error, please don't use it.

require (
	cloud.google.com/go/compute/metadata v0.3.0
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.12.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.6.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns v1.2.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns v1.2.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph v0.9.0
	github.com/Azure/go-autorest/autorest v0.11.29
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.13
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/BurntSushi/toml v1.4.0
	github.com/OpenDNS/vegadns2client v0.0.0-20180418235048-a3fa4a771d87
	github.com/akamai/AkamaiOPEN-edgegrid-golang v1.2.2
	github.com/aliyun/alibaba-cloud-sdk-go v1.62.712
	github.com/aws/aws-sdk-go-v2 v1.27.2
	github.com/aws/aws-sdk-go-v2/config v1.27.18
	github.com/aws/aws-sdk-go-v2/credentials v1.17.18
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.38.3
	github.com/aws/aws-sdk-go-v2/service/route53 v1.40.10
	github.com/aws/aws-sdk-go-v2/service/s3 v1.55.1
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.12
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/civo/civogo v0.3.11
	github.com/cloudflare/cloudflare-go v0.97.0
	github.com/cpu/goacmedns v0.1.1
	github.com/dnsimple/dnsimple-go v1.7.0
	github.com/exoscale/egoscale v0.102.3
	github.com/go-jose/go-jose/v4 v4.0.2
	github.com/go-viper/mapstructure/v2 v2.0.0
	github.com/google/go-querystring v1.1.0
	github.com/gophercloud/gophercloud v1.12.0
	github.com/gophercloud/utils v0.0.0-20231010081019-80377eca5d56
	github.com/hashicorp/go-retryablehttp v0.7.7
	github.com/iij/doapi v0.0.0-20190504054126-0bbf12d6d7df
	github.com/infobloxopen/infoblox-go-client v1.1.1
	github.com/labbsr0x/bindman-dns-webhook v1.0.2
	github.com/linode/linodego v1.28.0
	github.com/liquidweb/liquidweb-go v1.6.4
	github.com/mattn/go-isatty v0.0.20
	github.com/miekg/dns v1.1.59
	github.com/mimuret/golang-iij-dpf v0.9.1
	github.com/namedotcom/go v0.0.0-20180403034216-08470befbe04
	github.com/nrdcg/auroradns v1.1.0
	github.com/nrdcg/bunny-go v0.0.0-20240207213615-dde5bf4577a3
	github.com/nrdcg/desec v0.8.0
	github.com/nrdcg/dnspod-go v0.4.0
	github.com/nrdcg/freemyip v0.2.0
	github.com/nrdcg/goinwx v0.10.0
	github.com/nrdcg/mailinabox v0.2.0
	github.com/nrdcg/namesilo v0.2.1
	github.com/nrdcg/nodion v0.1.0
	github.com/nrdcg/porkbun v0.3.0
	github.com/nzdjb/go-metaname v1.0.0
	github.com/oracle/oci-go-sdk/v65 v65.63.1
	github.com/ovh/go-ovh v1.5.1
	github.com/pquerna/otp v1.4.0
	github.com/rainycape/memcache v0.0.0-20150622160815-1031fa0ce2f2
	github.com/sacloud/api-client-go v0.2.10
	github.com/sacloud/iaas-api-go v1.12.0
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.27
	github.com/selectel/domains-go v1.1.0
	github.com/selectel/go-selvpcclient/v3 v3.1.1
	github.com/softlayer/softlayer-go v1.1.5
	github.com/stretchr/testify v1.9.0
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.898
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod v1.0.898
	github.com/transip/gotransip/v6 v6.23.0
	github.com/ultradns/ultradns-go-sdk v1.6.1-20231103022937-8589b6a
	github.com/urfave/cli/v2 v2.27.2
	github.com/vinyldns/go-vinyldns v0.9.16
	github.com/vultr/govultr/v2 v2.17.2
	github.com/yandex-cloud/go-genproto v0.0.0-20240318083951-4fe6125f286e
	github.com/yandex-cloud/go-sdk v0.0.0-20240318084659-dfa50323a0b4
	golang.org/x/crypto v0.24.0
	golang.org/x/net v0.26.0
	golang.org/x/oauth2 v0.21.0
	golang.org/x/time v0.5.0
	google.golang.org/api v0.172.0
	gopkg.in/ns1/ns1-go.v2 v2.7.13
	gopkg.in/yaml.v2 v2.4.0
	software.sslmate.com/src/go-pkcs12 v0.4.0
)

require (
	github.com/AdamSLevy/jsonrpc2/v14 v14.1.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.9.0 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.22 // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.6 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.2 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.3.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.17.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.24.5 // indirect
	github.com/aws/smithy-go v1.20.2 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deepmap/oapi-codegen v1.9.1 // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-resty/resty/v2 v2.11.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/gofrs/flock v0.10.0 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.3 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/k0kubun/go-ansi v0.0.0-20180517002512-3bf9e2903213 // indirect
	github.com/kolo/xmlrpc v0.0.0-20220921171641-a4b6fa1dd06b // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/labbsr0x/goh v1.0.1 // indirect
	github.com/liquidweb/liquidweb-cli v0.6.9 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opentracing/opentracing-go v1.2.1-0.20220228012449-10b1cf09e00b // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sacloud/go-http v0.1.8 // indirect
	github.com/sacloud/packages-go v0.0.10 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/smartystreets/assertions v1.0.1 // indirect
	github.com/smartystreets/go-aws-auth v0.0.0-20180515143844-0c1422d1fdb9 // indirect
	github.com/softlayer/xmlrpc v0.0.0-20200409220501-5f089df7cb7e // indirect
	github.com/sony/gobreaker v0.5.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/xrash/smetrics v0.0.0-20240312152122-5f08fbb34913 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
	go.uber.org/ratelimit v0.3.0 // indirect
	golang.org/x/mod v0.18.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/tools v0.22.0 // indirect
	google.golang.org/genproto v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240311132316-a219d84964c2 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/grpc v1.63.1 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
