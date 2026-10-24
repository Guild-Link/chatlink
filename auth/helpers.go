package auth

import (
	"bytes"
	"compress/gzip"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"
)

func parsePublicKey(value string) (ed25519.PublicKey, error) {
	key, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(value, "="))
	if err != nil {
		return nil, fmt.Errorf("decode Ed25519 public key: %w", err)
	}

	if len(key) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("Ed25519 public key must be %d bytes", ed25519.PublicKeySize)
	}

	return ed25519.PublicKey(key), nil
}

func decodeToken(token string) (signature, payload []byte, err error) {
	compressed, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, nil, fmt.Errorf("decode token: %w", err)
	}

	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, nil, fmt.Errorf("open compressed token: %w", err)
	}
	defer gz.Close()

	data, err := io.ReadAll(io.LimitReader(gz, 1<<20))
	if err != nil {
		return nil, nil, fmt.Errorf("decompress token: %w", err)
	}

	if len(data) < ed25519.SignatureSize {
		return nil, nil, fmt.Errorf("token is too short")
	}

	return data[:ed25519.SignatureSize], data[ed25519.SignatureSize:], nil
}

func validateAccount(account *Account, knownKey ed25519.PublicKey) error {
	if account.Token == "" || account.UUID == "" || account.Username == "" {
		return fmt.Errorf("token is missing account details")
	}

	age := time.Since(time.UnixMilli(int64(account.CreatedAt)))
	if age > 5*time.Minute {
		return fmt.Errorf("token is older than 5 minutes")
	}

	if age < -time.Minute {
		return fmt.Errorf("token is from the future")
	}

	key, err := parsePublicKey(account.Key)
	if err != nil || !bytes.Equal(key, knownKey) {
		return fmt.Errorf("token was not encrypted for this instance")
	}

	return nil
}
