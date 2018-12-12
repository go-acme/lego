package credentials

type StsRoleArnCredential struct {
	AccessKeyId           string
	AccessKeySecret       string
	RoleArn               string
	RoleSessionName       string
	RoleSessionExpiration int
}

func NewStsRoleArnCredential(accessKeyId, accessKeySecret, roleArn, roleSessionName string, roleSessionExpiration int) *StsRoleArnCredential {
	return &StsRoleArnCredential{
		AccessKeyId:           accessKeyId,
		AccessKeySecret:       accessKeySecret,
		RoleArn:               roleArn,
		RoleSessionName:       roleSessionName,
		RoleSessionExpiration: roleSessionExpiration,
	}
}
