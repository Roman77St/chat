package server

import (
	"fmt"
	"net"

	"github.com/Roman77St/chat/pkg/protocol"
)

// Проверка и принятие соединения
func (s *TCPServer) Accept(ln net.Listener) (net.Conn, error) {
	conn, err := ln.Accept()
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	isFull := len(s.peers) >= 2
	s.mu.RUnlock()

	if isFull {
		conn.Close()
		// Всё равно, какое будет сообщение. У подключенных собеседников будет сообщение "Ошибка дешифровки."
		s.Alert("[SYSTEM]: Попытка внешнего присоединения.")
		return nil, protocol.ErrRoomFull
	}

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


// Alert отправляет системное сообщение всем участникам
func (s *TCPServer) Alert(msg string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := []byte(msg)

	for _, peer := range s.peers {
		_, err := peer.Write(data)
		if err != nil {
			fmt.Printf("Ошибка отправки Alert собеседнику: %v\n", err)
		}
	}
}

func (s *TCPServer) ConfirmEnter () {
		s.mu.RLock()
        count := len(s.peers)
        s.mu.RUnlock()

		if count == 2 {
            // Даем сигнал о готовности обоим
            s.Alert(protocol.ReadySignal)
        }
}