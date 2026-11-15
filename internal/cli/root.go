package cli

import (
	"database/sql"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/KhizerRehan/qarz-cli/internal/loan"
)

type App struct {
	DB       *sql.DB
	Now      func() time.Time
	Currency string
	Version  string
	Out      io.Writer
	Err      io.Writer
}

func NewRootCommand(app *App) *cobra.Command {
	root := &cobra.Command{
		Use:           "qarz",
		Short:         "A ledger for interest-free loans you have made",
		Version:       app.Version,
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
	)
	return root
}

func (a *App) today() loan.Date {
	return loan.DateOf(a.Now())
}

func (a *App) money(m loan.Minor) string {
	return m.WithCurrency(a.Currency)
}
