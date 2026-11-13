// Package config загружает параметры сервера из флагов и переменных окружения.
package config

import (
	"flag"
	"os"
	"strconv"
)

// Config — параметры запуска HTTP-сервера и хранилища.
type Config struct {
	// Address — адрес прослушивания (флаг -a, env SERVER_ADDRESS).
	Address         string
	BaseURL         string
	FileStoragePath string
	DataBaseDSN     string
	SecretKey       string
	AuditFile       string
	AuditURL        string
	// EnableHTTPS включает TLS (флаг -s, env ENABLE_HTTPS).
	EnableHTTPS bool
}

// Parse читает флаги командной строки и переопределяет их значениями из env.
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
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "enable HTTPS")
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

	if envHTTPS := os.Getenv("ENABLE_HTTPS"); envHTTPS != "" {
		if v, err := strconv.ParseBool(envHTTPS); err == nil {
			cfg.EnableHTTPS = v
		}
	}

	return cfg
}
