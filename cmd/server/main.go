package main

import (
	"crypto/tls"
	"fmt"

	"github.com/Roman77St/chat/internal/config"
	"github.com/Roman77St/chat/internal/network"
	"github.com/Roman77St/chat/internal/security"
)

func main() {
	cnf := config.New()
	srv := network.NewTCPServer()

	cert, _ := security.GenerateInMemoryCert()

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	ln, err := tls.Listen("tcp", cnf.Address(), tlsConfig)

	if err != nil {
		fmt.Printf("Не удалось запустить: %v\n", err)
		return
	}

	fmt.Printf("TCP server started on %s\n", cnf.Address())

	srv.Run(ln)
}
