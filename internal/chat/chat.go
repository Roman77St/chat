package chat

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Server struct {
	mu    sync.RWMutex
	peers []*websocket.Conn
}

func NewServer() *Server {
	return &Server{
		peers: make([]*websocket.Conn, 0, 2),
	}
}

// Проверка количества подключений и апгрейд
func (s *Server) Join(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	s.mu.RLock()
	if len(s.peers) >= 2 {
		s.mu.RUnlock()
		return nil, ErrRoomFull
	}
	s.mu.RUnlock()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.peers = append(s.peers, conn)
	s.mu.Unlock()

	return conn, nil
}

// Функция завершения (Cleanup)
func (s *Server) Leave(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.peers {
		if p == conn {
			s.peers = append(s.peers[:i], s.peers[i+1:]...)
			break
		}
	}
	conn.Close()
	fmt.Println("Клиент покинул чат")
}

// Логика рассылки (Broadcast)
func (s *Server) Broadcast(sender *websocket.Conn, message []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, peer := range s.peers {
		if peer != sender {
			// TODO здесь стоит добавить таймаут на запись ??
			_ = peer.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// Системное сообщение участникам
func (s *Server) Alert(message string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, peer := range s.peers {
		// Отправляем системное уведомление
		_ = peer.WriteMessage(websocket.TextMessage, []byte("SYSTEM: " + message))
	}
}

var ErrRoomFull = errors.New("external connection attempt")