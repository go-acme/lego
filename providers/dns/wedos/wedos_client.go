package wedos

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type WResponsePayload struct {
	Code        int             `json:"code"`
	Result      string          `json:"result"`
	Timestamp   int             `json:"timestamp"`
	SvTRID      string          `json:"svTRID"`
	Command     string          `json:"command"`
	Data        json.RawMessage `json:"data"`
	DNSRowsList []WDNSRow
}

type WDNSRow struct {
	ID            string      `json:"ID"`
	Name          string      `json:"name"`
	TTL           json.Number `json:"ttl" type:"integer"`
	RDType        string      `json:"rdtype"`
	RData         string      `json:"rdata"`
	ChangedDate   string      `json:"changed_date"`
	AuthorComment string      `json:"author_comment"`
}

func askWedos(ctx context.Context, httpClient *http.Client, userName string, wapiPass string, command string, payload interface{}) (*WResponsePayload, error) {
	requestObject := map[string]interface{}{
		"request": map[string]interface{}{
			"user":    userName,
			"auth":    authToken(userName, wapiPass),
			"command": command,
			"data":    payload,
		},
	}

	jsonBytes, err := json.Marshal(requestObject)
	if err != nil {
		return nil, err
	}

	form := url.Values{"request": {string(jsonBytes)}}
	requestBody := strings.NewReader(form.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.wedos.com/wapi/json", requestBody)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseWrapper := struct {
		Response WResponsePayload `json:"response"`
	}{}

	err = json.Unmarshal(body, &responseWrapper)
	if err != nil {
		return nil, err
	}

	if responseWrapper.Response.Code != 1000 {
		return nil, fmt.Errorf("wedos responded with error code %d = %s", responseWrapper.Response.Code, responseWrapper.Response.Result)
	}

	if responseWrapper.Response.Command == "dns-rows-list" {
		arrayWrapper := struct {
			Rows []WDNSRow `json:"row"`
		}{}

		err = json.Unmarshal(responseWrapper.Response.Data, &arrayWrapper)
		if err != nil {
			return nil, err
		}
		responseWrapper.Response.DNSRowsList = arrayWrapper.Rows
	}

	return &responseWrapper.Response, err
}

func authToken(userName string, wapiPass string) string {
	return sha1string(userName + sha1string(wapiPass) + czechHourString())
}

func sha1string(txt string) string {
	h := sha1.New()
	_, _ = io.WriteString(h, txt)
	return fmt.Sprintf("%x", h.Sum(nil))
}
