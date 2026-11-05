package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	Address         string
	BaseURL         string
	FileStoragePath string
	DataBaseDSN     string
	SecretKey       string
	AuditFile       string
	AuditURL        string
}

func Parse() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.Address, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "address and port for redirect url")
	//fileMemory.txt
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file name witch will be use for storage")
	// ps := fmt.Sprintf("postgres://%s:%s@localhost:5432/urls?sslmode=disable","urls","urls")
	flag.StringVar(&cfg.DataBaseDSN, "d", "", "dsn string for database setting")
	flag.StringVar(&cfg.SecretKey, "k", "J83YbUemqzZBG1iJoNz8", "Secret Key for JWT")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "path to audit log file")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "URL of remote audit receiver")
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

	if envRunAddr := os.Getenv("AUDIT_FILE"); envRunAddr != "" {
		cfg.AuditFile = envRunAddr
	}

	if envRunAddr := os.Getenv("AUDIT_URL"); envRunAddr != "" {
		cfg.AuditURL = envRunAddr
	}

	return cfg
}
