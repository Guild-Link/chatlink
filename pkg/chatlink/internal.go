package chatlink

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/valkey-io/valkey-go"
)

// parseMetadata normalizes and deep-copies metadata.
func parseMetadata(metadata Metadata) (Metadata, error) {
	if metadata == nil {
		metadata = Metadata{}
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}

	var parsed Metadata
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}

	return parsed, nil
}

func (w *Worker) timeoutCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(w.ctx, w.timeout)
}

func (w *Worker) broadcast(message, exclude string) error {
	ctx, cancel := w.timeoutCtx()
	defer cancel()

	if err := broadcast.Exec(ctx, w.valkey, []string{botState}, []string{message, exclude, inboundChat}).Error(); err != nil {
		return fmt.Errorf("broadcast message: %w", err)
	}

	return nil
}

func (w *Worker) deleteMessage(id string) error {
	ctx, cancel := w.timeoutCtx()
	defer cancel()

	if err := w.valkey.Do(ctx, w.valkey.B().Xdel().Key(outboundChat).Id(id).Build()).Error(); err != nil {
		return fmt.Errorf("delete message %s: %w", id, err)
	}

	return nil
}

func (w *Worker) getClient(uuid string) (*Client, error) {
	ctx, cancel := w.timeoutCtx()
	defer cancel()

	data, err := w.valkey.Do(ctx, w.valkey.B().Hget().Key(botState).Field(uuid).Build()).ToString()
	if valkey.IsValkeyNil(err) {
		return &Client{uuid: uuid, worker: w}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("get client %s: %w", uuid, err)
	}

	return w.newClient(uuid, data)
}

func (w *Worker) newClient(uuid, data string) (*Client, error) {
	var d clientData
	if err := json.Unmarshal([]byte(data), &d); err != nil {
		return nil, fmt.Errorf("decode client %s: %w", uuid, err)
	}

	if d.Metadata == nil {
		d.Metadata = Metadata{}
	}

	return &Client{
		metadata: d.Metadata,
		username: d.Username,
		enabled:  d.Enabled,
		uuid:     uuid,
		worker:   w,
	}, nil
}

func (w *Worker) handleMessages() error {
	streams, err := w.valkey.Do(w.ctx, w.valkey.B().Xread().Count(32).Block(w.timeout.Milliseconds()).Streams().Key(outboundChat).Id("0-0").Build()).AsXRead()
	if valkey.IsValkeyNil(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read outbound messages: %w", err)
	}

	for _, entry := range streams[outboundChat] {
		message, hasMessage := entry.FieldValues["message"]
		uuid, hasUUID := entry.FieldValues["uuid"]

		if hasMessage && hasUUID {
			client, err := w.getClient(uuid)
			if err != nil {
				return err
			}

			for _, handler := range w.messageHandlers {
				handler(client, message)
			}
		}

		if err := w.deleteMessage(entry.ID); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) set(field, value string) error {
	ctx, cancel := c.worker.timeoutCtx()
	defer cancel()

	updated, err := update.Exec(ctx, c.worker.valkey, []string{botState}, []string{c.uuid, field, value}).ToInt64()
	if err != nil {
		return fmt.Errorf("update client %s field %s: %w", c.uuid, field, err)
	}

	if updated == 0 {
		return fmt.Errorf("client %s does not exist", c.uuid)
	}

	return nil
}
