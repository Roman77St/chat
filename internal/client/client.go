package client

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/Roman77St/chat/internal/security"
)

func Start(addr string) {
    conn, _ := net.Dial("tcp", addr)
    defer conn.Close()

    // 1. Сначала вводим пароль (он нужен для генерации ключа)
    fmt.Print("Введите код доступа: ")
    var password string
    fmt.Scanln(&password)

    fmt.Println("Ожидание синхронизации...")

    // Сигнал READY от сервера
    readyBuf := make([]byte, 16)
    io.ReadFull(conn, readyBuf)

    // 2. Handshake
    priv, pubBytes, _ := security.GenerateDHKeys()
    conn.Write(pubBytes)

    remotePub := make([]byte, 65)
    io.ReadFull(conn, remotePub)

    // 3. Генерация ключа с использованием пароля
    chatKey, _ := security.DeriveKey(priv, remotePub, password)
    fmt.Println("Канал инициализирован.")

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if err != nil { return }

			decrypted, err := security.Decrypt(chatKey, buf[:n])
			if err != nil {
				// Если пароль неверный, AES-GCM не сможет проверить подпись пакета
				fmt.Printf("\nОшибка дешифровки")
				continue
			}
			fmt.Printf("\n>: %s\n> ", string(decrypted))
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		encrypted, _ := security.Encrypt(chatKey, scanner.Bytes())
		conn.Write(encrypted)
		fmt.Print("")
	}
}