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