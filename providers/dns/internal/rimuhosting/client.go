package rimuhosting

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
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
		apiKey:     apiKey,
		BaseURL:    DefaultZonomiBaseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// FindTXTRecords Finds TXT records.
// ex:
// - https://zonomi.com/app/dns/dyndns.jsp?action=QUERY&name=example.com&api_key=apikeyvaluehere
// - https://zonomi.com/app/dns/dyndns.jsp?action=QUERY&name=**.example.com&api_key=apikeyvaluehere
func (c *Client) FindTXTRecords(ctx context.Context, domain string) ([]Record, error) {
	action := ActionParameter{
		Action: QueryAction,
		Name:   domain,
		Type:   "TXT",
	}

	resp, err := c.DoActions(ctx, action)
	if err != nil {
		return nil, err
	}

	return resp.Actions.Action.Records, nil
}

// DoActions performs actions.
func (c *Client) DoActions(ctx context.Context, actions ...ActionParameter) (*DNSAPIResult, error) {
	if len(actions) == 0 {
		return nil, errors.New("no action")
	}

	resp := &DNSAPIResult{}

	if len(actions) == 1 {
		action := actionParameter{
			ActionParameter: actions[0],
			APIKey:          c.apiKey,
		}

		err := c.do(ctx, action, resp)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}

	multi := c.toMultiParameters(actions)

	err := c.do(ctx, multi, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) toMultiParameters(params []ActionParameter) multiActionParameter {
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

func (c *Client) do(ctx context.Context, params, result any) error {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return parseError(req, resp)
	}

	if result == nil {
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = xml.Unmarshal(raw, result)
	if err != nil {
		return fmt.Errorf("unmarshaling %T error: %w: %s", result, err, string(raw))
	}

	return nil
}

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	errAPI := APIError{}

	err := xml.Unmarshal(raw, &errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return errAPI
}

// NewAddRecordAction helper to create an action to add a TXT record.
func NewAddRecordAction(domain, content string, ttl int) ActionParameter {
	return ActionParameter{
		Action: SetAction,
		Name:   domain,
		Type:   "TXT",
		Value:  content,
		TTL:    ttl,
	}
}

// NewDeleteRecordAction helper to create an action to delete a TXT record.
func NewDeleteRecordAction(domain, content string) ActionParameter {
	return ActionParameter{
		Action: DeleteAction,
		Name:   domain,
		Type:   "TXT",
		Value:  content,
	}
}
