// Package config містить налаштування сервісу.
package config

import "os"

const ServiceName = "tech-store-go"

type Config struct {
	Port string
}

// Load читає конфігурацію зі змінних оточення (PORT, за замовчуванням 8080).
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return Config{Port: port}
}
