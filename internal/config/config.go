package config

import (
	"flag"
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
	flag.StringVar(&cfg.Address, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "address and port for redirect url")
	//fileMemory.txt
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file name witch will be use for storage")
	// ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",`localhost`, `urls`, `urls`, `urls`)
	flag.StringVar(&cfg.DataBaseDSN, "d", "", "dsn string for database setting")
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
