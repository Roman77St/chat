package main

import (
	"fmt"
	"net"

	"github.com/Roman77St/chat/internal/config"
	"github.com/Roman77St/chat/internal/network"
)

func main() {
	cfg := config.New()
	srv := network.NewTCPServer()

	ln, err := net.Listen("tcp", cfg.Address())
	if err != nil {
		fmt.Printf("Критическая ошибка: %v\n", err)
		return
	}

	fmt.Printf("TCP server started on %s\n", cfg.Address())

	// Запуск сервера. Вторым аргументом передаем логику того,
	// что делать с полученным сообщением.
	srv.Run(ln, func(sender net.Conn, data []byte) {
		srv.Broadcast(sender, data)
	})
}