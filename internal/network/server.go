package network

import (
	"fmt"
	"net"
)

func (s *TCPServer) Run(ln net.Listener, onMsg func(conn net.Conn, data []byte)) {
	for {
		conn, err := s.Accept(ln)
		if err != nil {
			// Если ошибка вызвана тем, что комната полна
			if err.Error() == "room full" {
				s.Alert("Попытка внешнего присоединения!")
			} else {
				fmt.Printf("Ошибка сетевого уровня: %v\n", err)
			}
			continue
		}

		s.mu.RLock()
        count := len(s.peers)
        s.mu.RUnlock()

		if count == 2 {
            // Даем сигнал обоим: "Поехали!"
            s.Alert("READY")
        }

		// Запускаем обработку конкретного клиента
		go s.handleClient(conn, onMsg)
	}
}

// Внутренний метод для чтения данных от клиента
func (s *TCPServer) handleClient(conn net.Conn, onMsg func(conn net.Conn, data []byte)) {
	defer s.Remove(conn)

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			break // Выход при разрыве связи или EOF
		}

		// Вызываем callback (рассылку)
		onMsg(conn, buf[:n])
	}
}

// Alert отправляет системное сообщение всем участникам
func (s *TCPServer) Alert(msg string) {
	s.Broadcast(nil, []byte("[SYSTEM]: READY "))
}