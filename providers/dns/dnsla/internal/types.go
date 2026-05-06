package internal

type Pager struct {
	PageIndex int `url:"pageIndex"`
	PageSize  int `url:"pageSize"`
}

type Record struct {
	DomainID string `json:"domainId"`
	Type     int    `json:"type"`
	Host     string `json:"host"`
	Data     string `json:"data"`
	TTL      int    `json:"ttl"`
}

type BaseResponse[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type RecordID struct {
	ID string `json:"id"`
}

type Domains struct {
	Total   int      `json:"total"`
	Results []Domain `json:"results"`
}

type Domain struct {
	ID            string `json:"id"`
	CreatedAt     int    `json:"createdAt"`
	UpdatedAt     int    `json:"updatedAt"`
	UserID        string `json:"userID"`
	UserAccount   string `json:"userAccount"`
	AssetID       string `json:"assetId"`
	GroupID       string `json:"groupId"`
	GroupName     string `json:"groupName"`
	Domain        string `json:"domain"`
	DisplayDomain string `json:"displayDomain"`
	State         int    `json:"state"`
	NsState       int    `json:"nsState"`
	NsCheckedAt   int    `json:"nsCheckedAt"`
	ProductCode   string `json:"productCode"`
	ProductName   string `json:"productName"`
	ExpiredAt     int64  `json:"expiredAt"`
	QuoteDomainID string `json:"quoteDomainId"`
	QuoteDomain   string `json:"quoteDomain"`
	Suffix        string `json:"suffix"`
	DisplaySuffix string `json:"displaySuffix"`
}
