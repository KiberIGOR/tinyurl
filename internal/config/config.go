package config

import (
	"flag"
	"os"
)

type Config struct {
	Address string
	BaseURL string
}

func Parse() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.Address, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "address and port for redirect url")
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
        cfg.Address = envRunAddr
  }
	
	if envRunAddr := os.Getenv("BASE_URL"); envRunAddr != "" {
        cfg.BaseURL = envRunAddr
  }

	return cfg
}
