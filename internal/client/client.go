package client

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

// Start запускает консольный клиент
func Start(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("Ошибка подключения: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("Подключено...")

	// Горутина для чтения входящих данных (от собеседника)
	go func() {
		// Используем буфер побольше для приема
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					fmt.Println("\nСоединение закрыто сервером.")
				} else {
					fmt.Printf("\nОшибка чтения: %v\n", err)
				}
				os.Exit(0) // Завершаем клиент при обрыве связи
				return
			}
			// Печатаем полученное сообщение
			fmt.Printf("\nСобеседник: %s\n> ", string(buf[:n]))
		}
	}()

	// Основной цикл: чтение ввода из консоли и отправка в сокет
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}

		_, err := conn.Write([]byte(text))
		if err != nil {
			fmt.Printf("Не удалось отправить: %v\n", err)
			break
		}
		fmt.Print("> ")
	}
}