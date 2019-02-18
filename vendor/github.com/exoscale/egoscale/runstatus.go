package egoscale

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RunstatusValidationErrorResponse represents an error in the API
type RunstatusValidationErrorResponse map[string][]string

// RunstatusErrorResponse represents the default errors
type RunstatusErrorResponse struct {
	Detail string `json:"detail"`
}

// runstatusPagesURL is the only URL that cannot be guessed
const runstatusPagesURL = "/pages"

// RunstatusPage runstatus page
type RunstatusPage struct {
	Created          *time.Time             `json:"created,omitempty"`
	DarkTheme        bool                   `json:"dark_theme,omitempty"`
	Domain           string                 `json:"domain,omitempty"`
	GradientEnd      string                 `json:"gradient_end,omitempty"`
	GradientStart    string                 `json:"gradient_start,omitempty"`
	HeaderBackground string                 `json:"header_background,omitempty"`
	ID               int                    `json:"id,omitempty"`
	Incidents        []RunstatusIncident    `json:"incidents,omitempty"`
	IncidentsURL     string                 `json:"incidents_url,omitempty"`
	Logo             string                 `json:"logo,omitempty"`
	Maintenances     []RunstatusMaintenance `json:"maintenances,omitempty"`
	MaintenancesURL  string                 `json:"maintenances_url,omitempty"`
	Name             string                 `json:"name"` //fake field (used to post a new runstatus page)
	OkText           string                 `json:"ok_text,omitempty"`
	Plan             string                 `json:"plan,omitempty"`
	PublicURL        string                 `json:"public_url,omitempty"`
	Services         []RunstatusService     `json:"services,omitempty"`
	ServicesURL      string                 `json:"services_url,omitempty"`
	State            string                 `json:"state,omitempty"`
	Subdomain        string                 `json:"subdomain"`
	SupportEmail     string                 `json:"support_email,omitempty"`
	TimeZone         string                 `json:"time_zone,omitempty"`
	Title            string                 `json:"title,omitempty"`
	TitleColor       string                 `json:"title_color,omitempty"`
	TwitterUsername  string                 `json:"twitter_username,omitempty"`
	URL              string                 `json:"url,omitempty"`
}

// Match returns true if the other page has got similarities with itself
func (page RunstatusPage) Match(other RunstatusPage) bool {
	if other.Subdomain != "" && page.Subdomain == other.Subdomain {
		return true
	}

	if other.ID > 0 && page.ID == other.ID {
		return true
	}

	return false
}

//RunstatusPageList runstatus page list
type RunstatusPageList struct {
	Count    int             `json:"count"`
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
	Results  []RunstatusPage `json:"results"`
}

// CreateRunstatusPage create runstatus page
func (client *Client) CreateRunstatusPage(ctx context.Context, page RunstatusPage) (*RunstatusPage, error) {
	resp, err := client.runstatusRequest(ctx, client.Endpoint+runstatusPagesURL, page, "POST")
	if err != nil {
		return nil, err
	}

	var p *RunstatusPage
	if err := json.Unmarshal(resp, &p); err != nil {
		return nil, err
	}

	return p, nil
}

// DeleteRunstatusPage delete runstatus page
func (client *Client) DeleteRunstatusPage(ctx context.Context, page RunstatusPage) error {
	if page.URL == "" {
		return fmt.Errorf("empty URL for %#v", page)
	}
	_, err := client.runstatusRequest(ctx, page.URL, nil, "DELETE")
	return err
}

// GetRunstatusPage fetches the runstatus page
func (client *Client) GetRunstatusPage(ctx context.Context, page RunstatusPage) (*RunstatusPage, error) {
	if page.URL != "" {
		return client.getRunstatusPage(ctx, page.URL)
	}

	ps, err := client.ListRunstatusPages(ctx)
	if err != nil {
		return nil, err
	}

	for i := range ps {
		if ps[i].Match(page) {
			return &ps[i], nil
		}
	}

	return nil, fmt.Errorf("%#v not found", page)
}

func (client *Client) getRunstatusPage(ctx context.Context, pageURL string) (*RunstatusPage, error) {
	resp, err := client.runstatusRequest(ctx, pageURL, nil, "GET")
	if err != nil {
		return nil, err
	}

	p := new(RunstatusPage)
	if err := json.Unmarshal(resp, p); err != nil {
		return nil, err
	}

	// NOTE: fix the missing IDs
	for i := range p.Maintenances {
		if err := p.Maintenances[i].FakeID(); err != nil {
			log.Printf("bad fake ID for %#v, %s", p.Maintenances[i], err)
		}
	}

	return p, nil
}

// ListRunstatusPages list all the runstatus pages
func (client *Client) ListRunstatusPages(ctx context.Context) ([]RunstatusPage, error) {
	resp, err := client.runstatusRequest(ctx, client.Endpoint+runstatusPagesURL, nil, "GET")
	if err != nil {
		return nil, err
	}

	var p *RunstatusPageList
	if err := json.Unmarshal(resp, &p); err != nil {
		return nil, err
	}

	// XXX: handle pagination
	return p.Results, nil
}

// Error formats the DNSerror into a string
func (req RunstatusErrorResponse) Error() string {
	return fmt.Sprintf("Runstatus error: %s", req.Detail)
}

// Error formats the DNSerror into a string
func (req RunstatusValidationErrorResponse) Error() string {
	if len(req) > 0 {
		errs := []string{}
		for name, ss := range req {
			if len(ss) > 0 {
				errs = append(errs, fmt.Sprintf("%s: %s", name, strings.Join(ss, ", ")))
			}
		}
		return fmt.Sprintf("Runstatus error: %s", strings.Join(errs, "; "))
	}
	return fmt.Sprintf("Runstatus error")
}

func (client *Client) runstatusRequest(ctx context.Context, uri string, structParam interface{}, method string) (json.RawMessage, error) {
	reqURL, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if reqURL.Scheme == "" {
		return nil, fmt.Errorf("only absolute URI are considered valid, got %q", uri)
	}

	var params string
	if structParam != nil {
		m, err := json.Marshal(structParam)
		if err != nil {
			return nil, err
		}
		params = string(m)
	}

	req, err := http.NewRequest(method, reqURL.String(), strings.NewReader(params))
	if err != nil {
		return nil, err
	}

	time := time.Now().Local().Format("2006-01-02T15:04:05-0700")

	payload := fmt.Sprintf("%s%s%s", req.URL.String(), time, params)

	mac := hmac.New(sha256.New, []byte(client.apiSecret))
	_, err = mac.Write([]byte(payload))
	if err != nil {
		return nil, err
	}
	signature := hex.EncodeToString(mac.Sum(nil))

	var hdr = make(http.Header)

	hdr.Add("Authorization", fmt.Sprintf("Exoscale-HMAC-SHA256 %s:%s", client.APIKey, signature))
	hdr.Add("Exoscale-Date", time)
	hdr.Add("User-Agent", fmt.Sprintf("exoscale/egoscale (%v)", Version))
	hdr.Add("Accept", "application/json")
	if params != "" {
		hdr.Add("Content-Type", "application/json")
	}
	req.Header = hdr

	req = req.WithContext(ctx)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // nolint: errcheck

	if resp.StatusCode == 204 {
		if method != "DELETE" {
			return nil, fmt.Errorf("only DELETE is expected to produce 204, was %q", method)
		}
		return nil, nil
	}

	contentType := resp.Header.Get("content-type")
	if !strings.Contains(contentType, "application/json") {
		return nil, fmt.Errorf(`response content-type expected to be "application/json", got %q`, contentType)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		rerr := new(RunstatusValidationErrorResponse)
		if err := json.Unmarshal(b, rerr); err == nil {
			return nil, rerr
		}
		rverr := new(RunstatusErrorResponse)
		if err := json.Unmarshal(b, rverr); err != nil {
			return nil, err
		}

		return nil, rverr
	}

	return b, nil
}
