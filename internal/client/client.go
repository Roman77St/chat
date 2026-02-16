package client

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/Roman77St/chat/pkg/protocol"
	"github.com/Roman77St/chat/pkg/security"
)

func Start(addr string) {
    conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("Ошибка подключения: %v\n", err)
		return
	}
    defer conn.Close()

    // Сначала вводим пароль (он нужен для генерации ключа)
    password := getPassword(os.Stdin)

    fmt.Println("Ожидание синхронизации...")

    // Сигнал READY от сервера
	if err := connectionReady(conn); err != nil {
		fmt.Printf("Ошибка синхронизации: %v\n", err)
		return
	}

	chatKey, err := genKey(conn, password)
	if err != nil {
		fmt.Printf("Ошибка получения ключа: %v\n", err)
		return
	}

    fmt.Println("Канал инициализирован.")

	startChatLoop(conn, chatKey, os.Stdout, os.Stdin)

}

func getPassword(r io.Reader) string {
	fmt.Print("Введите код доступа: ")
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
		fmt.Fprintf(w, "\npeer > %s\n> ", string(decrypted))
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