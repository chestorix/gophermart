package config

import (
	"flag"
	"os"
	"strings"
)

type ServerConfig struct {
	RunAddress string `env:"RUN_ADDRESS"`
	DBURI      string `env:"DATABASE_URI"`
	AccSysAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func ensureHTTP(address string) string {
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		return "http://" + address
	}
	return address
}
func Load() *ServerConfig {
	cfg := &ServerConfig{}
	flag.StringVar(&cfg.RunAddress, "a", "localhost:8090", "address and port to run server")
	flag.StringVar(&cfg.DBURI, "d", "", "host=<host> user=<user> password=<password> dbname=<dbname> sslmode=<disable/enable>")
	flag.StringVar(&cfg.AccSysAddr, "r", "", "accrual system address ")
	flag.Parse()

	if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
		cfg.RunAddress = envRunAddress
	}
	if envDBURI := os.Getenv("DATABASE_URI"); envDBURI != "" {
		cfg.DBURI = envDBURI
	}
	if envAccSysAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccSysAddr != "" {
		cfg.AccSysAddr = ensureHTTP(envAccSysAddr)

	}
	return cfg
}
