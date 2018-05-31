package egoscale

// AffinityGroup represents an (anti-)affinity group
//
// Affinity and Anti-Affinity groups provide a way to influence where VMs should run.
// See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/stable/virtual_machines.html#affinity-groups
type AffinityGroup struct {
	Account           string   `json:"account,omitempty" doc:"the account owning the affinity group"`
	Description       string   `json:"description,omitempty" doc:"the description of the affinity group"`
	Domain            string   `json:"domain,omitempty" doc:"the domain name of the affinity group"`
	DomainID          string   `json:"domainid,omitempty" doc:"the domain ID of the affinity group"`
	ID                string   `json:"id,omitempty" doc:"the ID of the affinity group"`
	Name              string   `json:"name,omitempty" doc:"the name of the affinity group"`
	Type              string   `json:"type,omitempty" doc:"the type of the affinity group"`
	VirtualMachineIDs []string `json:"virtualmachineIds,omitempty" doc:"virtual machine Ids associated with this affinity group "`
}

// AffinityGroupType represent an affinity group type
type AffinityGroupType struct {
	Type string `json:"type,omitempty" doc:"the type of the affinity group"`
}

// CreateAffinityGroup (Async) represents a new (anti-)affinity group
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/createAffinityGroup.html
type CreateAffinityGroup struct {
	Account     string `json:"account,omitempty" doc:"an account for the affinity group. Must be used with domainId."`
	Description string `json:"description,omitempty" doc:"optional description of the affinity group"`
	DomainID    string `json:"domainid,omitempty" doc:"domainId of the account owning the affinity group"`
	Name        string `json:"name" doc:"name of the affinity group"`
	Type        string `json:"type" doc:"Type of the affinity group from the available affinity/anti-affinity group types"`
}

// CreateAffinityGroupResponse represents the response of the creation of an (anti-)affinity group
type CreateAffinityGroupResponse struct {
	AffinityGroup AffinityGroup `json:"affinitygroup"`
}

// UpdateVMAffinityGroup (Async) represents a modification of a (anti-)affinity group
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/updateVMAffinityGroup.html
type UpdateVMAffinityGroup struct {
	ID                 string   `json:"id" doc:"The ID of the virtual machine"`
	AffinityGroupIDs   []string `json:"affinitygroupids,omitempty" doc:"comma separated list of affinity groups id that are going to be applied to the virtual machine. Should be passed only when vm is created from a zone with Basic Network support. Mutually exclusive with securitygroupnames parameter"`
	AffinityGroupNames []string `json:"affinitygroupnames,omitempty" doc:"comma separated list of affinity groups names that are going to be applied to the virtual machine. Should be passed only when vm is created from a zone with Basic Network support. Mutually exclusive with securitygroupids parameter"`
}

// UpdateVMAffinityGroupResponse represents the new VM
type UpdateVMAffinityGroupResponse VirtualMachineResponse

// DeleteAffinityGroup (Async) represents an (anti-)affinity group to be deleted
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/deleteAffinityGroup.html
type DeleteAffinityGroup struct {
	Account  string `json:"account,omitempty" doc:"the account of the affinity group. Must be specified with domain ID"`
	DomainID string `json:"domainid,omitempty" doc:"the domain ID of account owning the affinity group"`
	ID       string `json:"id,omitempty" doc:"The ID of the affinity group. Mutually exclusive with name parameter"`
	Name     string `json:"name,omitempty" doc:"The name of the affinity group. Mutually exclusive with ID parameter"`
}

// ListAffinityGroups represents an (anti-)affinity groups search
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listAffinityGroups.html
type ListAffinityGroups struct {
	Account          string `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	DomainID         string `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	ID               string `json:"id,omitempty" doc:"list the affinity group by the ID provided"`
	IsRecursive      *bool  `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword          string `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll          *bool  `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Name             string `json:"name,omitempty" doc:"lists affinity groups by name"`
	Page             int    `json:"page,omitempty"`
	PageSize         int    `json:"pagesize,omitempty"`
	Type             string `json:"type,omitempty" doc:"lists affinity groups by type"`
	VirtualMachineID string `json:"virtualmachineid,omitempty" doc:"lists affinity groups by virtual machine ID"`
}

// ListAffinityGroupsResponse represents a list of (anti-)affinity groups
type ListAffinityGroupsResponse struct {
	Count         int             `json:"count"`
	AffinityGroup []AffinityGroup `json:"affinitygroup"`
}

// ListAffinityGroupTypes represents an (anti-)affinity groups search
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listAffinityGroupTypes.html
type ListAffinityGroupTypes struct {
	Keyword  string `json:"keyword,omitempty" doc:"List by keyword"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"pagesize,omitempty"`
}

// ListAffinityGroupTypesResponse represents a list of (anti-)affinity group types
type ListAffinityGroupTypesResponse struct {
	Count             int                 `json:"count"`
	AffinityGroupType []AffinityGroupType `json:"affinitygrouptype"`
}
