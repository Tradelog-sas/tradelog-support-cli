package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultBrokerURL es el endpoint del broker de tokens (prod). El broker valida
// la api key contra el tenant y devuelve un token de CodeArtifact de corta vida.
const DefaultBrokerURL = "https://a93sb53vya.execute-api.us-east-1.amazonaws.com/prod/sdk/token"

// brokerToken es la credencial efímera para descargar el SDK del registry.
type brokerToken struct {
	AuthorizationToken string `json:"authorizationToken"`
	RegistryEndpoint   string `json:"registryEndpoint"`
	Domain             string `json:"domain"`
	Repository         string `json:"repository"`
	Expiration         string `json:"expiration"`
}

// fetchToken intercambia (tenant, apiKey) por un token de CodeArtifact vía el broker.
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
		return nil, fmt.Errorf("no se pudo contactar el broker: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	// El backend envuelve en {success, message, data}. Aceptamos también forma plana.
	var env struct {
		Success bool         `json:"success"`
		Message string       `json:"message"`
		Error   string       `json:"error"`
		Data    *brokerToken `json:"data"`
	}
	_ = json.Unmarshal(body, &env)

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("api key o tenant inválidos (revisa --api-key y --tenant)")
	}
	if resp.StatusCode != http.StatusOK {
		msg := env.Message
		if msg == "" {
			msg = env.Error
		}
		if msg == "" {
			msg = string(body)
		}
		return nil, fmt.Errorf("broker respondió %d: %s", resp.StatusCode, msg)
	}

	tok := env.Data
	if tok == nil { // forma plana
		tok = &brokerToken{}
		if err := json.Unmarshal(body, tok); err != nil {
			return nil, fmt.Errorf("respuesta del broker ilegible: %w", err)
		}
	}
	if tok.AuthorizationToken == "" || tok.RegistryEndpoint == "" {
		return nil, fmt.Errorf("el broker no devolvió token o endpoint")
	}
	return tok, nil
}
