package socket

import (
	"bytes"
	"fmt"

	"github.com/guild-link/chatlink/pkg/proto"
)

func (s *Socket) handlePacket(p *proto.Packet) error {
	if int(p.Len) > len(p.Payload) {
		return fmt.Errorf("[%d] payload too large: %d", p.PacketType, p.Len)
	}

	if p.ID < 0 {
		return fmt.Errorf("[%d] inbound packets cannot be negative: %d", p.PacketType, p.ID)
	}

	payload := p.Payload[:int(p.Len)]
	output := proto.Output{ID: p.ID, Content: string(payload)}

	switch p.PacketType {
	case proto.PacketMessage:
		handle(s.messageHandlers, proto.Message(output))

	case proto.PacketSignedMessage:
		handle(s.signedMessageHandlers, proto.SignedMessage(output))

	case proto.PacketDisconnect:
		handle(s.disconnectHandlers, proto.Disconnect(output))

	case proto.PacketBotInfo:
		bots, err := parseBots(payload)
		if err != nil {
			return err
		}

		handle(s.botInfoHandlers, bots)

	case proto.PacketStop:
		return fmt.Errorf("[%d] packet cannot be used inbound", p.PacketType)

	default:
		return fmt.Errorf("[%d] unknown packet type", p.PacketType)
	}

	return nil
}

func parseBots(payload []byte) ([]proto.BotInfo, error) {
	if len(payload)%proto.MaxUsernameSize != 0 {
		return nil, fmt.Errorf("invalid bot info payload length: %d", len(payload))
	}

	count := len(payload) / proto.MaxUsernameSize
	if count > 127 {
		return nil, fmt.Errorf("too many usernames: %d", count)
	}

	bots := make([]proto.BotInfo, count)
	for i := range bots {
		start := i * proto.MaxUsernameSize
		username := payload[start : start+proto.MaxUsernameSize]

		if end := bytes.IndexByte(username, 0); end >= 0 {
			username = username[:end]
		}

		bots[i] = proto.BotInfo{ID: int8(i + 1), Content: string(username)}
	}

	return bots, nil
}

func handle[T any](handlers []func(T), value T) {
	for _, handler := range handlers {
		handler(value)
	}
}
