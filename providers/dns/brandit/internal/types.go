package internal

import "fmt"

type Response[T any] struct {
	Response T      `json:"response,omitempty"`
	Code     int    `json:"code"`
	Status   string `json:"status"`
	Error    string `json:"error"`
}

type StatusResponse struct {
	RenewalMode                []string `json:"renewalmode"`
	Status                     []string `json:"status"`
	TransferLock               []int    `json:"transferlock"`
	Registrar                  []string `json:"registrar"`
	PaidUntilDate              []string `json:"paiduntildate"`
	Nameserver                 []string `json:"nameserver"`
	RegistrationExpirationDate []string `json:"registrationexpirationdate"`
	Domain                     []string `json:"domain"`
	RenewalDate                []string `json:"renewaldate"`
	UpdatedDate                []string `json:"updateddate"`
	BillingContact             []string `json:"billingcontact"`
	XDomainRoID                []string `json:"x-domain-roid"`
	AdminContact               []string `json:"admincontact"`
	TechContact                []string `json:"techcontact"`
	DomainIDN                  []string `json:"domainidn"`
	CreatedDate                []string `json:"createddate"`
	RegistrarTransferDate      []string `json:"registrartransferdate"`
	Zone                       []string `json:"zone"`
	Auth                       []string `json:"auth"`
	UpdatedBy                  []string `json:"updatedby"`
	RoID                       []string `json:"roid"`
	OwnerContact               []string `json:"ownercontact"`
	CreatedBy                  []string `json:"createdby"`
	TransferMode               []string `json:"transfermode"`
}

type ListRecordsResponse struct {
	Limit  []int    `json:"limit,omitempty"`
	Column []string `json:"column,omitempty"`
	Count  []int    `json:"count,omitempty"`
	First  []int    `json:"first,omitempty"`
	Total  []int    `json:"total,omitempty"`
	RR     []string `json:"rr,omitempty"`
	Last   []int    `json:"last,omitempty"`
}

type APIError struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"error"`
}

func (a APIError) Error() string {
	return fmt.Sprintf("code: %d, status: %s, message: %s", a.Code, a.Status, a.Message)
}

type AddRecord struct {
	Response AddRecordResponse `json:"response"`
	Record   string            `json:"record"`
	Code     int               `json:"code"`
	Status   string            `json:"status"`
	Error    string            `json:"error"`
}

type AddRecordResponse struct {
	ZoneType []string `json:"zonetype"`
	Signed   []int    `json:"signed"`
}

type Record struct {
	ID      int    `json:"id,omitempty"`
	Type    string `json:"type,omitempty"`
	Name    string `json:"name,omitempty"` // subdomain name or @ if you don't want subdomain
	Content string `json:"content,omitempty"`
	TTL     int    `json:"ttl,omitempty"` // default 600
}
