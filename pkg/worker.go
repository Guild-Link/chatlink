package chatlink

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/valkey-io/valkey-go"
)

func New(ctx context.Context, valkey valkey.Client, opts ...Option) *Worker {
	w := &Worker{
		timeout: 15 * time.Second,
		valkey:  valkey,
		ctx:     ctx,
	}

	for _, opt := range opts {
		opt(w)
	}

	if len(w.messageHandlers) > 0 {
		go func() {
			for w.ctx.Err() == nil {
				if err := w.handleMessages(); err != nil && w.ctx.Err() == nil {
					log.Printf("error handling messages: %v", err)
					time.Sleep(time.Second)
				}
			}
		}()
	}

	return w
}

func (w *Worker) Clients() ([]*Client, error) {
	ctx, cancel := w.timeoutCtx()
	defer cancel()

	bots, err := w.valkey.Do(ctx, w.valkey.B().Hgetall().Key(botState).Build()).AsStrMap()
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}

	clients := make([]*Client, 0, len(bots))
	for uuid, data := range bots {
		client, err := w.newClient(uuid, data)
		if err != nil {
			return nil, err
		}

		if client.username != "" {
			clients = append(clients, client)
		}
	}

	return clients, nil
}

func (w *Worker) AddClient(uuid, token string, metadata Metadata) error {
	if !strings.HasPrefix(token, "hpke:") || len(token) == len("hpke:") {
		return fmt.Errorf("invalid token format")
	}

	metadata, err := parseMetadata(metadata)
	if err != nil {
		return fmt.Errorf("normalize client metadata: %w", err)
	}

	data, err := json.Marshal(clientData{
		Metadata: metadata,
		Token:    token,
		Enabled:  true,
	})
	if err != nil {
		return fmt.Errorf("encode client %s: %w", uuid, err)
	}

	ctx, cancel := w.timeoutCtx()
	defer cancel()

	added, err := w.valkey.Do(ctx, w.valkey.B().Hsetnx().Key(botState).Field(uuid).Value(string(data)).Build()).ToInt64()
	if err != nil {
		return fmt.Errorf("add client %s: %w", uuid, err)
	}

	if added == 0 {
		return fmt.Errorf("client %s already exists", uuid)
	}

	return nil
}

func (w *Worker) Broadcast(message string) error {
	return w.broadcast(message, "")
}
