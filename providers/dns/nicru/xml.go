package nicru

import "encoding/xml"

const (
	xmlHeader = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
)

type nicRequest struct {
	XMLName xml.Name    `xml:"request"`
	Records []nicRecord `xml:"rr-list>rr"`
}

// only supports TXT records
type nicRecord struct {
	Name      string `xml:"name"`
	TTL       int    `xml:"ttl"`
	Type      string `xml:"type"`
	TxtString string `xml:"txt>string,omitempty"`
}

type nicResponse struct {
	XMLName xml.Name        `xml:"response"`
	Zone    nicZoneResponse `xml:"data>zone"`
}

type nicZoneResponse struct {
	ID      string              `xml:"id,attr"`
	Records []nicRecordResponse `xml:"rr"`
}

type nicRecordResponse struct {
	nicRecord
	ID      string `xml:"id,attr"`
	IdnName string `xml:"idn-name"`
}
