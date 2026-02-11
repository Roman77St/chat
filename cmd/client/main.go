package main

import (
	"github.com/Roman77St/chat/internal/client"
	"github.com/Roman77St/chat/internal/config"
)

func main() {
	cfg := config.New()
	client.Start(cfg.Address())
}
