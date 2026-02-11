package network

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

var ErrRoomFull = errors.New("room full")

type TCPServer struct {
	mu    sync.RWMutex
	peers []net.Conn
}

func NewTCPServer() *TCPServer {
	return &TCPServer{
		peers: make([]net.Conn, 0, 2),
	}
}

// Проверка и принятие соединения
func (s *TCPServer) Accept(ln net.Listener) (net.Conn, error) {
	conn, err := ln.Accept()
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	if len(s.peers) >= 2 {
		s.mu.RUnlock()
		conn.Close() // "Тихое" закрытие для лишнего участника
		return nil, ErrRoomFull
	}
	s.mu.RUnlock()

	s.mu.Lock()
	s.peers = append(s.peers, conn)
	s.mu.Unlock()

	return conn, nil
}

// Рассылка (Broadcast) через чистый TCP
func (s *TCPServer) Broadcast(sender net.Conn, data []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, peer := range s.peers {
		if peer != sender {
			_, err := peer.Write(data)
			if err != nil {
				fmt.Printf("Ошибка отправки: %v\n", err)
			}
		}
	}
}

// Удаление клиента
func (s *TCPServer) Remove(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.peers {
		if p == conn {
			s.peers = append(s.peers[:i], s.peers[i+1:]...)
			break
		}
	}
	conn.Close()
}