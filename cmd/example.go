package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/guild-link/chatlink/pkg/proto"
	"github.com/guild-link/chatlink/pkg/socket"
)

func handleBotInfo(bots []proto.BotInfo) {
	fmt.Printf("Bots:\n")
	for _, bot := range bots {
		fmt.Printf("[%d] Username: %s\n", bot.ID, bot.Content)
	}
}

func handleSignedMessage(msg proto.SignedMessage) {
	fmt.Printf("[%d] Signed Message: %s\n", msg.ID, msg.Content)
}

func handleMessage(msg proto.Message) {
	fmt.Printf("[%d] Message: %s\n", msg.ID, msg.Content)
}

func handleDisconnect(reason proto.Disconnect) {
	fmt.Printf("[%d] Disconnected: %s\n", reason.ID, reason.Content)
}

func handleStdin(s *socket.Socket) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}

		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}

		pack, err := proto.MessagePacket(0, text)
		if err != nil {
			fmt.Println("Error creating packet:", err)
			continue
		}

		if err := s.SendPacket(pack); err != nil {
			fmt.Println("Error sending packet:", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Failed to read stdin:", err)
	}
}

func main() {
	opts := []socket.Option{
		socket.WithSignedMessageHandler(handleSignedMessage),
		socket.WithSocketPath("/tmp/chatlink.sock"),
		socket.WithDisconnectHandler(handleDisconnect),
		socket.WithMessageHandler(handleMessage),
		socket.WithBotInfoHandler(handleBotInfo),
		socket.WithTimeout(15 * time.Second),
	}

	sock, err := socket.NewSocket(opts...)
	if err != nil {
		fmt.Println("Error creating socket:", err)
		return
	}
	defer sock.Close()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		<-interrupt
		sock.Close()
		os.Exit(0)
	}()

	handleStdin(sock)
}
