package server

import (
	"fmt"
	"net"
	"time"

	"github.com/Roman77St/chat/pkg/protocol"
)

func (s *TCPServer) Accept(ln net.Listener) (net.Conn, error) {
	conn, err := ln.Accept()
	if err != nil {
		return nil, err
	}

	// Извлекаем только IP (без порта)
	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		conn.Close()
		return nil, err
	}

	s.ipsMu.Lock()
	if s.ips[host] >= s.maxConnPerIP {
		s.ipsMu.Unlock()
		fmt.Printf("[SECURITY]: Лимит соединений превышен для IP: %s\n", host)
		conn.Close()
		return nil, fmt.Errorf("too many connections from %s", host)
	}
	s.ips[host]++
	s.ipsMu.Unlock()

	return conn, nil
}

// JoinRoom логика входа в комнату
func (s *TCPServer) JoinRoom(roomID string, conn net.Conn, timeout time.Duration) ([]net.Conn, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, exists := s.rooms[roomID]
	if !exists {
        room = &Room{
            Peers: make([]net.Conn, 0, 2),
        }

        // Запускаем таймер
        room.Timer = time.AfterFunc(timeout, func() {
            s.mu.Lock()
            defer s.mu.Unlock()

            // Если комната всё еще существует и там только 1 человек — удаляем
            if r, ok := s.rooms[roomID]; ok && len(r.Peers) < 2 {
                fmt.Printf("[DEBUG]: Комната %s удалена по таймауту\n", roomID)
                for _, p := range r.Peers {
                    p.Close()
                }
                delete(s.rooms, roomID)
            }
        })
        s.rooms[roomID] = room
    }

	if len(room.Peers) >= 2 {
		return nil, protocol.ErrRoomFull
	}

	room.Peers = append(room.Peers, conn)

	if len(room.Peers) == 2 && room.Timer != nil {
        room.Timer.Stop()
    }

	return room.Peers, nil
}

// Рассылка (Broadcast) через чистый TCP
func (s *TCPServer) Broadcast(roomID string, sender net.Conn, data []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	room := s.rooms[roomID]
	for _, peer := range room.Peers {
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

	room, ok := s.rooms[roomID]
	if !ok {
		conn.Close()
		return
	}

	for i, p := range room.Peers {
		if p == conn {
			room.Peers = append(room.Peers[:i], room.Peers[i+1:]...)
			break
		}
	}
	if len(room.Peers) == 0 {
		if room.Timer != nil {
			room.Timer.Stop()
		}
		delete(s.rooms, roomID)
	}

	// Уменьшаем счетчик соединений для IP
	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err == nil {
		s.ipsMu.Lock()
		if s.ips[host] > 0 {
			s.ips[host]--
			// Чистим мапу, если соединений больше нет
			if s.ips[host] == 0 {
				delete(s.ips, host)
			}
		}
		s.ipsMu.Unlock()
	}

	conn.Close()
}


// Alert отправляет системное сообщение всем участникам
func (s *TCPServer) Alert(roomID string, msg string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := []byte(msg)
	room := s.rooms[roomID]
	for _, peer := range room.Peers {
		_, err := peer.Write(data)
		if err != nil {
			fmt.Printf("Ошибка отправки Alert собеседнику: %v\n", err)
		}
	}
}
