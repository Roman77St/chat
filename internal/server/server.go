package server

import (
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Roman77St/chat/internal/config"
	"github.com/Roman77St/chat/pkg/protocol"
)
type Room struct {
    Peers   []net.Conn
    Timer   *time.Timer
}

type TCPServer struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewTCPServer() *TCPServer {
	return &TCPServer{
		rooms: make(map[string]*Room),
	}
}

func Run(cfg *config.Config) {

	s := NewTCPServer()

	ln, err := net.Listen("tcp", cfg.Address())
	if err != nil {
		fmt.Printf("Критическая ошибка: %v\n", err)
		return
	}

	fmt.Printf("TCP server started on %s\n", cfg.Address())


	for {
		conn, err := s.Accept(ln)
		if err != nil {
			fmt.Printf("Ошибка сетевого уровня: %v\n", err)
			continue
		}

		go s.handleClient(conn, cfg.RoomTimeout)

	}
}

func (s *TCPServer) handleClient(conn net.Conn, timeout time.Duration) {
	rawID, err := protocol.ReadRoomID(conn)
	if err != nil {
		conn.Close()
		return
	}
	roomID := hex.EncodeToString(rawID)

	peers, err := s.JoinRoom(roomID, conn, timeout)
	if err != nil {
		// Если комната полна — тихо закрываем
		conn.Close()
		s.Alert(roomID, "[SYSTEM]: Попытка входа в полную комнату.")
		return
	}

	defer s.Remove(roomID, conn)

	// Если пара собралась — шлем READY
	if len(peers) == 2 {
		s.Alert(roomID, protocol.ReadySignal)
	}

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			break // Выход при разрыве связи или EOF
		}
		s.Broadcast(roomID, conn, buf[:n])
	}
}
