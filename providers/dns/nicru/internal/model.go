package internal

import (
	"encoding/xml"
	"fmt"
)

type Request struct {
	XMLName xml.Name `xml:"request" json:"xml_name,omitempty"`
	Text    string   `xml:",chardata" json:"text,omitempty"`
	RrList  *RrList  `xml:"rr-list" json:"rr_list,omitempty"`
}

type RrList struct {
	Text string `xml:",chardata" json:"text,omitempty"`
	Rr   []*RR  `xml:"rr" json:"rr,omitempty"`
}

type RR struct {
	Text    string   `xml:",chardata" json:"text,omitempty"`
	ID      string   `xml:"id,attr,omitempty" json:"id,omitempty"`
	Name    string   `xml:"name" json:"name,omitempty"`
	IdnName string   `xml:"idn-name,omitempty" json:"idn_name,omitempty"`
	Ttl     string   `xml:"ttl" json:"ttl,omitempty"`
	Type    string   `xml:"type" json:"type,omitempty"`
	Soa     *Soa     `xml:"soa" xml:"soa,omitempty"`
	A       *Address `xml:"a" json:"a,omitempty"`
	AAAA    *Address `xml:"aaaa" json:"aaaa,omitempty"`
	Cname   *Cname   `xml:"cname" json:"cname,omitempty"`
	Ns      *Ns      `xml:"ns" json:"ns,omitempty"`
	Mx      *Mx      `xml:"mx" json:"mx,omitempty"`
	Srv     *Srv     `xml:"srv" json:"srv,omitempty"`
	Ptr     *Ptr     `xml:"ptr" json:"ptr,omitempty"`
	Txt     *Txt     `xml:"txt" json:"txt,omitempty"`
	Dname   *Dname   `xml:"dname" json:"dname,omitempty"`
	Hinfo   *Hinfo   `xml:"hinfo" json:"hinfo,omitempty"`
	Naptr   *Naptr   `xml:"naptr" json:"naptr,omitempty"`
	Rp      *Rp      `xml:"rp" json:"rp,omitempty"`
}

type Address string

func (address *Address) String() string {
	return string(*address)
}

type Service struct {
	Text         string `xml:",chardata" json:"text,omitempty"`
	Admin        string `xml:"admin,attr" json:"admin,omitempty"`
	DomainsLimit string `xml:"domains-limit,attr" json:"domains_limit,omitempty"`
	DomainsNum   string `xml:"domains-num,attr" json:"domains_num,omitempty"`
	Enable       string `xml:"enable,attr" json:"enable,omitempty"`
	HasPrimary   string `xml:"has-primary,attr" json:"has_primary,omitempty"`
	Name         string `xml:"name,attr" json:"name,omitempty"`
	Payer        string `xml:"payer,attr" json:"payer,omitempty"`
	Tariff       string `xml:"tariff,attr" json:"tariff,omitempty"`
	RrLimit      string `xml:"rr-limit,attr" json:"rr_limit,omitempty"`
	RrNum        string `xml:"rr-num,attr" json:"rr_num,omitempty"`
}

type Soa struct {
	Text    string `xml:",chardata" json:"text,omitempty"`
	Mname   *Mname `xml:"mname" json:"mname,omitempty"`
	Rname   *Rname `xml:"rname" json:"rname,omitempty"`
	Serial  string `xml:"serial" json:"serial,omitempty"`
	Refresh string `xml:"refresh" json:"refresh,omitempty"`
	Retry   string `xml:"retry" json:"retry,omitempty"`
	Expire  string `xml:"expire" json:"expire,omitempty"`
	Minimum string `xml:"minimum" json:"minimum,omitempty"`
}

type Mname struct {
	Text    string `xml:",chardata" json:"text,omitempty"`
	Name    string `xml:"name" json:"name,omitempty"`
	IdnName string `xml:"idn-name,omitempty" json:"idn_name,omitempty"`
}

type Rname struct {
	Text    string `xml:",chardata" json:"text,omitempty"`
	Name    string `xml:"name" json:"name,omitempty"`
	IdnName string `xml:"idn-name,omitempty" json:"idn_name,omitempty"`
}

type Ns struct {
	Text    string `xml:",chardata" json:"text,omitempty"`
	Name    string `xml:"name" json:"name,omitempty"`
	IdnName string `xml:"idn-name,omitempty" json:"idn_name,omitempty"`
}

type Mx struct {
	Text       string    `xml:",chardata" json:"text,omitempty"`
	Preference string    `xml:"preference" json:"preference,omitempty"`
	Exchange   *Exchange `xml:"exchange" json:"exchange,omitempty"`
}

type Exchange struct {
	Text string `xml:",chardata" json:"text,omitempty"`
	Name string `xml:"name" json:"name,omitempty"`
}

type Srv struct {
	Text     string  `xml:",chardata" json:"text,omitempty"`
	Priority string  `xml:"priority" json:"priority,omitempty"`
	Weight   string  `xml:"weight" json:"weight,omitempty"`
	Port     string  `xml:"port" json:"port,omitempty"`
	Target   *Target `xml:"target" json:"target,omitempty"`
}

type Target struct {
	Text string `xml:",chardata" json:"text,omitempty"`
	Name string `xml:"name" json:"name,omitempty"`
}

type Ptr struct {
	Text string `xml:",chardata" json:"text,omitempty"`
	Name string `xml:"name" json:"name,omitempty"`
}

type Hinfo struct {
	Text     string `xml:",chardata" json:"text,omitempty"`
	Hardware string `xml:"hardware" json:"hardware,omitempty"`
	Os       string `xml:"os" json:"os,omitempty"`
}

type Naptr struct {
	Text        string       `xml:",chardata" json:"text,omitempty"`
	Order       string       `xml:"order" json:"order,omitempty"`
	Preference  string       `xml:"preference" json:"preference,omitempty"`
	Flags       string       `xml:"flags" json:"flags,omitempty"`
	Service     string       `xml:"service" json:"service,omitempty"`
	Regexp      string       `xml:"regexp" json:"regexp,omitempty"`
	Replacement *Replacement `xml:"replacement" json:"replacement,omitempty"`
}

type Replacement struct {
	Text string `xml:",chardata" json:"text,omitempty"`
	Name string `xml:"name" json:"name,omitempty"`
}

type Rp struct {
	Text      string     `xml:",chardata" json:"text,omitempty"`
	MboxDname *MboxDname `xml:"mbox-dname" json:"mbox_dname,omitempty"`
	TxtDname  *TxtDname  `xml:"txt-dname" json:"txt_dname,omitempty"`
}

type MboxDname struct {
	Text string `xml:",chardata" json:"text,omitempty"`
	Name string `xml:"name" json:"name,omitempty"`
}

type TxtDname struct {
	Text string `xml:",chardata" json:"text,omitempty"`
	Name string `xml:"name" json:"name,omitempty"`
}

type Cname struct {
	Text    string `xml:",chardata" json:"text,omitempty"`
	Name    string `xml:"name" json:"name,omitempty"`
	IdnName string `xml:"idn-name,omitempty" json:"idn_name,omitempty"`
}

type Dname struct {
	Text string `xml:",chardata" json:"text,omitempty"`
	Name string `xml:"name" json:"name,omitempty"`
}

type Txt struct {
	Text   string `xml:",chardata" json:"text,omitempty"`
	String string `xml:"string" json:"string,omitempty"`
}

type Zone struct {
	Text       string `xml:",chardata" json:"text,omitempty"`
	Admin      string `xml:"admin,attr" json:"admin,omitempty"`
	Enable     string `xml:"enable,attr" json:"enable,omitempty"`
	HasChanges string `xml:"has-changes,attr" json:"has_changes,omitempty"`
	HasPrimary string `xml:"has-primary,attr" json:"has_primary,omitempty"`
	ID         string `xml:"id,attr" json:"id,omitempty"`
	IdnName    string `xml:"idn-name,attr" json:"idn_name,omitempty"`
	Name       string `xml:"name,attr" json:"name,omitempty"`
	Payer      string `xml:"payer,attr" json:"payer,omitempty"`
	Service    string `xml:"service,attr" json:"service,omitempty"`
	Rr         []*RR  `xml:"rr" json:"rr,omitempty"`
}

type Revision struct {
	Text   string `xml:",chardata" json:"text,omitempty"`
	Date   string `xml:"date,attr" json:"date,omitempty"`
	Ip     string `xml:"ip,attr" json:"ip,omitempty"`
	Number string `xml:"number,attr" json:"number,omitempty"`
}

type Error struct {
	Text string `xml:",chardata" json:"text,omitempty"`
	Code string `xml:"code,attr" json:"code,omitempty"`
}

func describeError(e Error) string {
	return fmt.Sprintf(`%s (code %s)`, e.Text, e.Code)
}

type Response struct {
	XMLName xml.Name `xml:"response" json:"xml_name,omitempty"`
	Text    string   `xml:",chardata" json:"text,omitempty"`
	Status  string   `xml:"status" json:"status,omitempty"`
	Errors  struct {
		Text  string `xml:",chardata" json:"text,omitempty"`
		Error Error  `xml:"error" json:"error,omitempty"`
	} `xml:"errors" json:"errors,omitempty"`
	Data *Data `xml:"data" json:"data,omitempty"`
}

type Data struct {
	Text     string      `xml:",chardata" json:"text,omitempty"`
	Service  []*Service  `xml:"service" json:"service,omitempty"`
	Zone     []*Zone     `xml:"zone" json:"zone,omitempty"`
	Address  []*Address  `xml:"address" json:"address,omitempty"`
	Revision []*Revision `xml:"revision" json:"revision,omitempty"`
}
