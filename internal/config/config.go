package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	DefaultCalendarID = "primary"
	DefaultTimezone   = "Asia/Karachi"
	DefaultCurrency   = "PKR"
)

type Config struct {
	CalendarID string `toml:"calendar_id"`
	Timezone   string `toml:"timezone"`
	Currency   string `toml:"currency"`
}

func Default() Config {
	return Config{
		CalendarID: DefaultCalendarID,
		Timezone:   DefaultTimezone,
		Currency:   DefaultCurrency,
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

func LedgerPath() (string, error)      { return pathIn("qarz.db") }
func ConfigPath() (string, error)      { return pathIn("config.toml") }
func CredentialsPath() (string, error) { return pathIn("credentials.json") }
func TokenPath() (string, error)       { return pathIn("token.json") }

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
		return Config{}, fmt.Errorf("read %s: %w (fix it, or delete it and run 'qarz setup')", path, err)
	}

	return withDefaults(cfg), nil
}

func Save(cfg Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	body := fmt.Sprintf("calendar_id = %q\ntimezone = %q\ncurrency = %q\n",
		cfg.CalendarID, cfg.Timezone, cfg.Currency)

	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func withDefaults(cfg Config) Config {
	defaults := Default()
	if cfg.CalendarID == "" {
		cfg.CalendarID = defaults.CalendarID
	}
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
