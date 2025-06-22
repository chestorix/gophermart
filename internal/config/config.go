package config

import (
	"flag"
	"os"
)

type ServerConfig struct {
	RunAddress string `env:"RUN_ADDRESS"`
	DbURL      string `env:"DATABASE_URL"`
	AccSysAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func Load() *ServerConfig {
	cfg := &ServerConfig{}
	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.DbURL, "d", "", "host=<host> user=<user> password=<password> dbname=<dbname> sslmode=<disable/enable>")
	flag.StringVar(&cfg.AccSysAddr, "r", "", "accrual system address ")
	flag.Parse()

	if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
		cfg.RunAddress = envRunAddress
	}
	if envDbURL := os.Getenv("DATABASE_URL"); envDbURL != "" {
		cfg.DbURL = envDbURL
	}
	if envAccSysAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccSysAddr != "" {
		cfg.AccSysAddr = envAccSysAddr

	}
	return cfg
}
