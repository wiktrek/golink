package utils

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	mysql "github.com/go-sql-driver/mysql"
)

type DatabaseConfig struct {
	Driver string
	DSN    string
}

func DatabaseConfigFromEnv() (DatabaseConfig, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return legacyDatabaseConfig()
	}
	u, err := url.Parse(databaseURL)
	if err != nil || u.Scheme == "" {
		return DatabaseConfig{}, errors.New("DATABASE_URL must be a valid database URL")
	}
	driver := strings.ToLower(u.Scheme)
	if configured := strings.ToLower(strings.TrimSpace(os.Getenv("DB_DRIVER"))); configured != "" && configured != driver {
		return DatabaseConfig{}, fmt.Errorf("DB_DRIVER %q does not match DATABASE_URL scheme %q", configured, driver)
	}
	switch driver {
	case "mysql":
		return mysqlConfigFromURL(u)
	case "sqlite":
		return sqliteConfigFromURL(u)
	default:
		return DatabaseConfig{}, fmt.Errorf("DATABASE_URL scheme must be mysql or sqlite, got %q", driver)
	}
}

func legacyDatabaseConfig() (DatabaseConfig, error) {
	driver := strings.ToLower(strings.TrimSpace(os.Getenv("DB_DRIVER")))
	if driver == "" {
		driver = "mysql"
	}
	if driver != "mysql" && driver != "sqlite" {
		return DatabaseConfig{}, fmt.Errorf("DB_DRIVER must be mysql or sqlite, got %q", driver)
	}
	dsn := strings.TrimSpace(os.Getenv("DB"))
	if dsn == "" {
		return DatabaseConfig{}, errors.New("DATABASE_URL is required (set it in the environment or .env file)")
	}
	return DatabaseConfig{Driver: driver, DSN: dsn}, nil
}

func mysqlConfigFromURL(u *url.URL) (DatabaseConfig, error) {
	if u.Host == "" || u.User == nil || strings.Trim(u.Path, "/") == "" {
		return DatabaseConfig{}, errors.New("MySQL DATABASE_URL must include user, host, and database name")
	}
	config := mysql.NewConfig()
	config.User = u.User.Username()
	config.Passwd, _ = u.User.Password()
	config.Net, config.Addr, config.DBName = "tcp", u.Host, strings.TrimPrefix(u.Path, "/")
	dsn := config.FormatDSN()
	if u.RawQuery != "" {
		dsn += "?" + u.RawQuery
	}
	return DatabaseConfig{Driver: "mysql", DSN: dsn}, nil
}

func sqliteConfigFromURL(u *url.URL) (DatabaseConfig, error) {
	path := u.Host + u.Path
	if path == "" {
		return DatabaseConfig{}, errors.New("SQLite DATABASE_URL must include a database path")
	}
	dsn := "file:" + path
	if u.RawQuery != "" {
		dsn += "?" + u.RawQuery
	}
	return DatabaseConfig{Driver: "sqlite", DSN: dsn}, nil
}

// ParseLinkDomain returns the host to route and base URL to generate.
func ParseLinkDomain(value string) (host, baseURL string, err error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", nil
	}
	if !strings.Contains(value, "://") {
		value = "https://" + value
	}
	u, err := url.Parse(value)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || (u.Path != "" && u.Path != "/") {
		return "", "", errors.New(`LINK_DOMAIN must be a hostname like "example.link" or a URL like "https://example.link"`)
	}
	return u.Hostname(), u.Scheme + "://" + u.Host, nil
}
