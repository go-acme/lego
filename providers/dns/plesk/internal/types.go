package internal

import (
	"encoding/xml"
	"fmt"
)

// Response status.
const (
	StatusOK    = "ok"
	StatusError = "error"
)

// Request.

type RequestPacketType struct {
	XMLName xml.Name `xml:"packet"`
	Text    string   `xml:",chardata"`

	DNS  *DNSInputType    `xml:"dns,omitempty"`
	Site *SiteTypeRequest `xml:"site,omitempty"`
}

type DNSInputType struct {
	Text string `xml:",chardata"`

	AddRec []AddRecRequest `xml:"add_rec,omitempty"`
	DelRec []DelRecRequest `xml:"del_rec,omitempty"`
}

type AddRecRequest struct {
	Text string `xml:",chardata"`

	SiteID int    `xml:"site-id,omitempty"`
	Type   string `xml:"type,omitempty"`
	Host   string `xml:"host,omitempty"`
	Value  string `xml:"value,omitempty"`
}

type DelRecRequest struct {
	Text string `xml:",chardata"`

	Filter DNSSelectionFilterType `xml:"filter"`
}

type DNSSelectionFilterType struct {
	Text string `xml:",chardata"`

	ID int `xml:"id"`
}

type SiteTypeRequest struct {
	Text string `xml:",chardata"`

	Get SiteGetRequest `xml:"get"`
}

type SiteGetRequest struct {
	Text string `xml:",chardata"`

	Filter  *SiteFilterType `xml:"filter,omitempty"`
	Dataset SiteDatasetType `xml:"dataset,omitempty"`
}

type SiteFilterType struct {
	Text string `xml:",chardata"`

	Name string `xml:"name"`
}

type SiteDatasetType struct {
	Text string `xml:",chardata"`

	GenInfo *SiteGenInfoType `xml:"gen_info,omitempty"`
}

type SiteGenInfoType struct {
	Text string `xml:",chardata"`

	CrDate       string `xml:"cr_date,omitempty"`
	Name         string `xml:"name,omitempty"`
	ASCIIName    string `xml:"ascii-name,omitempty"`
	Status       string `xml:"status,omitempty"`
	RealSize     string `xml:"real_size,omitempty"`
	DNSIPAddress string `xml:"dns_ip_address,omitempty"`
	HType        string `xml:"htype,omitempty"`
	GUID         string `xml:"guid,omitempty"`
	WebspaceGUID string `xml:"webspace-guid,omitempty"`
	SbSiteUUID   string `xml:"sb-site-uuid,omitempty"`
	WebspaceID   string `xml:"webspace-id,omitempty"`
	Description  string `xml:"description,omitempty"`
}

// Response.

type ResponsePacketType struct {
	XMLName xml.Name `xml:"packet"`
	Text    string   `xml:",chardata"`

	DNS    DNSResponseType  `xml:"dns,omitempty"`
	Site   SiteResponseType `xml:"site,omitempty"`
	System *System          `xml:"system,omitempty"`
}

type System struct {
	Text string `xml:",chardata"`

	Status  string `xml:"status"`
	ErrCode string `xml:"errcode"`
	ErrText string `xml:"errtext"`
}

func (s System) Error() string {
	return fmt.Sprintf("%s: %s - %s", s.Status, s.ErrCode, s.ErrText)
}

type DNSResponseType struct {
	Text string `xml:",chardata"`

	AddRec []AddRecResponse `xml:"add_rec,omitempty"`
	DelRec []DelRecResponse `xml:"del_rec,omitempty"`
}

type AddRecResponse struct {
	Text string `xml:",chardata"`

	Result RecResult `xml:"result,omitempty"`
}

type DelRecResponse struct {
	Text string `xml:",chardata"`

	Result RecResult `xml:"result"`
}

type RecResult struct {
	Text string `xml:",chardata"`

	ID int `xml:"id"`

	Status  string `xml:"status"`
	ErrCode string `xml:"errcode"`
	ErrText string `xml:"errtext"`
}

func (r RecResult) Error() string {
	return fmt.Sprintf("%s: %s - %s", r.Status, r.ErrCode, r.ErrText)
}

type SiteResponseType struct {
	Text string `xml:",chardata"`

	Get SiteGetResponse `xml:"get"`
}

type SiteGetResponse struct {
	Text string `xml:",chardata"`

	Result *SiteResult `xml:"result,omitempty"`
}

type SiteResult struct {
	Text string `xml:",chardata"`

	ID       int    `xml:"id"`
	FilterID string `xml:"filter-id"`

	Status  string `xml:"status"`
	ErrCode string `xml:"errcode"`
	ErrText string `xml:"errtext"`

	Data *SiteResultData `xml:"data"`
}

func (s SiteResult) Error() string {
	return fmt.Sprintf("%s: %s - %s", s.Status, s.ErrCode, s.ErrText)
}

type SiteResultData struct {
	Text string `xml:",chardata"`

	GenInfo *SiteGenInfoType `xml:"gen_info"`
}
