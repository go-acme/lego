package egoscale

// Template represents a machine to be deployed
//
// See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/latest/templates.html
type Template struct {
	Account               string            `json:"account,omitempty" doc:"the account name to which the template belongs"`
	AccountID             string            `json:"accountid,omitempty" doc:"the account id to which the template belongs"`
	Bootable              bool              `json:"bootable,omitempty" doc:"true if the ISO is bootable, false otherwise"`
	Checksum              string            `json:"checksum,omitempty" doc:"checksum of the template"`
	Created               string            `json:"created,omitempty" doc:"the date this template was created"`
	CrossZones            bool              `json:"crossZones,omitempty" doc:"true if the template is managed across all Zones, false otherwise"`
	Details               map[string]string `json:"details,omitempty" doc:"additional key/value details tied with template"`
	DisplayText           string            `json:"displaytext,omitempty" doc:"the template display text"`
	Domain                string            `json:"domain,omitempty" doc:"the name of the domain to which the template belongs"`
	DomainID              string            `json:"domainid,omitempty" doc:"the ID of the domain to which the template belongs"`
	Format                string            `json:"format,omitempty" doc:"the format of the template."`
	HostID                string            `json:"hostid,omitempty" doc:"the ID of the secondary storage host for the template"`
	HostName              string            `json:"hostname,omitempty" doc:"the name of the secondary storage host for the template"`
	Hypervisor            string            `json:"hypervisor,omitempty" doc:"the hypervisor on which the template runs"`
	ID                    string            `json:"id,omitempty" doc:"the template ID"`
	IsDynamicallyScalable bool              `json:"isdynamicallyscalable,omitempty" doc:"true if template contains XS/VMWare tools inorder to support dynamic scaling of VM cpu/memory"`
	IsExtractable         bool              `json:"isextractable,omitempty" doc:"true if the template is extractable, false otherwise"`
	IsFeatured            bool              `json:"isfeatured,omitempty" doc:"true if this template is a featured template, false otherwise"`
	IsPublic              bool              `json:"ispublic,omitempty" doc:"true if this template is a public template, false otherwise"`
	IsReady               bool              `json:"isready,omitempty" doc:"true if the template is ready to be deployed from, false otherwise."`
	Name                  string            `json:"name,omitempty" doc:"the template name"`
	OsTypeID              string            `json:"ostypeid,omitempty" doc:"the ID of the OS type for this template."`
	OsTypeName            string            `json:"ostypename,omitempty" doc:"the name of the OS type for this template."`
	PasswordEnabled       bool              `json:"passwordenabled,omitempty" doc:"true if the reset password feature is enabled, false otherwise"`
	Removed               string            `json:"removed,omitempty" doc:"the date this template was removed"`
	Size                  int64             `json:"size,omitempty" doc:"the size of the template"`
	SourceTemplateID      string            `json:"sourcetemplateid,omitempty" doc:"the template ID of the parent template if present"`
	SSHKeyEnabled         bool              `json:"sshkeyenabled,omitempty" doc:"true if template is sshkey enabled, false otherwise"`
	Status                string            `json:"status,omitempty" doc:"the status of the template"`
	Tags                  []ResourceTag     `json:"tags,omitempty" doc:"the list of resource tags associated with tempate"`
	TemplateDirectory     string            `json:"templatedirectory,omitempty" doc:"Template directory"`
	TemplateTag           string            `json:"templatetag,omitempty" doc:"the tag of this template"`
	TemplateType          string            `json:"templatetype,omitempty" doc:"the type of the template"`
	URL                   string            `json:"url,omitempty" doc:"Original URL of the template where it was downloaded"`
	ZoneID                string            `json:"zoneid,omitempty" doc:"the ID of the zone for this template"`
	ZoneName              string            `json:"zonename,omitempty" doc:"the name of the zone for this template"`
}

// ListTemplates represents a template query filter
//
// CloudStackAPI: http://cloudstack.apache.org/api/apidocs-4.10/apis/listTemplates.html
type ListTemplates struct {
	TemplateFilter string        `json:"templatefilter" doc:"possible values are \"featured\", \"self\", \"selfexecutable\",\"sharedexecutable\",\"executable\", and \"community\". * featured : templates that have been marked as featured and public. * self : templates that have been registered or created by the calling user. * selfexecutable : same as self, but only returns templates that can be used to deploy a new VM. * sharedexecutable : templates ready to be deployed that have been granted to the calling user by another user. * executable : templates that are owned by the calling user, or public templates, that can be used to deploy a VM. * community : templates that have been marked as public but not featured. * all : all templates (only usable by admins)."`
	Account        string        `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	DomainID       string        `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	Hypervisor     string        `json:"hypervisor,omitempty" doc:"the hypervisor for which to restrict the search"`
	ID             string        `json:"id,omitempty" doc:"the template ID"`
	IsRecursive    *bool         `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword        string        `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll        *bool         `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Name           string        `json:"name,omitempty" doc:"the template name"`
	Page           int           `json:"page,omitempty"`
	PageSize       int           `json:"pagesize,omitempty"`
	ShowRemoved    *bool         `json:"showremoved,omitempty" doc:"show removed templates as well"`
	Tags           []ResourceTag `json:"tags,omitempty" doc:"List resources by tags (key/value pairs)"`
	ZoneID         string        `json:"zoneid,omitempty" doc:"list templates by zoneId"`
}

// ListTemplatesResponse represents a list of templates
type ListTemplatesResponse struct {
	Count    int        `json:"count"`
	Template []Template `json:"template"`
}

// CreateTemplate (Async) represents a template creation
//
// CloudStackAPI: http://cloudstack.apache.org/api/apidocs-4.10/apis/createTemplate.html
type CreateTemplate struct {
	Bits                  int               `json:"bits,omitempty" doc:"32 or 64 bit"`
	Details               map[string]string `json:"details,omitempty" doc:"Template details in key/value pairs."`
	DisplayText           string            `json:"displaytext" doc:"the display text of the template. This is usually used for display purposes."`
	IsDynamicallyScalable *bool             `json:"isdynamicallyscalable,omitempty" doc:"true if template contains XS/VMWare tools inorder to support dynamic scaling of VM cpu/memory"`
	IsFeatured            *bool             `json:"isfeatured,omitempty" doc:"true if this template is a featured template, false otherwise"`
	IsPublic              *bool             `json:"ispublic,omitempty" doc:"true if this template is a public template, false otherwise"`
	Name                  string            `json:"name" doc:"the name of the template"`
	OsTypeID              string            `json:"ostypeid" doc:"the ID of the OS Type that best represents the OS of this template."`
	PasswordEnabled       *bool             `json:"passwordenabled,omitempty" doc:"true if the template supports the password reset feature; default is false"`
	RequiresHVM           *bool             `json:"requireshvm,omitempty" doc:"true if the template requres HVM, false otherwise"`
	SnapshotID            string            `json:"snapshotid,omitempty" doc:"the ID of the snapshot the template is being created from. Either this parameter, or volumeId has to be passed in"`
	TemplateTag           string            `json:"templatetag,omitempty" doc:"the tag for this template."`
	URL                   string            `json:"url,omitempty" doc:"Optional, only for baremetal hypervisor. The directory name where template stored on CIFS server"`
	VirtualMachineID      string            `json:"virtualmachineid,omitempty" doc:"Optional, VM ID. If this presents, it is going to create a baremetal template for VM this ID refers to. This is only for VM whose hypervisor type is BareMetal"`
	VolumeID              string            `json:"volumeid,omitempty" doc:"the ID of the disk volume the template is being created from. Either this parameter, or snapshotId has to be passed in"`
}

// CreateTemplateResponse represents a freshly created template
type CreateTemplateResponse struct {
	Template Template `json:"template"`
}

// CopyTemplate (Async) represents a template copy
//
// CloudStackAPI: http://cloudstack.apache.org/api/apidocs-4.10/apis/copyTemplate.html
type CopyTemplate struct {
	DestZoneID   string `json:"destzoneid" doc:"ID of the zone the template is being copied to."`
	ID           string `json:"id" doc:"Template ID."`
	SourceZoneID string `json:"sourcezoneid,omitempty" doc:"ID of the zone the template is currently hosted on. If not specified and template is cross-zone, then we will sync this template to region wide image store."`
}

// CopyTemplateResponse represents the copied template
type CopyTemplateResponse CreateTemplateResponse

// UpdateTemplate represents a template change
//
// CloudStackAPI: http://cloudstack.apache.org/api/apidocs-4.10/apis/updateTemplate.html
type UpdateTemplate struct {
	Bootable              *bool             `json:"bootable,omitempty" doc:"true if image is bootable, false otherwise"`
	Details               map[string]string `json:"details,omitempty" doc:"Details in key/value pairs."`
	DisplayText           string            `json:"displaytext,omitempty" doc:"the display text of the image"`
	Format                string            `json:"format,omitempty" doc:"the format for the image"`
	ID                    string            `json:"id" doc:"the ID of the image file"`
	IsDynamicallyScalable *bool             `json:"isdynamicallyscalable,omitempty" doc:"true if template/ISO contains XS/VMWare tools inorder to support dynamic scaling of VM cpu/memory"`
	IsRouting             *bool             `json:"isrouting,omitempty" doc:"true if the template type is routing i.e., if template is used to deploy router"`
	Name                  string            `json:"name,omitempty" doc:"the name of the image file"`
	OsTypeID              string            `json:"ostypeid,omitempty" doc:"the ID of the OS type that best represents the OS of this image."`
	PasswordEnabled       *bool             `json:"passwordenabled,omitempty" doc:"true if the image supports the password reset feature; default is false"`
	SortKey               int               `json:"sortkey,omitempty" doc:"sort key of the template, integer"`
}

// UpdateTemplateResponse represents the updated template
type UpdateTemplateResponse CreateTemplateResponse

// DeleteTemplate (Async) represents the deletion of a template
//
// CloudStackAPI: http://cloudstack.apache.org/api/apidocs-4.10/apis/deleteTemplate.html
type DeleteTemplate struct {
	ID     string `json:"id" doc:"the ID of the template"`
	ZoneID string `json:"zoneid,omitempty" doc:"the ID of zone of the template"`
}

// PrepareTemplate represents a template preparation
//
// CloudStackAPI: http://cloudstack.apache.org/api/apidocs-4.10/apis/prepareTemplate.html
type PrepareTemplate struct {
	TemplateID string `json:"templateid" doc:"template ID of the template to be prepared in primary storage(s)."`
	ZoneID     string `json:"zoneid" doc:"zone ID of the template to be prepared in primary storage(s)."`
}

// PrepareTemplateResponse represents the prepared template
type PrepareTemplateResponse CreateTemplateResponse

// RegisterTemplate represents a template registration
//
// CloudStackAPI: http://cloudstack.apache.org/api/apidocs-4.10/apis/registerTemplate.html
type RegisterTemplate struct {
	Account               string            `json:"account,omitempty" doc:"an optional accountName. Must be used with domainId."`
	Bits                  int               `json:"bits,omitempty" doc:"32 or 64 bits support. 64 by default"`
	Checksum              string            `json:"checksum,omitempty" doc:"the MD5 checksum value of this template"`
	Details               map[string]string `json:"details,omitempty" doc:"Template details in key/value pairs."`
	DisplayText           string            `json:"displaytext" doc:"the display text of the template. This is usually used for display purposes."`
	DomainID              string            `json:"domainid,omitempty" doc:"an optional domainId. If the account parameter is used, domainId must also be used."`
	Format                string            `json:"format" doc:"the format for the template. Possible values include QCOW2, RAW, and VHD."`
	Hypervisor            string            `json:"hypervisor" doc:"the target hypervisor for the template"`
	IsDynamicallyScalable *bool             `json:"isdynamicallyscalable,omitempty" doc:"true if template contains XS/VMWare tools inorder to support dynamic scaling of VM cpu/memory"`
	IsExtractable         *bool             `json:"isextractable,omitempty" doc:"true if the template or its derivatives are extractable; default is false"`
	IsFeatured            *bool             `json:"isfeatured,omitempty" doc:"true if this template is a featured template, false otherwise"`
	IsPublic              *bool             `json:"ispublic,omitempty" doc:"true if the template is available to all accounts; default is true"`
	IsRouting             *bool             `json:"isrouting,omitempty" doc:"true if the template type is routing i.e., if template is used to deploy router"`
	IsSystem              *bool             `json:"issystem,omitempty" doc:"true if the template type is system i.e., if template is used to deploy system VM"`
	Name                  string            `json:"name" doc:"the name of the template"`
	OsTypeID              string            `json:"ostypeid" doc:"the ID of the OS Type that best represents the OS of this template."`
	PasswordEnabled       *bool             `json:"passwordenabled,omitempty" doc:"true if the template supports the password reset feature; default is false"`
	RequiresHVM           *bool             `json:"requireshvm,omitempty" doc:"true if this template requires HVM"`
	SSHKeyEnabled         *bool             `json:"sshkeyenabled,omitempty" doc:"true if the template supports the sshkey upload feature; default is false"`
	TemplateTag           string            `json:"templatetag,omitempty" doc:"the tag for this template."`
	URL                   string            `json:"url" doc:"the URL of where the template is hosted. Possible URL include http:// and https://"`
	ZoneID                string            `json:"zoneid" doc:"the ID of the zone the template is to be hosted on"`
}

// RegisterTemplateResponse represents the registered template
type RegisterTemplateResponse CreateTemplateResponse
