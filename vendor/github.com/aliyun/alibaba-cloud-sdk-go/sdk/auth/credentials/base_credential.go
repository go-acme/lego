package credentials

type BaseCredential struct {
	AccessKeyId     string
	AccessKeySecret string
}

func NewBaseCredential(accessKeyId, accessKeySecret string) *BaseCredential {
	return &BaseCredential{
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
	}
}
