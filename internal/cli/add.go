package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"qarz/internal/loan"
	"qarz/internal/store"
)

func newAddCommand(app *App) *cobra.Command {
	var (
		borrower string
		amount   string
		due      string
		lentOn   string
		note     string
		details  []string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Record a loan you made",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			principal, err := loan.ParseAmount(amount)
			if err != nil {
				return err
			}

			dueOn, err := loan.ParseDate(due)
			if err != nil {
				return err
			}

			lent := app.today()
			if lentOn != "" {
				if lent, err = loan.ParseDate(lentOn); err != nil {
					return err
				}
			}

			borrower = strings.TrimSpace(borrower)
			if borrower == "" {
				return fmt.Errorf("a borrower is required")
			}

			if err := loan.ValidateNewLoan(principal, lent, dueOn); err != nil {
				return err
			}

			metadata, err := parseDetails(details)
			if err != nil {
				return err
			}

			uid, err := loan.NewUID()
			if err != nil {
				return err
			}

			id, err := store.InsertLoan(cmd.Context(), app.DB, loan.Loan{
				UID:            uid,
				Borrower:       borrower,
				PrincipalMinor: principal,
				LentOn:         lent,
				DueOn:          dueOn,
				Note:           note,
				Metadata:       metadata,
				SyncState:      loan.SyncPendingCreate,
			}, app.Now())
			if err != nil {
				return err
			}

			fmt.Fprintf(app.Out, "loan %d recorded: %s owes %s by %s\n",
				id, borrower, app.money(principal), dueOn)

			return app.project(cmd.Context(), id)
		},
	}

	cmd.Flags().StringVar(&borrower, "to", "", "who borrowed the money")
	cmd.Flags().StringVar(&amount, "amount", "", "amount lent, in whole rupees")
	cmd.Flags().StringVar(&due, "due", "", "date the money is due back (YYYY-MM-DD)")
	cmd.Flags().StringVar(&lentOn, "lent-on", "", "date the money was handed over (default today)")
	cmd.Flags().StringVar(&note, "note", "", "what the money was for")
	cmd.Flags().StringArrayVar(&details, "meta", nil, "extra detail as key=value, repeatable")

	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("amount")
	_ = cmd.MarkFlagRequired("due")

	return cmd
}

func parseDetails(pairs []string) (map[string]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	out := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		key, value, found := strings.Cut(pair, "=")
		key = strings.TrimSpace(key)
		if !found || key == "" {
			return nil, fmt.Errorf("detail %q is not in key=value form", pair)
		}
		out[key] = strings.TrimSpace(value)
	}
	return out, nil
}
