package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/KhizerRehan/qarz-cli/internal/store"
)

func newListCommand(app *App) *cobra.Command {
	var (
		open    bool
		paid    bool
		overdue bool
		all     bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List loans",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			filter, err := chooseFilter(open, paid, overdue, all)
			if err != nil {
				return err
			}

			rows, err := store.ListLoans(cmd.Context(), app.DB, filter, app.today())
			if err != nil {
				return err
			}

			if len(rows) == 0 {
				fmt.Fprintf(app.Out, "no %s loans\n", filter)
				return nil
			}

			w := tabwriter.NewWriter(app.Out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tBORROWER\tBALANCE\tDUE\tSTATUS")
			for _, row := range rows {
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
					row.Loan.ID, row.Loan.Borrower, app.money(row.BalanceMinor),
					row.Loan.DueOn, app.describeStatus(row))
			}
			return w.Flush()
		},
	}

	cmd.Flags().BoolVar(&open, "open", false, "loans not yet repaid (default)")
	cmd.Flags().BoolVar(&paid, "paid", false, "loans repaid in full")
	cmd.Flags().BoolVar(&overdue, "overdue", false, "open loans past their due date")
	cmd.Flags().BoolVar(&all, "all", false, "every loan, whatever its state")

	return cmd
}

func chooseFilter(open, paid, overdue, all bool) (store.Filter, error) {
	chosen := []store.Filter{}
	if open {
		chosen = append(chosen, store.FilterOpen)
	}
	if paid {
		chosen = append(chosen, store.FilterPaid)
	}
	if overdue {
		chosen = append(chosen, store.FilterOverdue)
	}
	if all {
		chosen = append(chosen, store.FilterAll)
	}

	switch len(chosen) {
	case 0:
		return store.FilterOpen, nil
	case 1:
		return chosen[0], nil
	default:
		return "", fmt.Errorf("pick one of --open, --paid, --overdue, --all")
	}
}
