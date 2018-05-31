package egoscale

// SecurityGroup represent a firewalling set of rules
type SecurityGroup struct {
	Account     string        `json:"account,omitempty" doc:"the account owning the security group"`
	Description string        `json:"description,omitempty" doc:"the description of the security group"`
	Domain      string        `json:"domain,omitempty" doc:"the domain name of the security group"`
	DomainID    string        `json:"domainid,omitempty" doc:"the domain ID of the security group"`
	EgressRule  []EgressRule  `json:"egressrule,omitempty" doc:"the list of egress rules associated with the security group"`
	ID          string        `json:"id,omitempty" doc:"the ID of the security group"`
	IngressRule []IngressRule `json:"ingressrule,omitempty" doc:"the list of ingress rules associated with the security group"`
	Name        string        `json:"name,omitempty" doc:"the name of the security group"`
	Tags        []ResourceTag `json:"tags,omitempty" doc:"the list of resource tags associated with the rule"`
}

// IngressRule represents the ingress rule
type IngressRule struct {
	Account               string              `json:"account,omitempty" doc:"account owning the security group rule"`
	Cidr                  string              `json:"cidr,omitempty" doc:"the CIDR notation for the base IP address of the security group rule"`
	Description           string              `json:"description,omitempty" doc:"description of the security group rule"`
	EndPort               uint16              `json:"endport,omitempty" doc:"the ending IP of the security group rule "`
	IcmpCode              uint8               `json:"icmpcode,omitempty" doc:"the code for the ICMP message response"`
	IcmpType              uint8               `json:"icmptype,omitempty" doc:"the type of the ICMP message response"`
	Protocol              string              `json:"protocol,omitempty" doc:"the protocol of the security group rule"`
	RuleID                string              `json:"ruleid,omitempty" doc:"the id of the security group rule"`
	SecurityGroupID       string              `json:"securitygroupid,omitempty"`
	SecurityGroupName     string              `json:"securitygroupname,omitempty" doc:"security group name"`
	StartPort             uint16              `json:"startport,omitempty" doc:"the starting IP of the security group rule"`
	Tags                  []ResourceTag       `json:"tags,omitempty" doc:"the list of resource tags associated with the rule"`
	UserSecurityGroupList []UserSecurityGroup `json:"usersecuritygrouplist,omitempty"`
}

// EgressRule represents the ingress rule
type EgressRule IngressRule

// UserSecurityGroup represents the traffic of another security group
type UserSecurityGroup struct {
	Group   string `json:"group,omitempty"`
	Account string `json:"account,omitempty"`
}

// SecurityGroupResponse represents a generic security group response
type SecurityGroupResponse struct {
	SecurityGroup SecurityGroup `json:"securitygroup"`
}

// CreateSecurityGroup represents a security group creation
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/createSecurityGroup.html
type CreateSecurityGroup struct {
	Name        string `json:"name" doc:"name of the security group"`
	Account     string `json:"account,omitempty" doc:"an optional account for the security group. Must be used with domainId."`
	Description string `json:"description,omitempty" doc:"the description of the security group"`
	DomainID    string `json:"domainid,omitempty" doc:"an optional domainId for the security group. If the account parameter is used, domainId must also be used."`
}

// CreateSecurityGroupResponse represents a new security group
type CreateSecurityGroupResponse SecurityGroupResponse

// DeleteSecurityGroup represents a security group deletion
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/deleteSecurityGroup.html
type DeleteSecurityGroup struct {
	Account  string `json:"account,omitempty" doc:"the account of the security group. Must be specified with domain ID"`
	DomainID string `json:"domainid,omitempty" doc:"the domain ID of account owning the security group"`
	ID       string `json:"id,omitempty" doc:"The ID of the security group. Mutually exclusive with name parameter"`
	Name     string `json:"name,omitempty" doc:"The ID of the security group. Mutually exclusive with id parameter"`
}

// AuthorizeSecurityGroupIngress (Async) represents the ingress rule creation
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/authorizeSecurityGroupIngress.html
type AuthorizeSecurityGroupIngress struct {
	Account               string              `json:"account,omitempty" doc:"an optional account for the security group. Must be used with domainId."`
	CidrList              []string            `json:"cidrlist,omitempty" doc:"the cidr list associated"`
	Description           string              `json:"description,omitempty" doc:"the description of the ingress/egress rule"`
	DomainID              string              `json:"domainid,omitempty" doc:"an optional domainId for the security group. If the account parameter is used, domainId must also be used."`
	EndPort               uint16              `json:"endport,omitempty" doc:"end port for this ingress/egress rule"`
	IcmpCode              uint8               `json:"icmpcode,omitempty" doc:"error code for this icmp message"`
	IcmpType              uint8               `json:"icmptype,omitempty" doc:"type of the icmp message being sent"`
	Protocol              string              `json:"protocol,omitempty" doc:"TCP is default. UDP, ICMP, ICMPv6, AH, ESP, GRE are the other supported protocols"`
	SecurityGroupID       string              `json:"securitygroupid,omitempty" doc:"The ID of the security group. Mutually exclusive with securityGroupName parameter"`
	SecurityGroupName     string              `json:"securitygroupname,omitempty" doc:"The name of the security group. Mutually exclusive with securityGroupId parameter"`
	StartPort             uint16              `json:"startport,omitempty" doc:"start port for this ingress/egress rule"`
	UserSecurityGroupList []UserSecurityGroup `json:"usersecuritygrouplist,omitempty" doc:"user to security group mapping"`
}

// AuthorizeSecurityGroupIngressResponse represents the new egress rule
// /!\ the Cloud Stack API document is not fully accurate. /!\
type AuthorizeSecurityGroupIngressResponse SecurityGroupResponse

// AuthorizeSecurityGroupEgress (Async) represents the egress rule creation
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/authorizeSecurityGroupEgress.html
type AuthorizeSecurityGroupEgress AuthorizeSecurityGroupIngress

// AuthorizeSecurityGroupEgressResponse represents the new egress rule
// /!\ the Cloud Stack API document is not fully accurate. /!\
type AuthorizeSecurityGroupEgressResponse CreateSecurityGroupResponse

// RevokeSecurityGroupIngress (Async) represents the ingress/egress rule deletion
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/revokeSecurityGroupIngress.html
type RevokeSecurityGroupIngress struct {
	ID string `json:"id" doc:"The ID of the ingress/egress rule"`
}

// RevokeSecurityGroupEgress (Async) represents the ingress/egress rule deletion
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/revokeSecurityGroupEgress.html
type RevokeSecurityGroupEgress RevokeSecurityGroupIngress

// ListSecurityGroups represents a search for security groups
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/listSecurityGroups.html
type ListSecurityGroups struct {
	Account           string        `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	DomainID          string        `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	ID                string        `json:"id,omitempty" doc:"list the security group by the id provided"`
	IsRecursive       *bool         `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword           string        `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll           *bool         `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Page              int           `json:"page,omitempty"`
	PageSize          int           `json:"pagesize,omitempty"`
	SecurityGroupName string        `json:"securitygroupname,omitempty" doc:"lists security groups by name"`
	Tags              []ResourceTag `json:"tags,omitempty" doc:"List resources by tags (key/value pairs)"`
	VirtualMachineID  string        `json:"virtualmachineid,omitempty" doc:"lists security groups by virtual machine id"`
}

// ListSecurityGroupsResponse represents a list of security groups
type ListSecurityGroupsResponse struct {
	Count         int             `json:"count"`
	SecurityGroup []SecurityGroup `json:"securitygroup"`
}
