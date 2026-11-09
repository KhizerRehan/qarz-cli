package cli

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"qarz/internal/projection"
	"qarz/internal/loan"
)

type App struct {
	DB        *sql.DB
	Projector *projection.Projector
	Now       func() time.Time
	Currency  string
	In        io.Reader
	Out       io.Writer
	Err       io.Writer
}

func NewRootCommand(app *App) *cobra.Command {
	root := &cobra.Command{
		Use:           "qarz",
		Short:         "A ledger for interest-free loans, projected onto your calendar",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(app.Out)
	root.SetErr(app.Err)
	root.AddCommand(
		newAddCommand(app),
		newListCommand(app),
		newRepayCommand(app),
		newShowCommand(app),
		newSetupCommand(app),
		newSyncCommand(app),
	)
	return root
}

func (a *App) today() loan.Date {
	return loan.DateOf(a.Now())
}

func (a *App) money(m loan.Minor) string {
	return m.WithCurrency(a.Currency)
}

func (a *App) project(ctx context.Context, loanID int64) error {
	if a.Projector == nil {
		fmt.Fprintln(a.Err, "warning: not signed in to Google, so no calendar event was made (run 'qarz setup')")
		return nil
	}

	err := a.Projector.Project(ctx, loanID)
	if errors.Is(err, projection.ErrPending) {
		fmt.Fprintf(a.Err, "warning: recorded, but the calendar event is pending (run 'qarz sync'): %v\n", err)
		return nil
	}
	return err
}
