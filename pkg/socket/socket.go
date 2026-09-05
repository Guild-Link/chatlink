package socket

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/guild-link/chatlink/pkg/proto"
)

func NewSocket(opts ...Option) (*Socket, error) {
	sock := &Socket{
		path:    "/tmp/chatlink.sock",
		timeout: 15 * time.Second,
	}

	for _, opt := range opts {
		opt(sock)
	}

	if err := os.Remove(sock.path); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to remove existing socket file: %w", err)
	}

	listener, err := net.Listen("unix", sock.path)
	if err != nil {
		return nil, err
	}

	sock.listener = listener
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			if err := sock.setConn(conn); err != nil {
				return
			}

			sock.acceptConn(conn)
		}
	}()

	return sock, nil
}

func (s *Socket) acceptConn(conn net.Conn) {
	defer s.closeConn(conn)

	for {
		p := proto.Packet{}
		if err := binary.Read(conn, binary.LittleEndian, &p); err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Println("Error while reading from the socket", err)
			}
			return
		}

		if err := s.handlePacket(&p); err != nil {
			log.Println(err)
			return
		}
	}
}
