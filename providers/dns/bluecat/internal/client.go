package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Object types.
const (
	ConfigType = "Configuration"
	ViewType   = "View"
	ZoneType   = "Zone"
	TXTType    = "TXTRecord"
)

type Client struct {
	HTTPClient *http.Client

	baseURL string

	token    string
	tokenExp *regexp.Regexp
}

func NewClient(baseURL string) *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
		tokenExp:   regexp.MustCompile("BAMAuthToken: [^ ]+"),
	}
}

// Login Logs in as API user.
// Authenticates and receives a token to be used in for subsequent requests.
// https://docs.bluecatnetworks.com/r/Address-Manager-API-Guide/GET/v1/login/9.1.0
func (c *Client) Login(username, password string) error {
	queryArgs := map[string]string{
		"username": username,
		"password": password,
	}

	resp, err := c.sendRequest(http.MethodGet, "login", nil, queryArgs)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Resource:   "login",
			Message:    string(data),
		}
	}

	authBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	authResp := string(authBytes)
	if strings.Contains(authResp, "Authentication Error") {
		return fmt.Errorf("request failed: %s", strings.Trim(authResp, `"`))
	}

	// Upon success, API responds with "Session Token-> BAMAuthToken: dQfuRMTUxNjc3MjcyNDg1ODppcGFybXM= <- for User : username"
	c.token = c.tokenExp.FindString(authResp)

	return nil
}

// Logout Logs out of the current API session.
// https://docs.bluecatnetworks.com/r/Address-Manager-API-Guide/GET/v1/logout/9.1.0
func (c *Client) Logout() error {
	if c.token == "" {
		// nothing to do
		return nil
	}

	resp, err := c.sendRequest(http.MethodGet, "logout", nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Resource:   "logout",
			Message:    string(data),
		}
	}

	authBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	authResp := string(authBytes)
	if !strings.Contains(authResp, "successfully") {
		return fmt.Errorf("request failed to delete session: %s", strings.Trim(authResp, `"`))
	}

	c.token = ""

	return nil
}

// Deploy the DNS config for the specified entity to the authoritative servers.
// https://docs.bluecatnetworks.com/r/Address-Manager-API-Guide/POST/v1/quickDeploy/9.1.0
func (c *Client) Deploy(entityID uint) error {
	queryArgs := map[string]string{
		"entityId": strconv.FormatUint(uint64(entityID), 10),
	}

	resp, err := c.sendRequest(http.MethodPost, "quickDeploy", nil, queryArgs)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// The API doc says that 201 is expected but in the reality 200 is return.
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Resource:   "quickDeploy",
			Message:    string(data),
		}
	}

	return nil
}

// AddEntity A generic method for adding configurations, DNS zones, and DNS resource records.
// https://docs.bluecatnetworks.com/r/Address-Manager-API-Guide/POST/v1/addEntity/9.1.0
func (c *Client) AddEntity(parentID uint, entity Entity) (uint64, error) {
	queryArgs := map[string]string{
		"parentId": strconv.FormatUint(uint64(parentID), 10),
	}

	resp, err := c.sendRequest(http.MethodPost, "addEntity", entity, queryArgs)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return 0, &APIError{
			StatusCode: resp.StatusCode,
			Resource:   "addEntity",
			Message:    string(data),
		}
	}

	addTxtBytes, _ := io.ReadAll(resp.Body)

	// addEntity responds only with body text containing the ID of the created record
	addTxtResp := string(addTxtBytes)
	id, err := strconv.ParseUint(addTxtResp, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("addEntity request failed: %s", addTxtResp)
	}

	return id, nil
}

// GetEntityByName Returns objects from the database referenced by their database ID and with its properties fields populated.
// https://docs.bluecatnetworks.com/r/Address-Manager-API-Guide/GET/v1/getEntityById/9.1.0
func (c *Client) GetEntityByName(parentID uint, name, objType string) (*EntityResponse, error) {
	queryArgs := map[string]string{
		"parentId": strconv.FormatUint(uint64(parentID), 10),
		"name":     name,
		"type":     objType,
	}

	resp, err := c.sendRequest(http.MethodGet, "getEntityByName", nil, queryArgs)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Resource:   "getEntityByName",
			Message:    string(data),
		}
	}

	var txtRec EntityResponse
	if err = json.NewDecoder(resp.Body).Decode(&txtRec); err != nil {
		return nil, fmt.Errorf("JSON decode: %w", err)
	}

	return &txtRec, nil
}

// Delete Deletes an object using the generic delete method.
// https://docs.bluecatnetworks.com/r/Address-Manager-API-Guide/DELETE/v1/delete/9.1.0
func (c *Client) Delete(objectID uint) error {
	queryArgs := map[string]string{
		"objectId": strconv.FormatUint(uint64(objectID), 10),
	}

	resp, err := c.sendRequest(http.MethodDelete, "delete", nil, queryArgs)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// The API doc says that 204 is expected but in the reality 200 is return.
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Resource:   "delete",
			Message:    string(data),
		}
	}

	return nil
}

// LookupViewID Find the DNS view with the given name within.
func (c *Client) LookupViewID(configName, viewName string) (uint, error) {
	// Lookup the entity ID of the configuration named in our properties.
	conf, err := c.GetEntityByName(0, configName, ConfigType)
	if err != nil {
		return 0, err
	}

	view, err := c.GetEntityByName(conf.ID, viewName, ViewType)
	if err != nil {
		return 0, err
	}

	return view.ID, nil
}

// LookupParentZoneID Return the entityId of the parent zone by recursing from the root view.
// Also return the simple name of the host.
func (c *Client) LookupParentZoneID(viewID uint, fqdn string) (uint, string, error) {
	if fqdn == "" {
		return viewID, "", nil
	}

	zones := strings.Split(strings.Trim(fqdn, "."), ".")

	name := zones[0]
	parentViewID := viewID

	for i := len(zones) - 1; i > -1; i-- {
		zone, err := c.GetEntityByName(parentViewID, zones[i], ZoneType)
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

// Send a REST request, using query parameters specified.
// The Authorization header will be set if we have an active auth token.
func (c *Client) sendRequest(method, resource string, payload interface{}, queryParams map[string]string) (*http.Response, error) {
	url := fmt.Sprintf("%s/Services/REST/v1/%s", c.baseURL, resource)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}

	q := req.URL.Query()
	for k, v := range queryParams {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()

	return c.HTTPClient.Do(req)
}
