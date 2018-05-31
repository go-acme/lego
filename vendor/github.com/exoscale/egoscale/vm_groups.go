package egoscale

// InstanceGroup represents a group of VM
type InstanceGroup struct {
	Account  string `json:"account,omitempty" doc:"the account owning the instance group"`
	Created  string `json:"created,omitempty" doc:"time and date the instance group was created"`
	Domain   string `json:"domain,omitempty" doc:"the domain name of the instance group"`
	DomainID string `json:"domainid,omitempty" doc:"the domain ID of the instance group"`
	ID       string `json:"id,omitempty" doc:"the id of the instance group"`
	Name     string `json:"name,omitempty" doc:"the name of the instance group"`
}

// InstanceGroupResponse represents a VM group
type InstanceGroupResponse struct {
	InstanceGroup InstanceGroup `json:"instancegroup"`
}

// CreateInstanceGroup creates a VM group
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/createInstanceGroup.html
type CreateInstanceGroup struct {
	Name     string `json:"name" doc:"the name of the instance group"`
	Account  string `json:"account,omitempty" doc:"the account of the instance group. The account parameter must be used with the domainId parameter."`
	DomainID string `json:"domainid,omitempty" doc:"the domain ID of account owning the instance group"`
}

// name returns the CloudStack API command name
func (*CreateInstanceGroup) name() string {
	return "createInstanceGroup"
}

func (*CreateInstanceGroup) response() interface{} {
	return new(CreateInstanceGroupResponse)
}

// CreateInstanceGroupResponse represents a freshly created VM group
type CreateInstanceGroupResponse InstanceGroupResponse

// UpdateInstanceGroup updates a VM group
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/updateInstanceGroup.html
type UpdateInstanceGroup struct {
	ID   string `json:"id" doc:"Instance group ID"`
	Name string `json:"name,omitempty" doc:"new instance group name"`
}

// name returns the CloudStack API command name
func (*UpdateInstanceGroup) name() string {
	return "updateInstanceGroup"
}

func (*UpdateInstanceGroup) response() interface{} {
	return new(UpdateInstanceGroupResponse)
}

// UpdateInstanceGroupResponse represents an updated VM group
type UpdateInstanceGroupResponse InstanceGroupResponse

// DeleteInstanceGroup deletes a VM group
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/deleteInstanceGroup.html
type DeleteInstanceGroup struct {
	ID string `json:"id" doc:"the ID of the instance group"`
}

// name returns the CloudStack API command name
func (*DeleteInstanceGroup) name() string {
	return "deleteInstanceGroup"
}

func (*DeleteInstanceGroup) response() interface{} {
	return new(booleanResponse)
}

// ListInstanceGroups lists VM groups
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listInstanceGroups.html
type ListInstanceGroups struct {
	Account     string `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	DomainID    string `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	ID          string `json:"id,omitempty" doc:"list instance groups by ID"`
	IsRecursive *bool  `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword     string `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll     *bool  `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Name        string `json:"name,omitempty" doc:"list instance groups by name"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"pagesize,omitempty"`
}

// name returns the CloudStack API command name
func (*ListInstanceGroups) name() string {
	return "listInstanceGroups"
}

func (*ListInstanceGroups) response() interface{} {
	return new(ListInstanceGroupsResponse)
}

// ListInstanceGroupsResponse represents a list of instance groups
type ListInstanceGroupsResponse struct {
	Count         int             `json:"count"`
	InstanceGroup []InstanceGroup `json:"instancegroup"`
}
