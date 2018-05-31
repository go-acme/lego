package egoscale

// https://github.com/apache/cloudstack/blob/master/api/src/main/java/com/cloud/configuration/Resource.java

// ResourceTypeName represents the name of a resource type (for limits)
type ResourceTypeName string

const (
	// VirtualMachineTypeName is the resource type name of a VM
	VirtualMachineTypeName ResourceTypeName = "user_vm"
	// IPAddressTypeName is the resource type name of an IP address
	IPAddressTypeName ResourceTypeName = "public_ip"
	// VolumeTypeName is the resource type name of a volume
	VolumeTypeName ResourceTypeName = "volume"
	// SnapshotTypeName is the resource type name of a snapshot
	SnapshotTypeName ResourceTypeName = "snapshot"
	// TemplateTypeName is the resource type name of a template
	TemplateTypeName ResourceTypeName = "template"
	// ProjectTypeName is the resource type name of a project
	ProjectTypeName ResourceTypeName = "project"
	// NetworkTypeName is the resource type name of a network
	NetworkTypeName ResourceTypeName = "network"
	// VPCTypeName is the resource type name of a VPC
	VPCTypeName ResourceTypeName = "vpc"
	// CPUTypeName is the resource type name of a CPU
	CPUTypeName ResourceTypeName = "cpu"
	// MemoryTypeName is the resource type name of Memory
	MemoryTypeName ResourceTypeName = "memory"
	// PrimaryStorageTypeName is the resource type name of primary storage
	PrimaryStorageTypeName ResourceTypeName = "primary_storage"
	// SecondaryStorageTypeName is the resource type name of secondary storage
	SecondaryStorageTypeName ResourceTypeName = "secondary_storage"
)

// ResourceType represents the ID of a resource type (for limits)
type ResourceType int64

//go:generate stringer -type=ResourceType
const (
	// VirtualMachineType is the resource type ID of a VM
	VirtualMachineType ResourceType = iota
	// IPAddressType is the resource type ID of an IP address
	IPAddressType
	// VolumeType is the resource type ID of a volume
	VolumeType
	// SnapshotType is the resource type ID of a snapshot
	SnapshotType
	// TemplateType is the resource type ID of a template
	TemplateType
	// ProjectType is the resource type ID of a project
	ProjectType
	// NetworkType is the resource type ID of a network
	NetworkType
	// VPCType is the resource type ID of a VPC
	VPCType
	// CPUType is the resource type ID of a CPU
	CPUType
	// MemoryType is the resource type ID of Memory
	MemoryType
	// PrimaryStorageType is the resource type ID of primary storage
	PrimaryStorageType
	// SecondaryStorageType is the resource type ID of secondary storage
	SecondaryStorageType
)

// ResourceLimit represents the limit on a particular resource
type ResourceLimit struct {
	Account          string           `json:"account,omitempty" doc:"the account of the resource limit"`
	Domain           string           `json:"domain,omitempty" doc:"the domain name of the resource limit"`
	DomainID         string           `json:"domainid,omitempty" doc:"the domain ID of the resource limit"`
	Max              int64            `json:"max,omitempty" doc:"the maximum number of the resource. A -1 means the resource currently has no limit."`
	ResourceType     ResourceType     `json:"resourcetype,omitempty" doc:"resource type. Values include 0, 1, 2, 3, 4, 6, 7, 8, 9, 10, 11. See the resourceType parameter for more information on these values."`
	ResourceTypeName ResourceTypeName `json:"resourcetypename,omitempty" doc:"resource type name. Values include user_vm, public_ip, volume, snapshot, template, project, network, vpc, cpu, memory, primary_storage, secondary_storage."`
}

// APILimit represents the limit count
type APILimit struct {
	Account     string `json:"account,omitempty" doc:"the account name of the api remaining count"`
	Accountid   string `json:"accountid,omitempty" doc:"the account uuid of the api remaining count"`
	APIAllowed  int    `json:"apiAllowed,omitempty" doc:"currently allowed number of apis"`
	APIIssued   int    `json:"apiIssued,omitempty" doc:"number of api already issued"`
	ExpireAfter int64  `json:"expireAfter,omitempty" doc:"seconds left to reset counters"`
}

// ListResourceLimits lists the resource limits
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.4/user/listResourceLimits.html
type ListResourceLimits struct {
	Account          string           `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	DomainID         string           `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	ID               int64            `json:"id,omitempty" doc:"Lists resource limits by ID."`
	IsRecursive      *bool            `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword          string           `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll          *bool            `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Page             int              `json:"page,omitempty"`
	PageSize         int              `json:"pagesize,omitempty"`
	ResourceType     ResourceType     `json:"resourcetype,omitempty" doc:"Type of resource. Values are 0, 1, 2, 3, 4, 6, 7, 8, 9, 10 and 11. 0 - Instance. Number of instances a user can create. 1 - IP. Number of public IP addresses an account can own. 2 - Volume. Number of disk volumes an account can own. 3 - Snapshot. Number of snapshots an account can own. 4 - Template. Number of templates an account can register/create. 5 - Project. Number of projects an account can own. 6 - Network. Number of networks an account can own. 7 - VPC. Number of VPC an account can own. 8 - CPU. Number of CPU an account can allocate for his resources. 9 - Memory. Amount of RAM an account can allocate for his resources. 10 - PrimaryStorage. Total primary storage space (in GiB) a user can use. 11 - SecondaryStorage. Total secondary storage space (in GiB) a user can use. 12 - Elastic IP. Number of public elastic IP addresses an account can own. 13 - SMTP. If the account is allowed SMTP outbound traffic."`
	ResourceTypeName ResourceTypeName `json:"resourcetypename,omitempty" doc:"Type of resource (wins over resourceType if both are provided). Values are: user_vm - Instance. Number of instances a user can create. public_ip - IP. Number of public IP addresses an account can own. volume - Volume. Number of disk volumes an account can own. snapshot - Snapshot. Number of snapshots an account can own. template - Template. Number of templates an account can register/create. project - Project. Number of projects an account can own. network - Network. Number of networks an account can own. vpc - VPC. Number of VPC an account can own. cpu - CPU. Number of CPU an account can allocate for his resources. memory - Memory. Amount of RAM an account can allocate for his resources. primary_storage - PrimaryStorage. Total primary storage space (in GiB) a user can use. secondary_storage - SecondaryStorage. Total secondary storage space (in GiB) a user can use. public_elastic_ip - IP. Number of public elastic IP addresses an account can own. smtp - SG. If the account is allowed SMTP outbound traffic."`
}

// ListResourceLimitsResponse represents a list of resource limits
type ListResourceLimitsResponse struct {
	Count         int             `json:"count"`
	ResourceLimit []ResourceLimit `json:"resourcelimit"`
}

// UpdateResourceLimit updates the resource limit
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.4/root_admin/updateResourceLimit.html
type UpdateResourceLimit struct {
	Account      string       `json:"account,omitempty" doc:"Update resource for a specified account. Must be used with the domainId parameter."`
	DomainID     string       `json:"domainid,omitempty" doc:"Update resource limits for all accounts in specified domain. If used with the account parameter, updates resource limits for a specified account in specified domain."`
	Max          int64        `json:"max,omitempty" doc:"Maximum resource limit."`
	ResourceType ResourceType `json:"resourcetype" doc:"Type of resource to update. Values are 0, 1, 2, 3, 4, 6, 7, 8, 9, 10 and 11. 0 - Instance. Number of instances a user can create. 1 - IP. Number of public IP addresses a user can own. 2 - Volume. Number of disk volumes a user can create. 3 - Snapshot. Number of snapshots a user can create. 4 - Template. Number of templates that a user can register/create. 6 - Network. Number of guest network a user can create. 7 - VPC. Number of VPC a user can create. 8 - CPU. Total number of CPU cores a user can use. 9 - Memory. Total Memory (in MB) a user can use. 10 - PrimaryStorage. Total primary storage space (in GiB) a user can use. 11 - SecondaryStorage. Total secondary storage space (in GiB) a user can use."`
}

// UpdateResourceLimitResponse represents an updated resource limit
type UpdateResourceLimitResponse struct {
	ResourceLimit ResourceLimit `json:"resourcelimit"`
}

// GetAPILimit gets API limit count for the caller
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.4/user/getApiLimit.html
type GetAPILimit struct{}

// GetAPILimitResponse represents the limits towards the API call
type GetAPILimitResponse struct {
	APILimit APILimit `json:"apilimit"`
}
