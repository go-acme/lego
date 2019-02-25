package hostingde

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const defaultBaseURL = "https://secure.hosting.de/api/dns/v1/json"

// RecordsAddRequest represents a DNS record to add
type RecordsAddRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

// RecordsDeleteRequest represents a DNS record to remove
type RecordsDeleteRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	//ID      string `json:"id"`
}

// RequestError represents an error in an API response
type RequestError struct {
	Code          int      `json:"code"`
	ContextObject string   `json:"contextObject"`
	ContextPath   string   `json:"contextPath"`
	Details       []string `json:"details"`
	Text          string   `json:"text"`
	Value         string   `json:"value"`
}

// RequestMetadata represents the metadata in an API response
type RequestMetadata struct {
	ClientTransactionID string `json:"clientTransactionId"`
	ServerTransactionID string `json:"serverTransactionId"`
}

// Filter is used to filter FindRequests to the hosting.de API
type Filter struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

// Sort is used to sort FindRequests from the hosting.de API
type Sort struct {
	Field string `json:"zoneName"`
	Order string `json:"order"`
}

// ZoneUpdateResponse represents a response from hosting.de API
type ZoneUpdateResponse struct {
	Errors   []RequestError  `json:"errors"`
	Metadata RequestMetadata `json:"metadata"`
	Warnings []string        `json:"warnings"`
	Status   string          `json:"status"`
	Response struct {
		Records []struct {
			Content          string `json:"content"`
			Type             string `json:"type"`
			ID               string `json:"id"`
			Name             string `json:"name"`
			LastChangeDate   string `json:"lastChangeDate"`
			Priority         int    `json:"priority"`
			RecordTemplateID string `json:"recordTemplateId"`
			ZoneConfigID     string `json:"zoneConfigId"`
			TTL              int    `json:"ttl"`
		} `json:"records"`
		ZoneConfig ZoneConfig `json:"zoneConfig"`
	} `json:"response"`
}

// ZoneConfig represents the ZoneConfig  object used in hosting.de API requests&replys
type ZoneConfig struct {
	ID                    string   `json:"id"`
	AccountID             string   `json:"accountId"`
	Status                string   `json:"status"`
	Name                  string   `json:"name"`
	NameUnicode           string   `json:"nameUnicode"`
	MasterIP              string   `json:"masterIp"`
	Type                  string   `json:"type"`
	EMailAddress          string   `json:"emailAddress"`
	ZoneTransferWhitelist []string `json:"zoneTransferWhitelist"`
	LastChangeDate        string   `json:"lastChangeDate"`
	DNSServerGroupID      string   `json:"dnsServerGroupId"`
	DNSSecMode            string   `json:"dnsSecMode"`
	SOAValues             struct {
		Refresh     int `json:"refresh"`
		Retry       int `json:"retry"`
		Expire      int `json:"expire"`
		TTL         int `json:"ttl"`
		NegativeTTL int `json:"negativeTtl"`
	} `json:"soaValues,omitempty"`
	TemplateValues json.RawMessage `json:"templateValues,omitempty"`
}

// ZoneUpdateRequest represents a hosting.de API ZoneUpdate request
type ZoneUpdateRequest struct {
	AuthToken       string `json:"authToken"`
	ZoneConfig      `json:"zoneConfig"`
	RecordsToAdd    []RecordsAddRequest    `json:"recordsToAdd"`
	RecordsToDelete []RecordsDeleteRequest `json:"recordsToDelete"`
}

// ZoneConfigsFindRequest represents a hosting.de API ZonesFind request
type ZoneConfigsFindRequest struct {
	AuthToken string `json:"authToken"`
	Filter    Filter `json:"filter"`
	Limit     int    `json:"limit"`
	Page      int    `json:"page"`
	Sort      *Sort  `json:"sort,omitempty"`
}

// ZoneConfigsFindResponse represents the API response for ZoneConfigsFind
type ZoneConfigsFindResponse struct {
	Errors   []RequestError  `json:"errors"`
	Metadata RequestMetadata `json:"metadata"`
	Warnings []string        `json:"warnings"`
	Status   string          `json:"status"`
	Response struct {
		Data         []ZoneConfig `json:"data"`
		Limit        int          `json:"limit"`
		Page         int          `json:"page"`
		TotalEntries int          `json:"totalEntries"`
		TotalPages   int          `json:"totalPages"`
		Type         string       `json:"type"`
	} `json:"response"`
}

func (d *DNSProvider) zoneConfigsFind(findRequest ZoneConfigsFindRequest) (*ZoneConfigsFindResponse, error) {
	body, err := json.Marshal(findRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, defaultBaseURL+"/zoneConfigsFind", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(toUnreadableBodyMessage(req, content))
	}

	findResponse := &ZoneConfigsFindResponse{}
	err = json.Unmarshal(content, findResponse)
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, toUnreadableBodyMessage(req, content))
	}

	if len(findResponse.Response.Data) == 0 {
		return nil, fmt.Errorf("%v: %s", err, toUnreadableBodyMessage(req, content))
	}

	if findResponse.Status != "success" && findResponse.Status != "pending" {
		return findResponse, errors.New(toUnreadableBodyMessage(req, content))
	}

	return findResponse, nil
}

func (d *DNSProvider) getZone(findRequest ZoneConfigsFindRequest) (*ZoneConfig, error) {

	findResponse, err := d.zoneConfigsFind(findRequest)
	if err != nil {
		return nil, err
	}

	if findResponse.Response.Data[0].Status != "active" {
		// retry in case the zone was edited recently and is not yet active
		for {
			time.Sleep(3 * time.Second)
			res, err := d.getZone(findRequest)
			if err != nil {
				return nil, err
			}
			if res.Status == "active" {
				return res, nil
			}
		}
	}

	return &findResponse.Response.Data[0], nil
}

func (d *DNSProvider) updateZone(updateRequest ZoneUpdateRequest) (*ZoneUpdateResponse, error) {
	body, err := json.Marshal(updateRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, defaultBaseURL+"/zoneUpdate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error querying API: %v", err)
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(toUnreadableBodyMessage(req, content))
	}

	// Everything looks good; but we'll need the ID later to delete the record
	updateResponse := &ZoneUpdateResponse{}
	err = json.Unmarshal(content, updateResponse)
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, toUnreadableBodyMessage(req, content))
	}

	if updateResponse.Status != "success" && updateResponse.Status != "pending" {
		return updateResponse, errors.New(toUnreadableBodyMessage(req, content))
	}

	return updateResponse, nil
}

func toUnreadableBodyMessage(req *http.Request, rawBody []byte) string {
	return fmt.Sprintf("the request %s sent a response with a body which is an invalid format: %q", req.URL, string(rawBody))
}
