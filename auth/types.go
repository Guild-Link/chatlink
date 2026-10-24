package auth

import "crypto/ed25519"

type Account struct {
	CreatedAt uint64 `json:"created_at"`
	Username  string `json:"username"`
	Token     string `json:"token"`
	UUID      string `json:"uuid"`
	Key       string `json:"key"`
}

type Client struct {
	EncKey ed25519.PublicKey
	SigKey ed25519.PublicKey
	URL    string
}
