package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	Address string
	BaseURL string
	FileStoragePath string
	DataBaseDSN string
}

func Parse() *Config {
	cfg := &Config{}
	ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
                      `localhost`, `video`, `video`, `video`)
	flag.StringVar(&cfg.Address, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "address and port for redirect url")
	flag.StringVar(&cfg.FileStoragePath, "f", "fileMemory.txt", "file name witch will be use for storage")
	flag.StringVar(&cfg.DataBaseDSN, "d", ps, "dsn string for database setting")
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
        cfg.Address = envRunAddr
  }

	if envRunAddr := os.Getenv("BASE_URL"); envRunAddr != "" {
        cfg.BaseURL = envRunAddr
  }

	if envRunAddr := os.Getenv("FILE_STORAGE_PATH"); envRunAddr != "" {
        cfg.FileStoragePath = envRunAddr
  }

	if envRunAddr := os.Getenv("DATABASE_DSN"); envRunAddr != "" {
        cfg.DataBaseDSN = envRunAddr
  }

	return cfg
}
