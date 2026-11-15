package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	DefaultTimezone = "Asia/Karachi"
	DefaultCurrency = "PKR"
)

type Config struct {
	Timezone string `toml:"timezone"`
	Currency string `toml:"currency"`
}

func Default() Config {
	return Config{
		Timezone: DefaultTimezone,
		Currency: DefaultCurrency,
	}
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home directory: %w", err)
	}

	dir := filepath.Join(home, ".qarz")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	return dir, nil
}

func LedgerPath() (string, error) { return pathIn("qarz.db") }
func ConfigPath() (string, error) { return pathIn("config.toml") }

func pathIn(name string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

func Load() (Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}

	cfg := Default()
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}

	if err := toml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("read %s: %w (fix it, or delete it to use the defaults)", path, err)
	}

	return withDefaults(cfg), nil
}

func withDefaults(cfg Config) Config {
	defaults := Default()
	if cfg.Timezone == "" {
		cfg.Timezone = defaults.Timezone
	}
	if cfg.Currency == "" {
		cfg.Currency = defaults.Currency
	}
	return cfg
}

func (c Config) Location() (*time.Location, error) {
	loc, err := time.LoadLocation(c.Timezone)
	if err != nil {
		return nil, fmt.Errorf("unknown timezone %q in config: %w", c.Timezone, err)
	}
	return loc, nil
}
