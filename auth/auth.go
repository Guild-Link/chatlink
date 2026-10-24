package auth

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
)

func New(endpoint, sigKeyValue, encKeyValue string) (*Client, error) {
	sigKey, err := parsePublicKey(sigKeyValue)
	if err != nil {
		return nil, fmt.Errorf("invalid mc-auth signing key: %w", err)
	}

	encKey, err := parsePublicKey(encKeyValue)
	if err != nil {
		return nil, fmt.Errorf("invalid mc-auth encryption key: %w", err)
	}

	return &Client{
		URL:    fmt.Sprintf("%s?key=%s", endpoint, encKeyValue),
		SigKey: sigKey,
		EncKey: encKey,
	}, nil
}

func (c *Client) ParseToken(token string) (*Account, error) {
	signature, payload, err := decodeToken(token)
	if err != nil {
		return nil, err
	}

	if !ed25519.Verify(c.SigKey, payload, signature) {
		return nil, fmt.Errorf("invalid token signature")
	}

	var account Account
	if err := json.Unmarshal(payload, &account); err != nil {
		return nil, fmt.Errorf("decode token JSON: %w", err)
	}

	if err := validateAccount(&account, c.EncKey); err != nil {
		return nil, err
	}

	return &account, nil
}
