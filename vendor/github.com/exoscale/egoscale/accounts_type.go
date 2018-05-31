package egoscale

// AccountType represents the type of an Account
//
// http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/4.8/accounts.html#accounts-users-and-domains
type AccountType int16

//go:generate stringer -type AccountType
const (
	// UserAccount represents a User
	UserAccount AccountType = iota
	// AdminAccount represents an Admin
	AdminAccount
	// DomainAdminAccount represents a Domain Admin
	DomainAdminAccount
)

// Account provides the detailed account information
type Account struct {
	AccountDetails            map[string]string `json:"accountdetails,omitempty" doc:"details for the account"`
	AccountType               AccountType       `json:"accounttype,omitempty" doc:"account type (admin, domain-admin, user)"`
	CPUAvailable              string            `json:"cpuavailable,omitempty" doc:"the total number of cpu cores available to be created for this account"`
	CPULimit                  string            `json:"cpulimit,omitempty" doc:"the total number of cpu cores the account can own"`
	CPUTotal                  int64             `json:"cputotal,omitempty" doc:"the total number of cpu cores owned by account"`
	DefaultZoneID             string            `json:"defaultzoneid,omitempty" doc:"the default zone of the account"`
	Domain                    string            `json:"domain,omitempty" doc:"name of the Domain the account belongs too"`
	DomainID                  string            `json:"domainid,omitempty" doc:"id of the Domain the account belongs too"`
	EipLimit                  string            `json:"eiplimit,omitempty" doc:"the total number of public elastic ip addresses this account can acquire"`
	Groups                    []string          `json:"groups,omitempty" doc:"the list of acl groups that account belongs to"`
	ID                        string            `json:"id,omitempty" doc:"the id of the account"`
	IPAvailable               string            `json:"ipavailable,omitempty" doc:"the total number of public ip addresses available for this account to acquire"`
	IPLimit                   string            `json:"iplimit,omitempty" doc:"the total number of public ip addresses this account can acquire"`
	IPTotal                   int64             `json:"iptotal,omitempty" doc:"the total number of public ip addresses allocated for this account"`
	IsCleanupRequired         bool              `json:"iscleanuprequired,omitempty" doc:"true if the account requires cleanup"`
	IsDefault                 bool              `json:"isdefault,omitempty" doc:"true if account is default, false otherwise"`
	MemoryAvailable           string            `json:"memoryavailable,omitempty" doc:"the total memory (in MB) available to be created for this account"`
	MemoryLimit               string            `json:"memorylimit,omitempty" doc:"the total memory (in MB) the account can own"`
	MemoryTotal               int64             `json:"memorytotal,omitempty" doc:"the total memory (in MB) owned by account"`
	Name                      string            `json:"name,omitempty" doc:"the name of the account"`
	NetworkAvailable          string            `json:"networkavailable,omitempty" doc:"the total number of networks available to be created for this account"`
	NetworkDomain             string            `json:"networkdomain,omitempty" doc:"the network domain"`
	NetworkLimit              string            `json:"networklimit,omitempty" doc:"the total number of networks the account can own"`
	NetworkTotal              int64             `json:"networktotal,omitempty" doc:"the total number of networks owned by account"`
	PrimaryStorageAvailable   string            `json:"primarystorageavailable,omitempty" doc:"the total primary storage space (in GiB) available to be used for this account"`
	PrimaryStorageLimit       string            `json:"primarystoragelimit,omitempty" doc:"the total primary storage space (in GiB) the account can own"`
	PrimaryStorageTotal       int64             `json:"primarystoragetotal,omitempty" doc:"the total primary storage space (in GiB) owned by account"`
	ProjectAvailable          string            `json:"projectavailable,omitempty" doc:"the total number of projects available for administration by this account"`
	ProjectLimit              string            `json:"projectlimit,omitempty" doc:"the total number of projects the account can own"`
	ProjectTotal              int64             `json:"projecttotal,omitempty" doc:"the total number of projects being administrated by this account"`
	SecondaryStorageAvailable string            `json:"secondarystorageavailable,omitempty" doc:"the total secondary storage space (in GiB) available to be used for this account"`
	SecondaryStorageLimit     string            `json:"secondarystoragelimit,omitempty" doc:"the total secondary storage space (in GiB) the account can own"`
	SecondaryStorageTotal     int64             `json:"secondarystoragetotal,omitempty" doc:"the total secondary storage space (in GiB) owned by account"`
	SMTP                      bool              `json:"smtp,omitempty" doc:"if SMTP outbound is allowed"`
	SnapshotAvailable         string            `json:"snapshotavailable,omitempty" doc:"the total number of snapshots available for this account"`
	SnapshotLimit             string            `json:"snapshotlimit,omitempty" doc:"the total number of snapshots which can be stored by this account"`
	SnapshotTotal             int64             `json:"snapshottotal,omitempty" doc:"the total number of snapshots stored by this account"`
	State                     string            `json:"state,omitempty" doc:"the state of the account"`
	TemplateAvailable         string            `json:"templateavailable,omitempty" doc:"the total number of templates available to be created by this account"`
	TemplateLimit             string            `json:"templatelimit,omitempty" doc:"the total number of templates which can be created by this account"`
	TemplateTotal             int64             `json:"templatetotal,omitempty" doc:"the total number of templates which have been created by this account"`
	User                      []User            `json:"user,omitempty" doc:"the list of users associated with account"`
	VMAvailable               string            `json:"vmavailable,omitempty" doc:"the total number of virtual machines available for this account to acquire"`
	VMLimit                   string            `json:"vmlimit,omitempty" doc:"the total number of virtual machines that can be deployed by this account"`
	VMRunning                 int               `json:"vmrunning,omitempty" doc:"the total number of virtual machines running for this account"`
	VMStopped                 int               `json:"vmstopped,omitempty" doc:"the total number of virtual machines stopped for this account"`
	VMTotal                   int64             `json:"vmtotal,omitempty" doc:"the total number of virtual machines deployed by this account"`
	VolumeAvailable           string            `json:"volumeavailable,omitempty" doc:"the total volume available for this account"`
	VolumeLimit               string            `json:"volumelimit,omitempty" doc:"the total volume which can be used by this account"`
	VolumeTotal               int64             `json:"volumetotal,omitempty" doc:"the total volume being used by this account"`
}

// ListAccounts represents a query to display the accounts
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listAccounts.html
type ListAccounts struct {
	AccountType       AccountType `json:"accounttype,omitempty" doc:"list accounts by account type. Valid account types are 1 (admin), 2 (domain-admin), and 0 (user)."`
	DomainID          string      `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	ID                string      `json:"id,omitempty" doc:"list account by account ID"`
	IsCleanUpRequired *bool       `json:"iscleanuprequired,omitempty" doc:"list accounts by cleanuprequred attribute (values are true or false)"`
	IsRecursive       *bool       `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword           string      `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll           *bool       `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Name              string      `json:"name,omitempty" doc:"list account by account name"`
	Page              int         `json:"page,omitempty"`
	PageSize          int         `json:"pagesize,omitempty"`
	State             string      `json:"state,omitempty" doc:"list accounts by state. Valid states are enabled, disabled, and locked."`
}

// ListAccountsResponse represents a list of accounts
type ListAccountsResponse struct {
	Count   int       `json:"count"`
	Account []Account `json:"account"`
}

// EnableAccount represents the activation of an account
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/enableAccount.html
type EnableAccount struct {
	Account  string `json:"account,omitempty" doc:"Enables specified account."`
	DomainID string `json:"domainid,omitempty" doc:"Enables specified account in this domain."`
	ID       string `json:"id,omitempty" doc:"Account id"`
}

// EnableAccountResponse represents the modified account
type EnableAccountResponse struct {
	Account Account `json:"account"`
}

// DisableAccount (Async) represents the deactivation of an account
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/disableAccount.html
type DisableAccount struct {
	Lock     *bool  `json:"lock" doc:"If true, only lock the account; else disable the account"`
	Account  string `json:"account,omitempty" doc:"Disables specified account."`
	DomainID string `json:"domainid,omitempty" doc:"Disables specified account in this domain."`
	ID       string `json:"id,omitempty" doc:"Account id"`
}

// DisableAccountResponse represents the modified account
type DisableAccountResponse EnableAccountResponse
