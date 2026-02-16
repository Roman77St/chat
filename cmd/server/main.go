package main

import (
	"github.com/Roman77St/chat/internal/config"
	"github.com/Roman77St/chat/internal/server"
)

func main() {
	cfg := config.New()

	server.Run(cfg)
}