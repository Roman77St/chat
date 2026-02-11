package network

import (
	"fmt"
	"io"
	"net"
)

func (s *TCPServer) handleClient(conn net.Conn) {
	defer s.Remove(conn)

	// Буфер для чтения данных
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Ошибка чтения: %v\n", err)
			}
			break
		}

		// Пересылаем данные второму участнику
		s.Broadcast(conn, buf[:n])
	}
}

func (s *TCPServer) Run(ln net.Listener)  {
	// Основной цикл
	for {
		conn, err := s.Accept(ln)
		if err != nil {
			// Если комната полна, оповещаем текущих участников
			if err == ErrRoomFull {
				s.Broadcast(nil, []byte("SYSTEM: Попытка внешнего присоединения!\n"))
			} else {
				fmt.Println(err)
			}
			continue
		}
		go s.handleClient(conn)
	}
}