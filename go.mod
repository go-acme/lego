module github.com/go-acme/lego/v4

go 1.22.0

require (
	cloud.google.com/go/compute/metadata v0.5.2
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.16.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.8.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns v1.2.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns v1.3.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph v0.9.0
	github.com/Azure/go-autorest/autorest v0.11.29
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.13
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/BurntSushi/toml v1.4.0
	github.com/OpenDNS/vegadns2client v0.0.0-20180418235048-a3fa4a771d87
	github.com/akamai/AkamaiOPEN-edgegrid-golang v1.2.2
	github.com/aliyun/alibaba-cloud-sdk-go v1.63.47
	github.com/aws/aws-sdk-go-v2 v1.32.3
	github.com/aws/aws-sdk-go-v2/config v1.28.1
	github.com/aws/aws-sdk-go-v2/credentials v1.17.42
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.42.3
	github.com/aws/aws-sdk-go-v2/service/route53 v1.46.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.66.2
	github.com/aws/aws-sdk-go-v2/service/sts v1.32.3
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/civo/civogo v0.3.11
	github.com/cloudflare/cloudflare-go v0.108.0
	github.com/cpu/goacmedns v0.1.1
	github.com/dnsimple/dnsimple-go v1.7.0
	github.com/exoscale/egoscale/v3 v3.1.7
	github.com/go-jose/go-jose/v4 v4.0.4
	github.com/go-viper/mapstructure/v2 v2.2.1
	github.com/google/go-querystring v1.1.0
	github.com/gophercloud/gophercloud v1.14.1
	github.com/gophercloud/utils v0.0.0-20231010081019-80377eca5d56
	github.com/hashicorp/go-retryablehttp v0.7.7
	github.com/hashicorp/go-version v1.7.0
	github.com/huaweicloud/huaweicloud-sdk-go-v3 v0.1.120
	github.com/iij/doapi v0.0.0-20190504054126-0bbf12d6d7df
	github.com/infobloxopen/infoblox-go-client v1.1.1
	github.com/labbsr0x/bindman-dns-webhook v1.0.2
	github.com/linode/linodego v1.42.0
	github.com/liquidweb/liquidweb-go v1.6.4
	github.com/mattn/go-isatty v0.0.20
	github.com/miekg/dns v1.1.62
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
	github.com/nrdcg/porkbun v0.4.0
	github.com/nzdjb/go-metaname v1.0.0
	github.com/oracle/oci-go-sdk/v65 v65.77.1
	github.com/ovh/go-ovh v1.6.0
	github.com/pquerna/otp v1.4.0
	github.com/rainycape/memcache v0.0.0-20150622160815-1031fa0ce2f2
	github.com/regfish/regfish-dnsapi-go v0.1.1
	github.com/sacloud/api-client-go v0.2.10
	github.com/sacloud/iaas-api-go v1.12.0
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.30
	github.com/selectel/domains-go v1.1.0
	github.com/selectel/go-selvpcclient/v3 v3.1.1
	github.com/softlayer/softlayer-go v1.1.7
	github.com/stretchr/testify v1.9.0
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.1034
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod v1.0.1034
	github.com/transip/gotransip/v6 v6.26.0
	github.com/ultradns/ultradns-go-sdk v1.8.0-20241010134910-243eeec
	github.com/urfave/cli/v2 v2.27.5
	github.com/vinyldns/go-vinyldns v0.9.16
	github.com/volcengine/volc-sdk-golang v1.0.183
	github.com/vultr/govultr/v3 v3.9.1
	github.com/yandex-cloud/go-genproto v0.0.0-20241101135610-76a0cfc1a773
	github.com/yandex-cloud/go-sdk v0.0.0-20241101143304-947cf519f6bd
	golang.org/x/crypto v0.28.0
	golang.org/x/net v0.30.0
	golang.org/x/oauth2 v0.23.0
	golang.org/x/time v0.7.0
	google.golang.org/api v0.204.0
	gopkg.in/ns1/ns1-go.v2 v2.12.2
	gopkg.in/yaml.v2 v2.4.0
	software.sslmate.com/src/go-pkcs12 v0.5.0
)

require (
	cloud.google.com/go/auth v0.10.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.5 // indirect
	github.com/AdamSLevy/jsonrpc2/v14 v14.1.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.10.0 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.22 // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.6 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.6 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.22 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.4.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.3 // indirect
	github.com/aws/smithy-go v1.22.0 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.5 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.16.0 // indirect
	github.com/go-resty/resty/v2 v2.13.1 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/gofrs/flock v0.12.1 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.13.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/k0kubun/go-ansi v0.0.0-20180517002512-3bf9e2903213 // indirect
	github.com/kolo/xmlrpc v0.0.0-20220921171641-a4b6fa1dd06b // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/labbsr0x/goh v1.0.1 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/liquidweb/liquidweb-cli v0.6.9 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opentracing/opentracing-go v1.2.1-0.20220228012449-10b1cf09e00b // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml/v2 v2.1.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sacloud/go-http v0.1.8 // indirect
	github.com/sacloud/packages-go v0.0.10 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/smartystreets/assertions v1.0.1 // indirect
	github.com/smartystreets/go-aws-auth v0.0.0-20180515143844-0c1422d1fdb9 // indirect
	github.com/softlayer/xmlrpc v0.0.0-20200409220501-5f089df7cb7e // indirect
	github.com/sony/gobreaker v0.5.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.18.2 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	go.mongodb.org/mongo-driver v1.12.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.54.0 // indirect
	go.opentelemetry.io/otel v1.29.0 // indirect
	go.opentelemetry.io/otel/metric v1.29.0 // indirect
	go.opentelemetry.io/otel/trace v1.29.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	go.uber.org/ratelimit v0.3.0 // indirect
	golang.org/x/exp v0.0.0-20240213143201-ec583247a57a // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/tools v0.25.0 // indirect
	google.golang.org/genproto v0.0.0-20241021214115-324edc3d5d38 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241015192408-796eee8c2d53 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241021214115-324edc3d5d38 // indirect
	google.golang.org/grpc v1.67.1 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
