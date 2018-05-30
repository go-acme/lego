package egoscale

import (
	"net"
)

// IPAddress represents an IP Address
type IPAddress struct {
	Account                   string        `json:"account,omitempty" doc:"the account the public IP address is associated with"`
	Allocated                 string        `json:"allocated,omitempty" doc:"date the public IP address was acquired"`
	Associated                string        `json:"associated,omitempty" doc:"date the public IP address was associated"`
	AssociatedNetworkID       string        `json:"associatednetworkid,omitempty" doc:"the ID of the Network associated with the IP address"`
	AssociatedNetworkName     string        `json:"associatednetworkname,omitempty" doc:"the name of the Network associated with the IP address"`
	Domain                    string        `json:"domain,omitempty" doc:"the domain the public IP address is associated with"`
	DomainID                  string        `json:"domainid,omitempty" doc:"the domain ID the public IP address is associated with"`
	ForDisplay                bool          `json:"fordisplay,omitempty" doc:"is public ip for display to the regular user"`
	ForVirtualNetwork         bool          `json:"forvirtualnetwork,omitempty" doc:"the virtual network for the IP address"`
	ID                        string        `json:"id,omitempty" doc:"public IP address id"`
	IPAddress                 net.IP        `json:"ipaddress,omitempty" doc:"public IP address"`
	IsElastic                 bool          `json:"iselastic,omitempty" doc:"is an elastic ip"`
	IsPortable                bool          `json:"isportable,omitempty" doc:"is public IP portable across the zones"`
	IsSourceNat               bool          `json:"issourcenat,omitempty" doc:"true if the IP address is a source nat address, false otherwise"`
	IsStaticNat               *bool         `json:"isstaticnat,omitempty" doc:"true if this ip is for static nat, false otherwise"`
	IsSystem                  bool          `json:"issystem,omitempty" doc:"true if this ip is system ip (was allocated as a part of deployVm or createLbRule)"`
	NetworkID                 string        `json:"networkid,omitempty" doc:"the ID of the Network where ip belongs to"`
	PhysicalNetworkID         string        `json:"physicalnetworkid,omitempty" doc:"the physical network this belongs to"`
	Purpose                   string        `json:"purpose,omitempty" doc:"purpose of the IP address. In Acton this value is not null for Ips with isSystem=true, and can have either StaticNat or LB value"`
	State                     string        `json:"state,omitempty" doc:"State of the ip address. Can be: Allocatin, Allocated and Releasing"`
	Tags                      []ResourceTag `json:"tags,omitempty" doc:"the list of resource tags associated with ip address"`
	VirtualMachineDisplayName string        `json:"virtualmachinedisplayname,omitempty" doc:"virtual machine display name the ip address is assigned to (not null only for static nat Ip)"`
	VirtualMachineID          string        `json:"virtualmachineid,omitempty" doc:"virtual machine id the ip address is assigned to (not null only for static nat Ip)"`
	VirtualMachineName        string        `json:"virtualmachinename,omitempty" doc:"virtual machine name the ip address is assigned to (not null only for static nat Ip)"`
	VlanID                    string        `json:"vlanid,omitempty" doc:"the ID of the VLAN associated with the IP address. This parameter is visible to ROOT admins only"`
	VlanName                  string        `json:"vlanname,omitempty" doc:"the VLAN associated with the IP address"`
	VMIPAddress               net.IP        `json:"vmipaddress,omitempty" doc:"virutal machine (dnat) ip address (not null only for static nat Ip)"`
	ZoneID                    string        `json:"zoneid,omitempty" doc:"the ID of the zone the public IP address belongs to"`
	ZoneName                  string        `json:"zonename,omitempty" doc:"the name of the zone the public IP address belongs to"`
}

// AssociateIPAddress (Async) represents the IP creation
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/associateIpAddress.html
type AssociateIPAddress struct {
	Account    string `json:"account,omitempty" doc:"the account to associate with this IP address"`
	DomainID   string `json:"domainid,omitempty" doc:"the ID of the domain to associate with this IP address"`
	ForDisplay *bool  `json:"fordisplay,omitempty" doc:"an optional field, whether to the display the ip to the end user or not"`
	IsPortable *bool  `json:"isportable,omitempty" doc:"should be set to true if public IP is required to be transferable across zones, if not specified defaults to false"`
	NetworkdID string `json:"networkid,omitempty" doc:"The network this ip address should be associated to."`
	RegionID   int    `json:"regionid,omitempty" doc:"region ID from where portable ip is to be associated."`
	ZoneID     string `json:"zoneid,omitempty" doc:"the ID of the availability zone you want to acquire an public IP address from"`
}

// AssociateIPAddressResponse represents the response to the creation of an IPAddress
type AssociateIPAddressResponse struct {
	IPAddress IPAddress `json:"ipaddress"`
}

// DisassociateIPAddress (Async) represents the IP deletion
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/disassociateIpAddress.html
type DisassociateIPAddress struct {
	ID string `json:"id" doc:"the id of the public ip address to disassociate"`
}

// UpdateIPAddress (Async) represents the IP modification
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/updateIpAddress.html
type UpdateIPAddress struct {
	ID         string `json:"id" doc:"the id of the public ip address to update"`
	CustomID   string `json:"customid,omitempty" doc:"an optional field, in case you want to set a custom id to the resource. Allowed to Root Admins only"`
	ForDisplay *bool  `json:"fordisplay,omitempty" doc:"an optional field, whether to the display the ip to the end user or not"`
}

// UpdateIPAddressResponse represents the modified IP Address
type UpdateIPAddressResponse AssociateIPAddressResponse

// ListPublicIPAddresses represents a search for public IP addresses
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/listPublicIpAddresses.html
type ListPublicIPAddresses struct {
	Account             string        `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	AllocatedOnly       *bool         `json:"allocatedonly,omitempty" doc:"limits search results to allocated public IP addresses"`
	AssociatedNetworkID string        `json:"associatednetworkid,omitempty" doc:"lists all public IP addresses associated to the network specified"`
	DomainID            string        `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	ForDisplay          *bool         `json:"fordisplay,omitempty" doc:"list resources by display flag; only ROOT admin is eligible to pass this parameter"`
	ForLoadBalancing    *bool         `json:"forloadbalancing,omitempty" doc:"list only ips used for load balancing"`
	ForVirtualNetwork   *bool         `json:"forvirtualnetwork,omitempty" doc:"the virtual network for the IP address"`
	ID                  string        `json:"id,omitempty" doc:"lists ip address by id"`
	IPAddress           net.IP        `json:"ipaddress,omitempty" doc:"lists the specified IP address"`
	IsElastic           *bool         `json:"iselastic,omitempty" doc:"list only elastic ip addresses"`
	IsRecursive         *bool         `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	IsSourceNat         *bool         `json:"issourcenat,omitempty" doc:"list only source nat ip addresses"`
	IsStaticNat         *bool         `json:"isstaticnat,omitempty" doc:"list only static nat ip addresses"`
	Keyword             string        `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll             *bool         `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Page                int           `json:"page,omitempty"`
	PageSize            int           `json:"pagesize,omitempty"`
	PhysicalNetworkID   string        `json:"physicalnetworkid,omitempty" doc:"lists all public IP addresses by physical network id"`
	Tags                []ResourceTag `json:"tags,omitempty" doc:"List resources by tags (key/value pairs)"`
	VlanID              string        `json:"vlanid,omitempty" doc:"lists all public IP addresses by VLAN ID"`
	ZoneID              string        `json:"zoneid,omitempty" doc:"lists all public IP addresses by Zone ID"`
}

// ListPublicIPAddressesResponse represents a list of public IP addresses
type ListPublicIPAddressesResponse struct {
	Count           int         `json:"count"`
	PublicIPAddress []IPAddress `json:"publicipaddress"`
}
