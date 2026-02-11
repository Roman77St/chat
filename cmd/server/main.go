package main

import (
	"fmt"
	"io"
	"net"

	"github.com/Roman77St/chat/internal/config"
	"github.com/Roman77St/chat/internal/network"
)

func main() {
	cnf := config.New()

	srv := network.NewTCPServer()

	// Запускаем TCP листенер
	ln, err := net.Listen("tcp", cnf.Address())
	if err != nil {
		fmt.Printf("Не удалось запустить: %v\n", err)
		return
	}

	fmt.Printf("TCP server started on %s\n", cnf.Address())

	// Основной цикл
	for {
		conn, err := srv.Accept(ln)
		if err != nil {
			// Если комната полна, оповещаем текущих участников
			if err.Error() == "room full" {
				srv.Broadcast(nil, []byte("SYSTEM: Попытка внешнего присоединения!\n"))
			} else {
				fmt.Println(err)
			}
			continue
		}
		go handleClient(conn, srv)
	}
}

func handleClient(conn net.Conn, srv *network.TCPServer) {
	defer srv.Remove(conn)

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
		srv.Broadcast(conn, buf[:n])
	}
}
