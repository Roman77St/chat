package protocol

import (
	"errors"
	"io"
)

const (
	// Размер буфера для системных команд
	SysMsgSize = 16
	// Сигнал готовности сервера (должен быть ровно SysMsgSize байт)
	ReadySignal = "[SYSTEM]: READY "
	// Размер публичного ключа P256
	PubKeySize = 65
	RoomIDSize = 16
)

var (
	ErrProtocolMismatch = errors.New("protocol mismatch: invalid signal")
	ErrRoomFull = errors.New("room full")
)

// SendReady отправляет сигнал готовности в поток
func SendReady(w io.Writer) error {
	_, err := w.Write([]byte(ReadySignal))
	return err
}

// ReadReady ждет точного сигнала готовности
func ReadReady(r io.Reader) error {
	buf := make([]byte, SysMsgSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	if string(buf) != ReadySignal {
		return ErrProtocolMismatch
	}
	return nil
}

// SendRoomID отправляет идентификатор комнаты серверу
func SendRoomID(w io.Writer, roomID []byte) error {
	_, err := w.Write(roomID)
	return err
}

// ReadRoomID читает идентификатор комнаты от клиента
func ReadRoomID(r io.Reader) ([]byte, error) {
	buf := make([]byte, RoomIDSize)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}