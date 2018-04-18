// Package nifcloud implements a DNS provider for solving the DNS-01 challenge
// using NIFCLOUD DNS.
package nifcloud

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultEndpoint = "https://dns.api.cloud.nifty.com"
	apiVersion      = "2012-12-12N2013-12-16"
	xmlns           = "https://route53.amazonaws.com/doc/2012-12-12/"
)

// ChangeResourceRecordSetsRequest is a complex type that contains change information for the resource record set.
type ChangeResourceRecordSetsRequest struct {
	Xmlns       string      `xml:"xmlns,attr"`
	ChangeBatch ChangeBatch `xml:"ChangeBatch"`
}

// ChangeResourceRecordSetsResponse is a complex type containing the response for the request.
type ChangeResourceRecordSetsResponse struct {
	ChangeInfo ChangeInfo `xml:"ChangeInfo"`
}

// GetChangeResponse is a complex type that contains the ChangeInfo element.
type GetChangeResponse struct {
	ChangeInfo ChangeInfo `xml:"ChangeInfo"`
}

// ErrorResponse is the information for any errors.
type ErrorResponse struct {
	Error struct {
		Type    string `xml:"Type"`
		Message string `xml:"Message"`
		Code    string `xml:"Code"`
	} `xml:"Error"`
	RequestID string `xml:"RequestId"`
}

// ChangeBatch is the information for a change request.
type ChangeBatch struct {
	Changes Changes `xml:"Changes"`
	Comment string  `xml:"Comment"`
}

// Changes is array of Change.
type Changes struct {
	Change []Change `xml:"Change"`
}

// Change is the information for each resource record set that you want to change.
type Change struct {
	Action            string            `xml:"Action"`
	ResourceRecordSet ResourceRecordSet `xml:"ResourceRecordSet"`
}

// ResourceRecordSet is the information about the resource record set to create or delete.
type ResourceRecordSet struct {
	Name            string          `xml:"Name"`
	Type            string          `xml:"Type"`
	TTL             int             `xml:"TTL"`
	ResourceRecords ResourceRecords `xml:"ResourceRecords"`
}

// ResourceRecords is array of ResourceRecord.
type ResourceRecords struct {
	ResourceRecord []ResourceRecord `xml:"ResourceRecord"`
}

// ResourceRecord is the information specific to the resource record.
type ResourceRecord struct {
	Value string `xml:"Value"`
}

// ChangeInfo is A complex type that describes change information about changes made to your hosted zone.
type ChangeInfo struct {
	ID          string `xml:"Id"`
	Status      string `xml:"Status"`
	SubmittedAt string `xml:"SubmittedAt"`
}

// Client client of NIFCLOUD DNS
type Client struct {
	accessKey string
	secretKey string
	endpoint  string
}

// ChangeResourceRecordSets Call ChangeResourceRecordSets API and return response.
func (c *Client) ChangeResourceRecordSets(hostedZoneID string, input ChangeResourceRecordSetsRequest) (*ChangeResourceRecordSetsResponse, error) {
	requestURL := fmt.Sprintf("%s/%s/hostedzone/%s/rrset", c.endpoint, apiVersion, hostedZoneID)

	body := &bytes.Buffer{}
	body.Write([]byte(xml.Header))
	err := xml.NewEncoder(body).Encode(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, requestURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "text/xml; charset=utf-8")

	err = c.sign(req)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errResp := &ErrorResponse{}
		err = xml.NewDecoder(res.Body).Decode(errResp)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf(errResp.Error.Message)
	}

	output := &ChangeResourceRecordSetsResponse{}
	err = xml.NewDecoder(res.Body).Decode(output)
	if err != nil {
		return nil, err
	}
	return output, err
}

// GetChange Call GetChange API and return response.
func (c *Client) GetChange(statusID string) (*GetChangeResponse, error) {
	requestURL := fmt.Sprintf("%s/%s/change/%s", c.endpoint, apiVersion, statusID)

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	err = c.sign(req)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errResp := &ErrorResponse{}
		err = xml.NewDecoder(res.Body).Decode(errResp)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf(errResp.Error.Message)
	}

	output := &GetChangeResponse{}
	err = xml.NewDecoder(res.Body).Decode(output)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Client) sign(req *http.Request) error {
	location, err := time.LoadLocation("GMT")
	if err != nil {
		return err
	}

	ts := time.Now().In(location).Format(time.RFC1123)

	if req.Header.Get("Date") == "" {
		req.Header.Set("Date", ts)
	}

	if req.URL.Path == "" {
		req.URL.Path += "/"
	}

	mac := hmac.New(sha1.New, []byte(c.secretKey))
	_, err = mac.Write([]byte(req.Header.Get("Date")))
	if err != nil {
		return err
	}

	hashed := mac.Sum(nil)
	signature := base64.StdEncoding.EncodeToString(hashed)

	req.Header.Set("X-Nifty-Authorization", "NIFTY3-HTTPS NiftyAccessKeyId="+c.accessKey+",Algorithm=HmacSHA1,Signature="+signature)
	return err
}
