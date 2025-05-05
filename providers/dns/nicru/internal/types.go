package internal

import (
	"encoding/xml"
	"fmt"
)

type Request struct {
	XMLName xml.Name `xml:"request"`
	Text    string   `xml:",chardata"`
	RRList  *RRList  `xml:"rr-list"`
}

type RRList struct {
	Text string `xml:",chardata"`
	RR   []RR   `xml:"rr"`
}

type RR struct {
	Text    string `xml:",chardata"`
	ID      string `xml:"id,attr,omitempty"`
	Name    string `xml:"name"`
	IDNName string `xml:"idn-name"`
	TTL     string `xml:"ttl"`
	Type    string `xml:"type"`
	SOA     *SOA   `xml:"soa,omitempty"`
	A       string `xml:"a,omitempty"`
	AAAA    string `xml:"aaaa,omitempty"`
	CName   *CName `xml:"cname,omitempty"`
	NS      *NS    `xml:"ns,omitempty"`
	MX      *MX    `xml:"mx,omitempty"`
	SRV     *SRV   `xml:"srv,omitempty"`
	PTR     *PTR   `xml:"ptr,omitempty"`
	TXT     *TXT   `xml:"txt,omitempty"`
	DName   *DName `xml:"dname,omitempty"`
	HInfo   *HInfo `xml:"hinfo,omitempty"`
	NAPTR   *NAPTR `xml:"naptr,omitempty"`
	RP      *RP    `xml:"rp,omitempty"`
}

type SOA struct {
	Text    string `xml:",chardata"`
	MName   *MName `xml:"mname"`
	RName   *RName `xml:"rname"`
	Serial  string `xml:"serial"`
	Refresh string `xml:"refresh"`
	Retry   string `xml:"retry"`
	Expire  string `xml:"expire"`
	Minimum string `xml:"minimum"`
}

type MName struct {
	Text    string `xml:",chardata"`
	Name    string `xml:"name"`
	IDNName string `xml:"idn-name,omitempty"`
}

type RName struct {
	Text    string `xml:",chardata"`
	Name    string `xml:"name"`
	IDNName string `xml:"idn-name,omitempty"`
}

type NS struct {
	Text    string `xml:",chardata"`
	Name    string `xml:"name"`
	IDNName string `xml:"idn-name,omitempty"`
}

type MX struct {
	Text       string    `xml:",chardata"`
	Preference string    `xml:"preference"`
	Exchange   *Exchange `xml:"exchange"`
}

type Exchange struct {
	Name string `xml:"name"`
}

type SRV struct {
	Text     string  `xml:",chardata"`
	Priority string  `xml:"priority"`
	Weight   string  `xml:"weight"`
	Port     string  `xml:"port"`
	Target   *Target `xml:"target"`
}

type Target struct {
	Text string `xml:",chardata"`
	Name string `xml:"name"`
}

type PTR struct {
	Text string `xml:",chardata"`
	Name string `xml:"name"`
}

type HInfo struct {
	Text     string `xml:",chardata"`
	Hardware string `xml:"hardware"`
	OS       string `xml:"os"`
}

type NAPTR struct {
	Text        string       `xml:",chardata"`
	Order       string       `xml:"order"`
	Preference  string       `xml:"preference"`
	Flags       string       `xml:"flags"`
	Service     string       `xml:"service"`
	Regexp      string       `xml:"regexp"`
	Replacement *Replacement `xml:"replacement"`
}

type Replacement struct {
	Text string `xml:",chardata"`
	Name string `xml:"name"`
}

type RP struct {
	Text      string     `xml:",chardata"`
	MboxDName *MboxDName `xml:"mbox-dname"`
	TxtDName  *TxtDName  `xml:"txt-dname"`
}

type MboxDName struct {
	Text string `xml:",chardata"`
	Name string `xml:"name"`
}

type TxtDName struct {
	Text string `xml:",chardata"`
	Name string `xml:"name"`
}

type CName struct {
	Text    string `xml:",chardata"`
	Name    string `xml:"name"`
	IDNName string `xml:"idn-name,omitempty"`
}

type DName struct {
	Text string `xml:",chardata"`
	Name string `xml:"name"`
}

type TXT struct {
	Text   string `xml:",chardata"`
	String string `xml:"string"`
}

type Response struct {
	XMLName xml.Name `xml:"response"`
	Text    string   `xml:",chardata"`
	Status  string   `xml:"status"`
	Data    *Data    `xml:"data"`
	Errors  Errors   `xml:"errors"`
}

type Data struct {
	Text     string     `xml:",chardata"`
	Service  []Service  `xml:"service"`
	Zone     []Zone     `xml:"zone"`
	Address  []string   `xml:"address"`
	Revision []Revision `xml:"revision"`
}

type Errors struct {
	Text  string `xml:",chardata"`
	Error Error  `xml:"error"`
}

type Error struct {
	Text string `xml:",chardata"`
	Code string `xml:"code,attr"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%s (code %s)", e.Text, e.Code)
}

type Service struct {
	Text         string `xml:",chardata"`
	Admin        string `xml:"admin,attr"`
	DomainsLimit string `xml:"domains-limit,attr"`
	DomainsNum   string `xml:"domains-num,attr"`
	Enable       string `xml:"enable,attr"`
	HasPrimary   string `xml:"has-primary,attr"`
	Name         string `xml:"name,attr"`
	Payer        string `xml:"payer,attr"`
	Tariff       string `xml:"tariff,attr"`
	RRLimit      string `xml:"rr-limit,attr"`
	RRNum        string `xml:"rr-num,attr"`
}

type Zone struct {
	Text       string `xml:",chardata"`
	Admin      string `xml:"admin,attr"`
	Enable     string `xml:"enable,attr"`
	HasChanges string `xml:"has-changes,attr"`
	HasPrimary string `xml:"has-primary,attr"`
	ID         string `xml:"id,attr"`
	IDNName    string `xml:"idn-name,attr"`
	Name       string `xml:"name,attr"`
	Payer      string `xml:"payer,attr"`
	Service    string `xml:"service,attr"`
	RR         []RR   `xml:"rr"`
}

type Revision struct {
	Text   string `xml:",chardata"`
	Date   string `xml:"date,attr"`
	IP     string `xml:"ip,attr"`
	Number string `xml:"number,attr"`
}
