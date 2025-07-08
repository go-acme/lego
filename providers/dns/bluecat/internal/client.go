package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// Object types.
const (
	ConfigType = "Configuration"
	ViewType   = "View"
	ZoneType   = "Zone"
	TXTType    = "TXTRecord"
)

const authorizationHeader = "Authorization"

type Client struct {
	username string
	password string

	tokenExp *regexp.Regexp

	baseURL    *url.URL
	HTTPClient *http.Client
}

func NewClient(baseURL, username, password string) *Client {
	bu, _ := url.Parse(baseURL)

	return &Client{
		username:   username,
		password:   password,
		tokenExp:   regexp.MustCompile("BAMAuthToken: [^ ]+"),
		baseURL:    bu,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Deploy the DNS config for the specified entity to the authoritative servers.
// https://docs.bluecatnetworks.com/r/Address-Manager-Legacy-v1-API-Guide/POST/v1/quickDeploy/9.5.0
func (c *Client) Deploy(ctx context.Context, entityID uint) error {
	endpoint := c.createEndpoint("quickDeploy")

	q := endpoint.Query()
	q.Set("entityId", strconv.FormatUint(uint64(entityID), 10))
	endpoint.RawQuery = q.Encode()

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.doAuthenticated(ctx, req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	// The API doc says that 201 is expected but in the reality 200 is return.
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	return nil
}

// AddEntity A generic method for adding configurations, DNS zones, and DNS resource records.
// https://docs.bluecatnetworks.com/r/Address-Manager-Legacy-v1-API-Guide/POST/v1/addEntity/9.5.0
func (c *Client) AddEntity(ctx context.Context, parentID uint, entity Entity) (uint64, error) {
	endpoint := c.createEndpoint("addEntity")

	q := endpoint.Query()
	q.Set("parentId", strconv.FormatUint(uint64(parentID), 10))
	endpoint.RawQuery = q.Encode()

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, entity)
	if err != nil {
		return 0, err
	}

	resp, err := c.doAuthenticated(ctx, req)
	if err != nil {
		return 0, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 0, errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, _ := io.ReadAll(resp.Body)

	// addEntity responds only with body text containing the ID of the created record
	addTxtResp := string(raw)
	id, err := strconv.ParseUint(addTxtResp, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("addEntity request failed: %s", addTxtResp)
	}

	return id, nil
}

// GetEntityByName Returns objects from the database referenced by their database ID and with its properties fields populated.
// https://docs.bluecatnetworks.com/r/Address-Manager-Legacy-v1-API-Guide/GET/v1/getEntityById/9.5.0
func (c *Client) GetEntityByName(ctx context.Context, parentID uint, name, objType string) (*EntityResponse, error) {
	endpoint := c.createEndpoint("getEntityByName")

	q := endpoint.Query()
	q.Set("parentId", strconv.FormatUint(uint64(parentID), 10))
	q.Set("name", name)
	q.Set("type", objType)
	endpoint.RawQuery = q.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doAuthenticated(ctx, req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	var entity EntityResponse
	err = json.Unmarshal(raw, &entity)
	if err != nil {
		return nil, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return &entity, nil
}

// Delete Deletes an object using the generic delete method.
// https://docs.bluecatnetworks.com/r/Address-Manager-Legacy-v1-API-Guide/DELETE/v1/delete/9.5.0
func (c *Client) Delete(ctx context.Context, objectID uint) error {
	endpoint := c.createEndpoint("delete")

	q := endpoint.Query()
	q.Set("objectId", strconv.FormatUint(uint64(objectID), 10))
	endpoint.RawQuery = q.Encode()

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.doAuthenticated(ctx, req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	// The API doc says that 204 is expected but in the reality 200 is returned.
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	return nil
}

// LookupViewID Find the DNS view with the given name within.
func (c *Client) LookupViewID(ctx context.Context, configName, viewName string) (uint, error) {
	// Lookup the entity ID of the configuration named in our properties.
	conf, err := c.GetEntityByName(ctx, 0, configName, ConfigType)
	if err != nil {
		return 0, err
	}

	view, err := c.GetEntityByName(ctx, conf.ID, viewName, ViewType)
	if err != nil {
		return 0, err
	}

	return view.ID, nil
}

// LookupParentZoneID returns the entityId of the parent zone by iterating through the root labels.
// Also return the simple name of the host.
func (c *Client) LookupParentZoneID(ctx context.Context, viewID uint, fqdn string) (uint, string, error) {
	if fqdn == "" {
		return viewID, "", nil
	}

	zones := strings.Split(strings.Trim(fqdn, "."), ".")

	name := zones[0]
	parentViewID := viewID

	for i := len(zones) - 1; i > -1; i-- {
		zone, err := c.GetEntityByName(ctx, parentViewID, zones[i], ZoneType)
		if err != nil {
			return 0, "", fmt.Errorf("could not find zone named %s: %w", name, err)
		}

		if zone == nil || zone.ID == 0 {
			break
		}

		if i > 0 {
			name = strings.Join(zones[0:i], ".")
		}

		parentViewID = zone.ID
	}

	return parentViewID, name, nil
}

func (c *Client) createEndpoint(resource string) *url.URL {
	return c.baseURL.JoinPath("Services", "REST", "v1", resource)
}

func (c *Client) doAuthenticated(ctx context.Context, req *http.Request) (*http.Response, error) {
	tok := getToken(ctx)
	if tok != "" {
		req.Header.Set(authorizationHeader, tok)
	}

	return c.HTTPClient.Do(req)
}

func newJSONRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}
