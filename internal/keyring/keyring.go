package keyring

import (
	"github.com/zalando/go-keyring"
)

// ServiceKey represents a unique identifier for a service's API key
type ServiceKey struct {
	Service  string
	Username string
}

// Common service identifiers
var (
	Anthropic = ServiceKey{Service: "gitai", Username: "anthropic"}
)

// StoreAPIKey stores an API key in the system keyring for a specific service
func StoreAPIKey(key ServiceKey, apiKey string) error {
	return keyring.Set(key.Service, key.Username, apiKey)
}

// GetAPIKey retrieves an API key from the system keyring for a specific service
func GetAPIKey(key ServiceKey) (string, error) {
	return keyring.Get(key.Service, key.Username)
}

// DeleteAPIKey removes an API key from the system keyring for a specific service
func DeleteAPIKey(key ServiceKey) error {
	return keyring.Delete(key.Service, key.Username)
}
