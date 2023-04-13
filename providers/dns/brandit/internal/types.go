package internal

type StatusDomain struct {
	Response StatusResponse `json:"response,omitempty"`
	Code     int            `json:"code"`
	Status   string         `json:"status"`
	Error    string         `json:"error"`
}

type StatusResponse struct {
	Renewalmode                []string `json:"renewalmode"`
	Status                     []string `json:"status"`
	Transferlock               []int    `json:"transferlock"`
	Registrar                  []string `json:"registrar"`
	Paiduntildate              []string `json:"paiduntildate"`
	Nameserver                 []string `json:"nameserver"`
	Registrationexpirationdate []string `json:"registrationexpirationdate"`
	Domain                     []string `json:"domain"`
	Renewaldate                []string `json:"renewaldate"`
	Updateddate                []string `json:"updateddate"`
	Billingcontact             []string `json:"billingcontact"`
	XDomainRoid                []string `json:"x-domain-roid"`
	Admincontact               []string `json:"admincontact"`
	Techcontact                []string `json:"techcontact"`
	Domainidn                  []string `json:"domainidn"`
	Createddate                []string `json:"createddate"`
	Registrartransferdate      []string `json:"registrartransferdate"`
	Zone                       []string `json:"zone"`
	Auth                       []string `json:"auth"`
	Updatedby                  []string `json:"updatedby"`
	Roid                       []string `json:"roid"`
	Ownercontact               []string `json:"ownercontact"`
	Createdby                  []string `json:"createdby"`
	Transfermode               []string `json:"transfermode"`
}

type ListRecords struct {
	Response ListRecordsResponse `json:"response,omitempty"`
	Code     int                 `json:"code"`
	Status   string              `json:"status"`
	Error    string              `json:"error"`
}

type ListRecordsResponse struct {
	Limit  []int    `json:"limit,omitempty"`
	Column []string `json:"column,omitempty"`
	Count  []int    `json:"count,omitempty"`
	First  []int    `json:"first,omitempty"`
	Total  []int    `json:"total,omitempty"`
	Rr     []string `json:"rr,omitempty"`
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

type DeleteRecord struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Error  string `json:"error"`
}
