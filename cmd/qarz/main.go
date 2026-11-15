package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/KhizerRehan/qarz-cli/internal/cli"
	"github.com/KhizerRehan/qarz-cli/internal/config"
	"github.com/KhizerRehan/qarz-cli/internal/store"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "qarz: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	location, err := cfg.Location()
	if err != nil {
		fmt.Fprintf(os.Stderr, "qarz: %v, falling back to this machine's timezone\n", err)
		location = time.Local
	}

	path, err := config.LedgerPath()
	if err != nil {
		return err
	}

	db, err := store.Open(path)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := store.Migrate(ctx, db); err != nil {
		return err
	}

	root := cli.NewRootCommand(&cli.App{
		DB:       db,
		Now:      func() time.Time { return time.Now().In(location) },
		Currency: cfg.Currency,
		Version:  version,
		Out:      os.Stdout,
		Err:      os.Stderr,
	})

	return root.ExecuteContext(ctx)
}
