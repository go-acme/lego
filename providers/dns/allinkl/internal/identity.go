package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// authEndpoint represents the Identity API endpoint to call.
const authEndpoint = "https://kasapi.kasserver.com/soap/KasAuth.php"

type token string

const tokenKey token = "token"

// Identifier generates credential tokens.
type Identifier struct {
	login    string
	password string

	authEndpoint string
	HTTPClient   *http.Client
}

// NewIdentifier creates a new Identifier.
func NewIdentifier(login, password string) *Identifier {
	return &Identifier{
		login:        login,
		password:     password,
		authEndpoint: authEndpoint,
		HTTPClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Authentication Creates a credential token.
// - sessionLifetime: Validity of the token in seconds.
// - sessionUpdateLifetime: with `true` the session is extended with every request.
func (c *Identifier) Authentication(ctx context.Context, sessionLifetime int, sessionUpdateLifetime bool) (string, error) {
	sul := "N"
	if sessionUpdateLifetime {
		sul = "Y"
	}

	ar := AuthRequest{
		Login:                 c.login,
		AuthData:              c.password,
		AuthType:              "plain",
		SessionLifetime:       sessionLifetime,
		SessionUpdateLifetime: sul,
	}

	body, err := json.Marshal(ar)
	if err != nil {
		return "", fmt.Errorf("failed to create request JSON body: %w", err)
	}

	payload := []byte(strings.TrimSpace(fmt.Sprintf(kasAuthEnvelope, body)))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.authEndpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("unable to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	envlp, err := decodeXML[KasAuthEnvelope](resp.Body)
	if err != nil {
		return "", err
	}

	if envlp.Body.Fault != nil {
		return "", envlp.Body.Fault
	}

	return envlp.Body.KasAuthResponse.Return.Text, nil
}

func WithContext(ctx context.Context, credential string) context.Context {
	return context.WithValue(ctx, tokenKey, credential)
}

func getToken(ctx context.Context) string {
	credential, ok := ctx.Value(tokenKey).(string)
	if !ok {
		return ""
	}

	return credential
}
