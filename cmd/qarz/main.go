package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/KhizerRehan/qarz-cli/internal/cli"
	"github.com/KhizerRehan/qarz-cli/internal/config"
	"github.com/KhizerRehan/qarz-cli/internal/gauth"
	"github.com/KhizerRehan/qarz-cli/internal/projection"
	"github.com/KhizerRehan/qarz-cli/internal/store"
)

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

	projector, err := buildProjector(ctx, cfg, location, db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qarz: calendar projection is off: %v\n", err)
	}

	root := cli.NewRootCommand(&cli.App{
		DB:        db,
		Projector: projector,
		Now:       func() time.Time { return time.Now().In(location) },
		Currency:  cfg.Currency,
		In:        os.Stdin,
		Out:       os.Stdout,
		Err:       os.Stderr,
	})

	return root.ExecuteContext(ctx)
}

func buildProjector(ctx context.Context, cfg config.Config, location *time.Location, db *sql.DB) (*projection.Projector, error) {
	credentialsPath, err := config.CredentialsPath()
	if err != nil {
		return nil, err
	}
	tokenPath, err := config.TokenPath()
	if err != nil {
		return nil, err
	}

	if !gauth.HasToken(tokenPath) {
		return nil, nil
	}

	client, err := gauth.HTTPClient(ctx, credentialsPath, tokenPath)
	if err != nil {
		return nil, err
	}

	calendar, err := projection.NewGoogle(ctx, client, cfg.CalendarID, location)
	if err != nil {
		return nil, err
	}

	return &projection.Projector{
		DB:       db,
		Calendar: calendar,
		Now:      func() time.Time { return time.Now().In(location) },
		Currency: cfg.Currency,
	}, nil
}
