package client

import (
	"bytes"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/Roman77St/chat/pkg/protocol"
	"github.com/Roman77St/chat/pkg/security"
)

func TestGetPassword(t *testing.T) {
	// Имитируем ввод пользователя: пишем пароль и символ переноса строки
	input := "my_secret_password\n"
	reader := strings.NewReader(input)

	password := getPassword(reader, input)

	if password != "my_secret_password" {
		t.Errorf("Ожидалось 'my_secret_password', получено '%s'", password)
	}
}

func TestConnectionReady_Success(t *testing.T) {
	// Создаем виртуальное соединение
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	// Имитируем действия сервера в отдельной горутине
	go func() {
		// Сервер отправляет правильный сигнал
		serverConn.Write([]byte(protocol.ReadySignal))
	}()

	// Клиент пытается прочитать сигнал
	err := connectionReady(clientConn)

	if err != nil {
		t.Errorf("Ожидался успех, получена ошибка: %v", err)
	}
}

func TestConnectionReady_WrongSignal(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	go func() {
		// Сервер отправляет неправильный сигнал (меньше байт или другой текст)
		serverConn.Write([]byte("WRONG_SIGNAL____"))
	}()

	err := connectionReady(clientConn)

	if err == nil {
		t.Error("Ожидалась ошибка из-за неверного сигнала, но ошибка nil")
	}
}

func TestGenKey(t *testing.T) {
	// Создаем виртуальное соединение
	clientConn, peerConn := net.Pipe()
	defer clientConn.Close()
	defer peerConn.Close()

	password := "test_password"

	// Канал для передачи ключа, который сгенерирует "собеседник"
	peerKeyChan := make(chan []byte)

	// Имитируем второго участника (Peer)
	go func() {
		// 1. Генерируем ключи для "собеседника"
		privPeer, pubPeer, err := security.GenerateDHKeys()
		if err != nil {
			t.Errorf("Ошибка генерации ключа: %v", err)
			return
		}

		// 2. Читаем публичный ключ от нашего клиента (из genKey)
		remotePubFromClient := make([]byte, protocol.PubKeySize)
		io.ReadFull(peerConn, remotePubFromClient)

		// 3. Отправляем публичный ключ "собеседника" нашему клиенту
		peerConn.Write(pubPeer)

		// 4. Генерируем общий секрет на стороне "собеседника"
		chatKeyPeer, _ := security.DeriveKey(privPeer, remotePubFromClient, password)
		peerKeyChan <- chatKeyPeer
	}()

	// Вызываем тестируемую функцию genKey
	chatKeyClient, err := genKey(clientConn, password)
	if err != nil {
		t.Fatalf("genKey вернул ошибку: %v", err)
	}

	// Получаем ключ собеседника из горутины
	chatKeyPeer := <-peerKeyChan

	// Сравниваем ключи: они должны быть идентичны
	if !bytes.Equal(chatKeyClient, chatKeyPeer) {
		t.Error("Ключи клиента и собеседника не совпадают!")
	}
}

func TestReceiveAndDecrypt(t *testing.T) {
	clientConn, peerConn := net.Pipe()
	var out bytes.Buffer
	// Используем фиксированный ключ для теста
	testKey := make([]byte, 32)
	copy(testKey, "a_very_secret_32_byte_key_12345")

	messageText := "Привет, это тестовое сообщение!"

	go func() {
		defer peerConn.Close() // Закрытие соединения прервет цикл Read в функции

		// 1. Шифруем сообщение, как будто его отправил собеседник
		encrypted, err := security.Encrypt(testKey, []byte(messageText))
		if err != nil {
			t.Errorf("Ошибка шифрования в тесте: %v", err)
			return
		}

		// 2. Отправляем зашифрованные байты клиенту
		peerConn.Write(encrypted)
	}()

	// receiveAndDecrypt пишет в stdout (fmt.Printf).
	// В Unit-тестах мы обычно проверяем, что функция не падает
	// и корректно обрабатывает данные до закрытия сокета.

	receiveAndDecrypt(&out, clientConn, testKey)

	if !strings.Contains(out.String(), messageText) {
        t.Errorf("Ожидалось получение сообщения %s, получено: %s", messageText, out.String())
    }
}

func TestReadAndEncrypt(t *testing.T) {
	clientConn, peerConn := net.Pipe()
	defer clientConn.Close()
	defer peerConn.Close()

	testKey := make([]byte, 32)
	copy(testKey, "fixed_key_for_testing_purposes_1")

	messageText := "Hello, Server!"
	// Имитируем ввод пользователя (сообщение + перенос строки для Scanner)
	input := strings.NewReader(messageText + "\n")

	// Канал для синхронизации: сигнализирует, что проверка завершена
	done := make(chan bool)

	// В горутине имитируем сервер, который принимает данные
	go func() {
		buf := make([]byte, 4096)
		n, err := peerConn.Read(buf)
		if err != nil {
			t.Errorf("Ошибка чтения из Pipe: %v", err)
		}

		// Расшифровываем то, что прислал клиент
		decrypted, err := security.Decrypt(testKey, buf[:n])
		if err != nil {
			t.Errorf("Не удалось расшифровать отправленное сообщение: %v", err)
		}

		if string(decrypted) != messageText {
			t.Errorf("Ожидалось '%s', получено '%s'", messageText, string(decrypted))
		}
		done <- true
	}()

	// Запускаем функцию (она завершится, когда strings.Reader дойдет до конца)
	readAndEncrypt(input, clientConn, testKey)

	// Ждем подтверждения из горутины
	<-done
}

func TestStartChatLoop(t *testing.T) {
    clientConn, serverConn := net.Pipe()
    defer clientConn.Close()

    key := make([]byte, 32)
    userInput := "Привет из теста!\n"
    input := strings.NewReader(userInput)
    var output bytes.Buffer

    // Канал, чтобы дождаться полного завершения цикла чтения клиента
    readerFinished := make(chan struct{})

    go func() {
        // Мы запускаем чтение и передаем канал, чтобы знать, когда оно реально кончилось
        receiveAndDecrypt(&output, clientConn, key)
        close(readerFinished)
    }()

    go func() {
        buf := make([]byte, 1024)
        _, _ = serverConn.Read(buf)

        msg, _ := security.Encrypt(key, []byte("Ответ сервера"))
        serverConn.Write(msg)

        // Ключевой момент: закрываем серверный конец.
        // Это вызовет ошибку в clientConn.Read внутри receiveAndDecrypt.
        serverConn.Close()
    }()

    // Отправляем сообщение из основного потока
    readAndEncrypt(input, clientConn, key)

    // Ждем, пока горутина receiveAndDecrypt реально выйдет из цикла
    <-readerFinished

    if !strings.Contains(output.String(), "Ответ сервера") {
        t.Errorf("Клиент не получил ответ. Содержимое: %q", output.String())
    }
}