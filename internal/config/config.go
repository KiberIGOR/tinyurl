// Package config загружает параметры сервера из JSON-файла, флагов и переменных окружения.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
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

type fileConfig struct {
	ServerAddress   *string `json:"server_address"`
	BaseURL         *string `json:"base_url"`
	FileStoragePath *string `json:"file_storage_path"`
	DatabaseDSN     *string `json:"database_dsn"`
	SecretKey       *string `json:"secret_key"`
	AuditFile       *string `json:"audit_file"`
	AuditURL        *string `json:"audit_url"`
	EnableHTTPS     *bool   `json:"enable_https"`
}

// Parse читает конфигурацию: значения по умолчанию, JSON-файл, флаги, переменные окружения.
// У файла наименьший приоритет; env переопределяет флаги.
func Parse() (*Config, error) {
	return ParseFromArgs(os.Args[1:])
}

// ParseFromArgs — то же, что Parse, но с явным списком аргументов (удобно в тестах).
func ParseFromArgs(args []string) (*Config, error) {
	return parseConfig(args)
}

func parseConfig(args []string) (*Config, error) {
	cfg := defaultConfig()

	configPath := configPathFromEnvAndArgs(args)
	if configPath != "" {
		if err := loadFileConfig(configPath, cfg); err != nil {
			return nil, err
		}
	}

	fs := flag.NewFlagSet("shortener", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	bindFlags(fs, cfg, configPath)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	applyEnv(cfg)
	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Address:     ":8080",
		BaseURL:     "http://localhost:8080/",
		SecretKey:   "J83YbUemqzZBG1iJoNz8",
		// FileStoragePath: "fileMemory.txt",
		// DataBaseDSN: "postgres://urls:urls@localhost:5432/urls?sslmode=disable",
	}
}

func bindFlags(fs *flag.FlagSet, cfg *Config, configPath string) {
	var configFile = configPath
	fs.StringVar(&configFile, "c", configPath, "path to JSON config file")
	fs.StringVar(&configFile, "config", configPath, "path to JSON config file")
	fs.StringVar(&cfg.Address, "a", cfg.Address, "address and port to run server")
	fs.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "address and port for redirect url")
	fs.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file name witch will be use for storage")
	fs.StringVar(&cfg.DataBaseDSN, "d", cfg.DataBaseDSN, "dsn string for database setting")
	fs.StringVar(&cfg.SecretKey, "k", cfg.SecretKey, "Secret Key for JWT")
	fs.StringVar(&cfg.AuditFile, "audit-file", cfg.AuditFile, "path to audit log file")
	fs.StringVar(&cfg.AuditURL, "audit-url", cfg.AuditURL, "URL of remote audit receiver")
	fs.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "enable HTTPS")
}

func configPathFromEnvAndArgs(args []string) string {
	path := os.Getenv("CONFIG")
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-c", "-config":
			if i+1 < len(args) {
				path = args[i+1]
				i++
			}
		default:
			if strings.HasPrefix(arg, "-c=") {
				path = strings.TrimPrefix(arg, "-c=")
			} else if strings.HasPrefix(arg, "-config=") {
				path = strings.TrimPrefix(arg, "-config=")
			}
		}
	}
	return path
}

func loadFileConfig(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config %q: %w", path, err)
	}
	var fc fileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return fmt.Errorf("parse config %q: %w", path, err)
	}
	mergeFileConfig(cfg, &fc)
	return nil
}

func mergeFileConfig(cfg *Config, fc *fileConfig) {
	if fc.ServerAddress != nil {
		cfg.Address = *fc.ServerAddress
	}
	if fc.BaseURL != nil {
		cfg.BaseURL = *fc.BaseURL
	}
	if fc.FileStoragePath != nil {
		cfg.FileStoragePath = *fc.FileStoragePath
	}
	if fc.DatabaseDSN != nil {
		cfg.DataBaseDSN = *fc.DatabaseDSN
	}
	if fc.SecretKey != nil {
		cfg.SecretKey = *fc.SecretKey
	}
	if fc.AuditFile != nil {
		cfg.AuditFile = *fc.AuditFile
	}
	if fc.AuditURL != nil {
		cfg.AuditURL = *fc.AuditURL
	}
	if fc.EnableHTTPS != nil {
		cfg.EnableHTTPS = *fc.EnableHTTPS
	}
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("SERVER_ADDRESS"); v != "" {
		cfg.Address = v
	}
	if v := os.Getenv("BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv("FILE_STORAGE_PATH"); v != "" {
		cfg.FileStoragePath = v
	}
	if v := os.Getenv("DATABASE_DSN"); v != "" {
		cfg.DataBaseDSN = v
	}
	if v := os.Getenv("AUDIT_FILE"); v != "" {
		cfg.AuditFile = v
	}
	if v := os.Getenv("AUDIT_URL"); v != "" {
		cfg.AuditURL = v
	}
	if v := os.Getenv("ENABLE_HTTPS"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.EnableHTTPS = b
		}
	}
}
