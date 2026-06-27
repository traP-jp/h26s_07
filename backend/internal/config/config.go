package config

import (
	"net"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Config struct {
	Port             string
	CORSAllowOrigins []string
	Database         DatabaseConfig
}

type DatabaseConfig struct {
	URL      string
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func Load() Config {
	return Config{
		Port:             env("PORT", "8080"),
		CORSAllowOrigins: envList("CORS_ALLOW_ORIGINS", []string{"http://localhost:5173", "http://127.0.0.1:5173"}),
		Database: DatabaseConfig{
			URL:      env("DATABASE_URL", ""),
			Host:     env("DB_HOST", "127.0.0.1"),
			Port:     env("DB_PORT", "3306"),
			User:     env("DB_USER", env("MARIADB_USER", "")),
			Password: env("DB_PASSWORD", env("MARIADB_PASSWORD", "")),
			Name:     env("DB_NAME", env("MARIADB_DATABASE", "")),
		},
	}
}

func (c Config) Addr() string {
	return ":" + c.Port
}

func (c DatabaseConfig) DSN() string {
	if strings.TrimSpace(c.URL) != "" {
		return strings.TrimSpace(c.URL)
	}
	if c.User == "" || c.Name == "" {
		return ""
	}

	cfg := mysql.Config{
		User:                 c.User,
		Passwd:               c.Password,
		Net:                  "tcp",
		Addr:                 net.JoinHostPort(c.Host, c.Port),
		DBName:               c.Name,
		ParseTime:            true,
		Loc:                  time.UTC,
		AllowNativePasswords: true,
		Params: map[string]string{
			"charset": "utf8mb4",
		},
	}
	return cfg.FormatDSN()
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envList(key string, fallback []string) []string {
	rawValue := strings.TrimSpace(os.Getenv(key))
	if rawValue == "" {
		return fallback
	}

	values := make([]string, 0)
	for item := range strings.SplitSeq(rawValue, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			values = append(values, item)
		}
	}
	return values
}
