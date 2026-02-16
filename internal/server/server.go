package server

import (
	"fmt"
	"net"
	"sync"

	"github.com/Roman77St/chat/internal/config"
)

type TCPServer struct {
	mu    sync.RWMutex
	peers []net.Conn
}

func NewTCPServer() *TCPServer {
	return &TCPServer{
		peers: make([]net.Conn, 0, 2),
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

		s.ConfirmEnter()

		// Запускаем обработку конкретного клиента
		go s.handleClient(conn)

	}
}

// Внутренний метод для чтения данных от клиента
func (s *TCPServer) handleClient(conn net.Conn) {
	defer s.Remove(conn)

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			break // Выход при разрыве связи или EOF
		}
		s.Broadcast(conn, buf[:n])
	}
}
