package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultBrokerURL is the token broker endpoint (prod). The broker validates the
// API key against the tenant and returns a short-lived download token.
const DefaultBrokerURL = "https://a93sb53vya.execute-api.us-east-1.amazonaws.com/prod/sdk/token"

// brokerToken is the ephemeral credential used to download the SDK.
type brokerToken struct {
	AuthorizationToken string `json:"authorizationToken"`
	RegistryEndpoint   string `json:"registryEndpoint"`
	Domain             string `json:"domain"`
	Repository         string `json:"repository"`
	Expiration         string `json:"expiration"`
}

// fetchToken exchanges (tenant, apiKey) for a download token via the broker.
func fetchToken(brokerURL, tenant, apiKey string) (*brokerToken, error) {
	req, err := http.NewRequest(http.MethodGet, brokerURL, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(tenant, apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach the broker: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	// The backend wraps responses in {success, message, data}. Also accept flat.
	var env struct {
		Success bool         `json:"success"`
		Message string       `json:"message"`
		Error   string       `json:"error"`
		Data    *brokerToken `json:"data"`
	}
	_ = json.Unmarshal(body, &env)

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid API key or tenant (check --api-key and --tenant)")
	}
	if resp.StatusCode != http.StatusOK {
		msg := env.Message
		if msg == "" {
			msg = env.Error
		}
		if msg == "" {
			msg = string(body)
		}
		return nil, fmt.Errorf("broker responded %d: %s", resp.StatusCode, msg)
	}

	tok := env.Data
	if tok == nil { // flat form
		tok = &brokerToken{}
		if err := json.Unmarshal(body, tok); err != nil {
			return nil, fmt.Errorf("unreadable broker response: %w", err)
		}
	}
	if tok.AuthorizationToken == "" || tok.RegistryEndpoint == "" {
		return nil, fmt.Errorf("the broker did not return a token or endpoint")
	}
	return tok, nil
}
