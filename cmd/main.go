package main

import (
	"fmt"
	"net/http"

	"github.com/Roman77St/chat/internal/chat"
	"github.com/Roman77St/chat/internal/config"
)

func main() {
	cnf := config.New()

	srv := chat.NewServer()

	http.HandleFunc(cnf.Path, func(w http.ResponseWriter, r *http.Request) {
		// 1. Пытаемся зайти
		conn, err := srv.Join(w, r)
		if err == chat.ErrRoomFull {
			srv.Alert(err.Error())
			return
		}

		// 2. Гарантируем очистку
		defer srv.Leave(conn)

		fmt.Println("Новое соединение установлено")

		// 3. Основной цикл (Message Loop)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				// Если ошибка (например, закрытие вкладки), цикл прервется и сработает defer srv.Leave
				break
			}
			srv.Broadcast(conn, msg)
		}
	})

	fmt.Println("Сервер на :8080")
	http.ListenAndServe(":8080", nil)
}