package internal

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"net/http"
	"strconv"
)

const (
	BaseURL                 = `https://api.nic.ru`
	TokenURL                = BaseURL + `/oauth/token`
	GetZonesUrlPattern      = BaseURL + `/dns-master/services/%s/zones`
	GetRecordsUrlPattern    = BaseURL + `/dns-master/services/%s/zones/%s/records`
	DeleteRecordsUrlPattern = BaseURL + `/dns-master/services/%s/zones/%s/records/%d`
	AddRecordsUrlPattern    = BaseURL + `/dns-master/services/%s/zones/%s/records`
	CommitUrlPattern        = BaseURL + `/dns-master/services/%s/zones/%s/commit`
	SuccessStatus           = `success`
	OAuth2Scope             = `.+:/dns-master/.+`
)

// Provider facilitates DNS record manipulation with NIC.ru.
type Provider struct {
	OAuth2ClientID string `json:"oauth2_client_id"`
	OAuth2SecretID string `json:"oauth2_secret_id"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	ServiceName    string `json:"service_name"`
}

type Client struct {
	client   *http.Client
	provider *Provider
	token    string
}

func NewClient(provider *Provider) (*Client, error) {
	client := Client{provider: provider}
	err := client.validateAuthOptions()
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (client *Client) GetOauth2Client() error {
	ctx := context.TODO()

	oauth2Config := oauth2.Config{
		ClientID:     client.provider.OAuth2ClientID,
		ClientSecret: client.provider.OAuth2SecretID,
		Endpoint: oauth2.Endpoint{
			TokenURL:  TokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{OAuth2Scope},
	}

	oauth2Token, err := oauth2Config.PasswordCredentialsToken(ctx, client.provider.Username, client.provider.Password)
	if err != nil {
		return fmt.Errorf("nicru: %s", err.Error())
	}

	client.client = oauth2Config.Client(ctx, oauth2Token)
	return nil
}

func (client *Client) Do(r *http.Request) (*http.Response, error) {
	if client.client == nil {
		err := client.GetOauth2Client()
		if err != nil {
			return nil, err
		}
	}
	return client.client.Do(r)
}

func (client *Client) GetZones() ([]*Zone, error) {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf(GetZonesUrlPattern, client.provider.ServiceName), nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(response.Body); err != nil {
		return nil, err
	}

	apiResponse := &Response{}
	if err := xml.NewDecoder(buf).Decode(&apiResponse); err != nil {
		return nil, err
	} else {
		var zones []*Zone
		for _, zone := range apiResponse.Data.Zone {
			zones = append(zones, zone)
		}
		return zones, nil
	}
}

func (client *Client) GetRecords(fqdn string) ([]*RR, error) {
	request, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(GetRecordsUrlPattern, client.provider.ServiceName, fqdn),
		nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(response.Body); err != nil {
		return nil, err
	}

	apiResponse := &Response{}
	if err := xml.NewDecoder(buf).Decode(&apiResponse); err != nil {
		return nil, err
	} else {
		var records []*RR
		for _, zone := range apiResponse.Data.Zone {
			records = append(records, zone.Rr...)
		}
		return records, nil
	}
}

func (client *Client) add(zoneName string, request *Request) (*Response, error) {

	buf := bytes.NewBuffer(nil)
	if err := xml.NewEncoder(buf).Encode(request); err != nil {
		return nil, err
	}

	url := fmt.Sprintf(AddRecordsUrlPattern, client.provider.ServiceName, zoneName)

	req, err := http.NewRequest(http.MethodPut, url, buf)
	if err != nil {
		return nil, err
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	buf = bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(response.Body); err != nil {
		return nil, err
	}

	apiResponse := &Response{}
	if err := xml.NewDecoder(buf).Decode(&apiResponse); err != nil {
		return nil, err
	}

	if apiResponse.Status != SuccessStatus {
		return nil, fmt.Errorf(describeError(apiResponse.Errors.Error))
	} else {
		return apiResponse, nil
	}
}

func (client *Client) deleteRecord(zoneName string, id int) (*Response, error) {
	url := fmt.Sprintf(DeleteRecordsUrlPattern, client.provider.ServiceName, zoneName, id)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	apiResponse := Response{}
	if err := xml.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}
	if apiResponse.Status != SuccessStatus {
		return nil, err
	} else {
		return &apiResponse, nil
	}
}

func (client *Client) GetTXTRecords(fqdn string) ([]*Txt, error) {
	records, err := client.GetRecords(fqdn)
	if err != nil {
		return nil, err
	}

	txt := make([]*Txt, 0)
	for _, record := range records {
		if record.Txt != nil {
			txt = append(txt, record.Txt)
		}
	}

	return txt, nil
}

func (client *Client) AddTxtRecord(zoneName string, name string, content string, ttl int) (*Response, error) {
	request := &Request{
		RrList: &RrList{
			Rr: []*RR{},
		},
	}
	request.RrList.Rr = append(request.RrList.Rr, &RR{
		Name: name,
		Ttl:  strconv.Itoa(ttl),
		Type: `TXT`,
		Txt: &Txt{
			String: content,
		},
	})

	return client.add(zoneName, request)
}

func (client *Client) DeleteRecord(zoneName string, id int) (*Response, error) {
	url := fmt.Sprintf(DeleteRecordsUrlPattern, client.provider.ServiceName, zoneName, id)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	apiResponse := Response{}
	if err := xml.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}
	if apiResponse.Status != SuccessStatus {
		return nil, err
	} else {
		return &apiResponse, nil
	}
}

func (client *Client) CommitZone(zoneName string) (*Response, error) {
	url := fmt.Sprintf(CommitUrlPattern, client.provider.ServiceName, zoneName)
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	apiResponse := Response{}
	if err := xml.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}
	if apiResponse.Status != SuccessStatus {
		return nil, err
	} else {
		return &apiResponse, nil
	}
}

func (client *Client) validateAuthOptions() error {

	msg := " is missing in credentials information"

	if client.provider.ServiceName == "" {
		return errors.New("service name" + msg)
	}

	if client.provider.Username == "" {
		return errors.New("username" + msg)
	}

	if client.provider.Password == "" {
		return errors.New("password" + msg)
	}

	if client.provider.OAuth2ClientID == "" {
		return errors.New("serviceId" + msg)
	}

	if client.provider.OAuth2SecretID == "" {
		return errors.New("secret" + msg)
	}

	return nil
}
