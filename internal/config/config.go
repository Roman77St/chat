package config

import "time"

type Config struct {
	Host string
	Port string
	Path string
	RoomTimeout time.Duration
}

func New() *Config {
	return &Config{
		Host: "127.0.0.1", // По умолчанию только локальный хост (безопасно)
		Port: "8080",
		Path: "/ws",
		RoomTimeout: time.Second*30,
	}
}

func (c *Config) Address() string {
	return c.Host + ":" + c.Port
}