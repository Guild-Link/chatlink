package socket

import (
	"time"

	"github.com/guild-link/chatlink/pkg/proto"
)

type Option func(*Socket)

func validateOpt(ok bool) {
	if !ok {
		panic("invalid option")
	}
}

func WithSignedMessageHandler(fn func(proto.SignedMessage)) Option {
	validateOpt(fn != nil)
	return func(s *Socket) {
		s.signedMessageHandlers = append(s.signedMessageHandlers, fn)
	}
}

func WithDisconnectHandler(fn func(proto.Disconnect)) Option {
	validateOpt(fn != nil)
	return func(s *Socket) {
		s.disconnectHandlers = append(s.disconnectHandlers, fn)
	}
}

func WithBotInfoHandler(fn func([]proto.BotInfo)) Option {
	validateOpt(fn != nil)
	return func(s *Socket) {
		s.botInfoHandlers = append(s.botInfoHandlers, fn)
	}
}

func WithMessageHandler(fn func(proto.Message)) Option {
	validateOpt(fn != nil)
	return func(s *Socket) {
		s.messageHandlers = append(s.messageHandlers, fn)
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(s *Socket) {
		s.timeout = timeout
	}
}

func WithSocketPath(path string) Option {
	validateOpt(path != "")

	return func(s *Socket) {
		s.path = path
	}
}
