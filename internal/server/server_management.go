package server

import (
	"fmt"
	"net"

	"github.com/Roman77St/chat/pkg/protocol"
)

func (s *TCPServer) Accept(ln net.Listener) (net.Conn, error) {
	return ln.Accept()
}

// JoinRoom логика входа в комнату
func (s *TCPServer) JoinRoom(roomID string, conn net.Conn) ([]net.Conn, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	peers := s.rooms[roomID]
	if len(peers) >= 2 {
		return nil, protocol.ErrRoomFull
	}

	s.rooms[roomID] = append(peers, conn)
	return s.rooms[roomID], nil
}

// Рассылка (Broadcast) через чистый TCP
func (s *TCPServer) Broadcast(roomID string, sender net.Conn, data []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, peer := range s.rooms[roomID] {
		if peer != sender {
			_, err := peer.Write(data)
			if err != nil {
				fmt.Printf("Ошибка отправки: %v\n", err)
			}
		}
	}
}

// Удаление клиента
func (s *TCPServer) Remove(roomID string, conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	peers := s.rooms[roomID]
	for i, p := range peers {
		if p == conn {
			s.rooms[roomID] = append(peers[:i], peers[i+1:]...)
			break
		}
	}
	if len(s.rooms[roomID]) == 0 {
		delete(s.rooms, roomID)
	}
	conn.Close()
}


// Alert отправляет системное сообщение всем участникам
func (s *TCPServer) Alert(roomID string, msg string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := []byte(msg)
	peers := s.rooms[roomID]
	for _, peer := range peers {
		_, err := peer.Write(data)
		if err != nil {
			fmt.Printf("Ошибка отправки Alert собеседнику: %v\n", err)
		}
	}
}
