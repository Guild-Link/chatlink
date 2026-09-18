package chatlink

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// SetEnabled sets the enabled state of the client and updates it in the bot state.
func (c *Client) SetEnabled(enabled bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.set("enabled", strconv.FormatBool(enabled)); err != nil {
		return err
	}
	c.enabled = enabled

	return nil
}

// SetMetadata sets the metadata for the client and updates it in the bot state.
func (c *Client) SetMetadata(metadata Metadata) error {
	metadata, err := parseMetadata(metadata)
	if err != nil {
		return fmt.Errorf("normalize client metadata: %w", err)
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("encode client metadata: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.set("metadata", string(data)); err != nil {
		return err
	}
	c.metadata = metadata

	return nil
}

// Delete removes the client from the bot state.
func (c *Client) Delete() error {
	ctx, cancel := c.worker.timeoutCtx()
	defer cancel()

	removed, err := c.worker.valkey.Do(ctx, c.worker.valkey.B().Hdel().Key(botState).Field(c.uuid).Build()).ToInt64()
	if err != nil {
		return fmt.Errorf("delete client %s: %w", c.uuid, err)
	}

	if removed == 0 {
		return fmt.Errorf("client %s does not exist", c.uuid)
	}

	return nil
}

// Broadcast sends a message through all clients except the current one.
func (c *Client) Broadcast(message string) error {
	return c.worker.broadcast(message, c.uuid)
}

// Send sends a message through the current client.
func (c *Client) Send(message string) error {
	ctx, cancel := c.worker.timeoutCtx()
	defer cancel()

	if err := send.Exec(ctx, c.worker.valkey, []string{inboundChat + c.uuid}, []string{message}).Error(); err != nil {
		return fmt.Errorf("send message through client %s: %w", c.uuid, err)
	}

	return nil
}

// UUID returns the UUID of the client.
func (c *Client) UUID() string {
	return c.uuid
}

// Username returns the username of the client.
func (c *Client) Username() string {
	return c.username
}

// Enabled returns the enabled state of the client.
func (c *Client) Enabled() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.enabled
}

// Metadata returns the metadata of the client.
func (c *Client) Metadata() Metadata {
	c.mu.Lock()
	defer c.mu.Unlock()

	metadata, err := parseMetadata(c.metadata)
	if err != nil {
		panic(err)
	}

	return metadata
}
