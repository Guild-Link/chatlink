package socket

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/guild-link/chatlink/pkg/proto"
)

func (s *Socket) SendPacket(message *proto.Packet) error {
	if message == nil {
		return fmt.Errorf("packet is nil")
	}

	if int(message.Len) > len(message.Payload) {
		return fmt.Errorf("payload too large: %d", message.Len)
	}

	if message.PacketType == proto.PacketBotInfo {
		return fmt.Errorf("[%d] packet cannot be used outbound", message.PacketType)
	}

	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("no active client connection")
	}

	if s.timeout > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(s.timeout)); err != nil {
			s.closeConn(conn)
			return fmt.Errorf("failed to set write deadline: %w", err)
		}
	}

	if err := binary.Write(conn, binary.LittleEndian, message); err != nil {
		s.closeConn(conn)
		return fmt.Errorf("failed to send packet: %w", err)
	}

	return nil
}
