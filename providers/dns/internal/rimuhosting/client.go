package rimuhosting

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"

	querystring "github.com/google/go-querystring/query"
)

// Base URL for the RimuHosting DNS services.
const (
	DefaultZonomiBaseURL      = "https://zonomi.com/app/dns/dyndns.jsp"
	DefaultRimuHostingBaseURL = "https://rimuhosting.com/dns/dyndns.jsp"
)

// Action names.
const (
	SetAction    = "SET"
	QueryAction  = "QUERY"
	DeleteAction = "DELETE"
)

// Client the RimuHosting/Zonomi client.
type Client struct {
	apiKey string

	HTTPClient *http.Client
	BaseURL    string
}

// NewClient Creates a RimuHosting/Zonomi client.
func NewClient(apiKey string) *Client {
	return &Client{
		HTTPClient: http.DefaultClient,
		BaseURL:    DefaultZonomiBaseURL,
		apiKey:     apiKey,
	}
}

// FindTXTRecords Finds TXT records.
// ex:
// - https://zonomi.com/app/dns/dyndns.jsp?action=QUERY&name=example.com&api_key=apikeyvaluehere
// - https://zonomi.com/app/dns/dyndns.jsp?action=QUERY&name=**.example.com&api_key=apikeyvaluehere
func (c Client) FindTXTRecords(domain string) ([]Record, error) {
	action := ActionParameter{
		Action: QueryAction,
		Name:   domain,
		Type:   "TXT",
	}

	resp, err := c.DoActions(action)
	if err != nil {
		return nil, err
	}

	return resp.Actions.Action.Records, nil
}

// DoActions performs actions.
func (c Client) DoActions(actions ...ActionParameter) (*DNSAPIResult, error) {
	if len(actions) == 0 {
		return nil, errors.New("no action")
	}

	resp := &DNSAPIResult{}

	if len(actions) == 1 {
		action := actionParameter{
			ActionParameter: actions[0],
			APIKey:          c.apiKey,
		}

		err := c.do(action, resp)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	multi := c.toMultiParameters(actions)
	err := c.do(multi, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c Client) toMultiParameters(params []ActionParameter) multiActionParameter {
	multi := multiActionParameter{
		APIKey: c.apiKey,
	}

	for _, parameters := range params {
		multi.Action = append(multi.Action, parameters.Action)
		multi.Name = append(multi.Name, parameters.Name)
		multi.Type = append(multi.Type, parameters.Type)
		multi.Value = append(multi.Value, parameters.Value)
		multi.TTL = append(multi.TTL, parameters.TTL)
	}

	return multi
}

func (c Client) do(params, data interface{}) error {
	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}

	v, err := querystring.Values(params)
	if err != nil {
		return err
	}

	exp := regexp.MustCompile(`(%5B)(%5D)(\d+)=`)

	baseURL.RawQuery = exp.ReplaceAllString(v.Encode(), "${1}${3}${2}=")

	req, err := http.NewRequest(http.MethodGet, baseURL.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode/100 != 2 {
		r := APIError{}
		err = xml.Unmarshal(all, &r)
		if err != nil {
			return err
		}
		return r
	}

	if data != nil {
		err := xml.Unmarshal(all, data)
		if err != nil {
			return err
		}
	}

	return nil
}

// AddRecord helper to create an action to add a TXT record.
func AddRecord(domain, content string, ttl int) ActionParameter {
	return ActionParameter{
		Action: SetAction,
		Name:   domain,
		Type:   "TXT",
		Value:  content,
		TTL:    ttl,
	}
}

// DeleteRecord helper to create an action to delete a TXT record.
func DeleteRecord(domain, content string) ActionParameter {
	return ActionParameter{
		Action: DeleteAction,
		Name:   domain,
		Type:   "TXT",
		Value:  content,
	}
}
