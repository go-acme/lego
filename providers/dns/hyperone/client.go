package hyperone

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-acme/lego/v3/providers/dns/hyperone/internal"
)

type Client struct {
	ZoneFullURI string
	Signer      *internal.TokenSigner
}

type Recordset struct {
	RecordType string  `json:"type"`
	Name       string  `json:"name"`
	TTL        int     `json:"ttl,omitempty"`
	ID         string  `json:"id,omitempty"`
	Record     *Record `json:"record,omitempty"`
}

type Record struct {
	ID      string `json:"id,omitempty"`
	Content string `json:"content"`
	Enabled bool   `json:"enabled,omitempty"`
}

type Zone struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	DNSName string `json:"dnsName"`
	FQDN    string `json:"fqdn"`
	URI     string `json:"uri"`
}

// findRecordset looks for recordset with given recordType and name
// and returns it. In case if recordset is not found returns nil.
func (c *Client) findRecordset(recordType, name string) (*Recordset, error) {
	resourceURL := fmt.Sprintf("%s/recordset", c.ZoneFullURI)

	body, err := c.performRequest("GET", resourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to get recordsets from server:%+v", err)
	}

	var recordsets []Recordset
	err = json.Unmarshal(body, &recordsets)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse recordsets JSON response from server:%+v", err)
	}

	for _, v := range recordsets {
		if v.RecordType == recordType && v.Name == name {
			return &v, nil
		}
	}

	// when recordset is not present returns nil, but error is not thrown
	return nil, nil
}

// getRecords gets all records within specified recordset.
func (c *Client) getRecords(recordsetID string) ([]Record, error) {
	resourceURL := fmt.Sprintf("%s/recordset/%s/record", c.ZoneFullURI, recordsetID)

	body, err := c.performRequest("GET", resourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to get records from server:%+v", err)
	}

	var records []Record
	err = json.Unmarshal(body, &records)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse records JSON response from server:%+v", err)
	}

	return records, err
}

// createRecordsetWithRecord creates recordset and record with given value within one request.
func (c *Client) createRecordsetWithRecord(recordType, name, recordValue string, ttl int) (*Recordset, error) {
	resourceURL := fmt.Sprintf("%s/recordset", c.ZoneFullURI)

	record := &Record{Content: recordValue}
	recordsetInput := Recordset{RecordType: recordType, Name: name, TTL: ttl, Record: record}
	requestBody, err := json.Marshal(recordsetInput)
	if err != nil {
		return nil, fmt.Errorf("Error when marshaling recordset JSON:%+v", err)
	}

	body, err := c.performRequest("POST", resourceURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("Failed to create recordset:%+v", err)
	}

	var recordsetResponse Recordset
	err = json.Unmarshal(body, &recordsetResponse)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse recordset JSON response from server:%+v", err)
	}

	return &recordsetResponse, nil
}

func (c *Client) setRecord(recordsetID, recordContent string) (*Record, error) {
	resourceURL := fmt.Sprintf("%s/recordset/%s/record", c.ZoneFullURI, recordsetID)

	recordInput := Record{Content: recordContent}
	requestBody, err := json.Marshal(recordInput)
	if err != nil {
		return nil, fmt.Errorf("Error when marshaling record JSON:%+v", err)
	}

	body, err := c.performRequest("POST", resourceURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("Failed to set record:%+v", err)
	}

	var recordResponse Record
	err = json.Unmarshal(body, &recordResponse)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse recordset JSON response from server:%+v", err)
	}

	return &recordResponse, nil
}

func (c *Client) deleteRecord(recordsetID, recordID string) error {
	resourceURL := fmt.Sprintf("%s/recordset/%s/record/%s", c.ZoneFullURI, recordsetID, recordID)

	_, err := c.performRequest("DELETE", resourceURL, nil)
	if err != nil {
		return fmt.Errorf("Error when deleting record:%+v", err)
	}

	return nil
}

func (c *Client) deleteRecordset(recordsetID string) error {
	resourceURL := fmt.Sprintf("%s/recordset/%s", c.ZoneFullURI, recordsetID)

	_, err := c.performRequest("DELETE", resourceURL, nil)
	if err != nil {
		return fmt.Errorf("Error when deleting recordset:%+v", err)
	}

	return nil
}

// findZone looks for DNS Zone URI and returns nil if it does not exist.
func (c *Client) findZone(zone, projectID, locationID, apiEndpoint string) (uri string, err error) {
	resourceURL := fmt.Sprintf("%s/dns/%s/project/%s/zone", apiEndpoint, locationID, projectID)

	body, err := c.performRequest("GET", resourceURL, nil)
	if err != nil {
		return "", fmt.Errorf("Error when fetching aviable zones:%+v", err)
	}

	var zones []Zone
	err = json.Unmarshal(body, &zones)
	if err != nil {
		return "", fmt.Errorf("Failed to unmarshal zone data:%+v", err)
	}

	for _, v := range zones {
		if v.DNSName == zone {
			return v.URI, nil
		}
	}

	return "", errors.New("Can't find zone with given DNSName")
}

// performRequest creates new request and signs it with JWT and returns response body.
func (c *Client) performRequest(method string, url string, body io.Reader) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	jwt, err := c.Signer.GetJWT()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response from server:%+v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		errorMessage := identifyError(resp.StatusCode)
		stringifiedBody := string(responseBody)
		// add response body to error message if not empty
		if stringifiedBody != "" {
			errorMessage = fmt.Sprintf("%s: %s", errorMessage, stringifiedBody)
		}
		return nil, errors.New(errorMessage)
	}

	return responseBody, nil
}

func identifyError(code int) string {
	switch code {
	case 403:
		return "Forbidden- check if service account you are trying to use has permissions required for managing DNS"
	default:
		return fmt.Sprintf("%d- unknown error", code)
	}
}
