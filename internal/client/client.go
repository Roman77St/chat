package client

import (
	"bufio"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/Roman77St/chat/pkg/protocol"
	"github.com/Roman77St/chat/pkg/security"
)

func Start() {
	addr := getString(os.Stdin, "Введите адрес сервера (например, localhost:8080):")
	conf := &tls.Config{
    	InsecureSkipVerify: true, // ОБЯЗАТЕЛЬНО для самоподписанных сертификатов в RAM
	}
    conn, err := tls.Dial("tcp", addr, conf)
	if err != nil {
		fmt.Printf("Ошибка подключения: %v\n", err)
		return
	}
    defer conn.Close()

    roomPass := getString(os.Stdin, "Введите ID комнаты (пароль): ")

	hash := sha256.Sum256([]byte(roomPass))
    roomID := hash[:16]

	if err := protocol.SendRoomID(conn, roomID); err != nil {
        fmt.Println("Ошибка входа в комнату:", err)
		return
    }

    // Сначала вводим пароль (он нужен для генерации ключа)

    fmt.Println("Ожидание синхронизации...")

    // Сигнал READY от сервера
	if err := connectionReady(conn); err != nil {
		fmt.Printf("Ошибка синхронизации: %v\n", err)
		return
	}

	password := getString(os.Stdin, "Введите код доступа: ")

	chatKey, err := genKey(conn, password)
	if err != nil {
		fmt.Printf("Ошибка получения ключа: %v\n", err)
		return
	}

    fmt.Print("Канал инициализирован.\n>")

	startChatLoop(conn, chatKey, os.Stdout, os.Stdin)

}

func getString(r io.Reader, text string) string {
	fmt.Print(text)
    var password string
    fmt.Fscanln(r, &password)
	return password
}

func connectionReady(conn net.Conn) error {
	fmt.Println("Ожидание собеседника...")
	return protocol.ReadReady(conn)
}

func genKey(conn net.Conn, password string) ([]byte, error) {
	    // 2. Handshake
    priv, pubBytes, err := security.GenerateDHKeys()
	if err != nil {
		fmt.Printf("Ошибка генерации ключа: %v\n", err)
		return nil, err
	}
    conn.Write(pubBytes)

	remotePub := make([]byte, protocol.PubKeySize)
    io.ReadFull(conn, remotePub)
	chatKey, err := security.DeriveKey(priv, remotePub, password)
	if err != nil {
		fmt.Printf("Ошибка генерации ключа: %v\n", err)
		return nil, err
	}
	return chatKey, nil
}

func receiveAndDecrypt(w io.Writer, conn net.Conn, key []byte) {
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil { return }

		decrypted, err := security.Decrypt(key, buf[:n])
		if err != nil {
			// Если пароль неверный, AES-GCM не сможет проверить подпись пакета
			fmt.Printf("\nОшибка дешифровки\n> ")
			continue
		}
		fmt.Fprintf(w, "\npeer: > %s\n> ", string(decrypted))
	}
}

func readAndEncrypt(r io.Reader, conn net.Conn, key []byte) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		encrypted, err := security.Encrypt(key, scanner.Bytes())
		if err != nil {
			fmt.Printf("\nОшибка шифровки")
			continue
		}
		conn.Write(encrypted)
		fmt.Print("> ")
	}
}

func startChatLoop (conn net.Conn, key []byte, w io.Writer, r io.Reader)  {

	go receiveAndDecrypt(w, conn, key)

	readAndEncrypt(r, conn, key)

}