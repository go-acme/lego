package egoscale

import (
	"net"
)

// Nic represents a Network Interface Controller (NIC)
//
// See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/latest/networking_and_traffic.html#configuring-multiple-ip-addresses-on-a-single-nic
type Nic struct {
	BroadcastURI     string           `json:"broadcasturi,omitempty" doc:"the broadcast uri of the nic"`
	DeviceID         string           `json:"deviceid,omitempty" doc:"device id for the network when plugged into the virtual machine"`
	Gateway          net.IP           `json:"gateway,omitempty" doc:"the gateway of the nic"`
	ID               string           `json:"id,omitempty" doc:"the ID of the nic"`
	IP6Address       net.IP           `json:"ip6address,omitempty" doc:"the IPv6 address of network"`
	IP6Cidr          string           `json:"ip6cidr,omitempty" doc:"the cidr of IPv6 network"`
	IP6Gateway       net.IP           `json:"ip6gateway,omitempty" doc:"the gateway of IPv6 network"`
	IPAddress        net.IP           `json:"ipaddress,omitempty" doc:"the ip address of the nic"`
	IsDefault        bool             `json:"isdefault,omitempty" doc:"true if nic is default, false otherwise"`
	IsolationURI     string           `json:"isolationuri,omitempty" doc:"the isolation uri of the nic"`
	MacAddress       string           `json:"macaddress,omitempty" doc:"true if nic is default, false otherwise"`
	Netmask          net.IP           `json:"netmask,omitempty" doc:"the netmask of the nic"`
	NetworkID        string           `json:"networkid,omitempty" doc:"the ID of the corresponding network"`
	NetworkName      string           `json:"networkname,omitempty" doc:"the name of the corresponding network"`
	SecondaryIP      []NicSecondaryIP `json:"secondaryip,omitempty" doc:"the Secondary ipv4 addr of nic"`
	TrafficType      string           `json:"traffictype,omitempty" doc:"the traffic type of the nic"`
	Type             string           `json:"type,omitempty" doc:"the type of the nic"`
	VirtualMachineID string           `json:"virtualmachineid,omitempty" doc:"Id of the vm to which the nic belongs"`
}

// NicSecondaryIP represents a link between NicID and IPAddress
type NicSecondaryIP struct {
	ID               string `json:"id,omitempty" doc:"the ID of the secondary private IP addr"`
	IPAddress        net.IP `json:"ipaddress,omitempty" doc:"Secondary IP address"`
	NetworkID        string `json:"networkid,omitempty" doc:"the ID of the network"`
	NicID            string `json:"nicid,omitempty" doc:"the ID of the nic"`
	VirtualMachineID string `json:"virtualmachineid,omitempty" doc:"the ID of the vm"`
}

// ListNics represents the NIC search
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listNics.html
type ListNics struct {
	ForDisplay       bool   `json:"fordisplay,omitempty" doc:"list resources by display flag; only ROOT admin is eligible to pass this parameter"`
	Keyword          string `json:"keyword,omitempty" doc:"List by keyword"`
	NetworkID        string `json:"networkid,omitempty" doc:"list nic of the specific vm's network"`
	NicID            string `json:"nicid,omitempty" doc:"the ID of the nic to to list IPs"`
	Page             int    `json:"page,omitempty"`
	PageSize         int    `json:"pagesize,omitempty"`
	VirtualMachineID string `json:"virtualmachineid" doc:"the ID of the vm"`
}

// ListNicsResponse represents a list of templates
type ListNicsResponse struct {
	Count int   `json:"count"`
	Nic   []Nic `json:"nic"`
}

// AddIPToNic (Async) represents the assignation of a secondary IP
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/addIpToNic.html
type AddIPToNic struct {
	NicID     string `json:"nicid" doc:"the ID of the nic to which you want to assign private IP"`
	IPAddress net.IP `json:"ipaddress,omitempty" doc:"Secondary IP Address"`
}

// AddIPToNicResponse represents the addition of an IP to a NIC
type AddIPToNicResponse struct {
	NicSecondaryIP NicSecondaryIP `json:"nicsecondaryip"`
}

// RemoveIPFromNic (Async) represents a deletion request
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/removeIpFromNic.html
type RemoveIPFromNic struct {
	ID string `json:"id" doc:"the ID of the secondary ip address to nic"`
}

// ActivateIP6 (Async) activates the IP6 on the given NIC
//
// Exoscale specific API: https://community.exoscale.ch/api/compute/#activateip6_GET
type ActivateIP6 struct {
	NicID string `json:"nicid" doc:"the ID of the nic to which you want to assign the IPv6"`
}

// ActivateIP6Response represents the modified NIC
type ActivateIP6Response struct {
	Nic Nic `json:"nic"`
}
