package internal

import "fmt"

type ResourceRecord struct {
	ErrorCode string `json:"errno,omitempty"`

	DelayedCreateTime                       string `json:"delayed_create_time,omitempty"`
	DelayedDeleteTime                       string `json:"delayed_delete_time,omitempty"`
	DelayedTime                             string `json:"delayed_time,omitempty"`
	DNSClassName                            string `json:"dns_class_name,omitempty"`
	DNSCloud                                string `json:"dns_cloud,omitempty"`
	DNSComment                              string `json:"dns_comment,omitempty"`
	DNSID                                   string `json:"dns_id,omitempty"`
	DNSName                                 string `json:"dns_name,omitempty"`
	DNSType                                 string `json:"dns_type,omitempty"`
	DNSVersion                              string `json:"dns_version,omitempty"`
	DNSViewClassName                        string `json:"dnsview_class_name,omitempty"`
	DNSViewClassParameters                  string `json:"dnsview_class_parameters,omitempty"`
	DNSViewClassParametersInheritanceSource string `json:"dnsview_class_parameters_inheritance_source,omitempty"`
	DNSViewClassParametersProperties        string `json:"dnsview_class_parameters_properties,omitempty"`
	DNSViewID                               string `json:"dnsview_id,omitempty"`
	DNSViewName                             string `json:"dnsview_name,omitempty"`
	DNSZoneClassName                        string `json:"dnszone_class_name,omitempty"`
	DNSZoneForwarders                       string `json:"dnszone_forwarders,omitempty"`
	DNSZoneID                               string `json:"dnszone_id,omitempty"`
	DNSZoneIsReverse                        string `json:"dnszone_is_reverse,omitempty"`
	DNSZoneIsRpz                            string `json:"dnszone_is_rpz,omitempty"`
	DNSZoneMasters                          string `json:"dnszone_masters,omitempty"`
	DNSZoneName                             string `json:"dnszone_name,omitempty"`
	DNSZoneNameUTF                          string `json:"dnszone_name_utf,omitempty"`
	DNSZoneSiteName                         string `json:"dnszone_site_name,omitempty"`
	DNSZoneSortZone                         string `json:"dnszone_sort_zone,omitempty"`
	DNSZoneType                             string `json:"dnszone_type,omitempty"`
	MultiStatus                             string `json:"multistatus,omitempty"`
	RRAllValue                              string `json:"rr_all_value,omitempty"`
	RRAuthGsstsig                           string `json:"rr_auth_gsstsig,omitempty"`
	RRClassName                             string `json:"rr_class_name,omitempty"`
	RRClassParameters                       string `json:"rr_class_parameters,omitempty"`
	RRClassParametersInheritanceSource      string `json:"rr_class_parameters_inheritance_source,omitempty"`
	RRClassParametersProperties             string `json:"rr_class_parameters_properties,omitempty"`
	RRFullName                              string `json:"rr_full_name,omitempty"`
	RRFullNameUTF                           string `json:"rr_full_name_utf,omitempty"`
	RRGlue                                  string `json:"rr_glue,omitempty"`
	RRGlueID                                string `json:"rr_glue_id,omitempty"`
	RRID                                    string `json:"rr_id,omitempty"`
	RRLastUpdateDays                        string `json:"rr_last_update_days,omitempty"`
	RRLastUpdateTime                        string `json:"rr_last_update_time,omitempty"`
	RRName                                  string `json:"rr_name,omitempty"`
	RRNameID                                string `json:"rr_name_id,omitempty"`
	RRNameIP4Addr                           string `json:"rr_name_ip4_addr,omitempty"`
	RRNameIPAddr                            string `json:"rr_name_ip_addr,omitempty"`
	RRType                                  string `json:"rr_type,omitempty"`
	RRTypeID                                string `json:"rr_type_id,omitempty"`
	RRValueID                               string `json:"rr_value_id,omitempty"`
	RRValueIP4Addr                          string `json:"rr_value_ip4_addr,omitempty"`
	RRValueIPAddr                           string `json:"rr_value_ip_addr,omitempty"`
	TTL                                     string `json:"ttl,omitempty"`
	Value1                                  string `json:"value1,omitempty"`
	Value2                                  string `json:"value2,omitempty"`
	Value3                                  string `json:"value3,omitempty"`
	Value4                                  string `json:"value4,omitempty"`
	Value5                                  string `json:"value5,omitempty"`
	Value6                                  string `json:"value6,omitempty"`
	Value7                                  string `json:"value7,omitempty"`
	VDNSParentID                            string `json:"vdns_parent_id,omitempty"`
	VDNSParentName                          string `json:"vdns_parent_name,omitempty"`
}

type DeleteInputParameters struct {
	RRID        string `url:"rr_id,omitempty"`
	DNSName     string `url:"dns_name,omitempty"`
	DNSViewName string `url:"dnsview_name,omitempty"`
	RRName      string `url:"rr_name,omitempty"`
	RRType      string `url:"rr_type,omitempty"`
	RRValue1    string `url:"rr_value1,omitempty"`
}

type BaseOutput struct {
	RetOID string `json:"ret_oid,omitempty"`
}

type APIError struct {
	ErrorCode   string `json:"errno,omitempty"`
	ErrMsg      string `json:"errmsg,omitempty"`
	Severity    string `json:"severity,omitempty"`
	Category    string `json:"category,omitempty"`
	Parameters  string `json:"parameters,omitempty"`
	ParamFormat string `json:"param_format,omitempty"`
	ParamValue  string `json:"param_value,omitempty"`
}

func (a APIError) Error() string {
	msg := fmt.Sprintf("%s: %s %s %s", a.Category, a.Severity, a.ErrorCode, a.ErrMsg)

	if a.Parameters != "" {
		msg += fmt.Sprintf(" parameters: %s", a.Parameters)
	}

	if a.ParamFormat != "" {
		msg += fmt.Sprintf(" param_format: %s", a.ParamFormat)
	}

	if a.ParamValue != "" {
		msg += fmt.Sprintf(" param_value: %s", a.ParamValue)
	}

	return msg
}
