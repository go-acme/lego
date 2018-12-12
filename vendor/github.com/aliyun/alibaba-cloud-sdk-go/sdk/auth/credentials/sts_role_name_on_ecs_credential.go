package credentials

type StsRoleNameOnEcsCredential struct {
	RoleName string
}

func NewStsRoleNameOnEcsCredential(roleName string) *StsRoleNameOnEcsCredential {
	return &StsRoleNameOnEcsCredential{
		RoleName: roleName,
	}
}
