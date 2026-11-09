package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"qarz/internal/loan"
	"qarz/internal/store"
)

func newRepayCommand(app *App) *cobra.Command {
	var (
		amount string
		paidOn string
		note   string
	)

	cmd := &cobra.Command{
		Use:   "repay <loan>",
		Short: "Record money a borrower paid back",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseLoanNumber(args[0])
			if err != nil {
				return err
			}

			paid, err := loan.ParseAmount(amount)
			if err != nil {
				return err
			}

			on := app.today()
			if paidOn != "" {
				if on, err = loan.ParseDate(paidOn); err != nil {
					return err
				}
			}

			row, err := store.GetLoan(cmd.Context(), app.DB, id, app.today())
			if err != nil {
				return err
			}
			if row.Loan.Voided() {
				return fmt.Errorf("loan %d is void; restore it before recording a repayment", id)
			}
			if err := loan.ValidateRepayment(paid, row.BalanceMinor); err != nil {
				return err
			}

			if _, err := store.InsertRepayment(cmd.Context(), app.DB, loan.Repayment{
				LoanID:      id,
				AmountMinor: paid,
				PaidOn:      on,
				Note:        note,
			}); err != nil {
				return err
			}

			remaining := loan.Balance(row.Loan.PrincipalMinor, row.RepaidMinor+paid)
			if remaining == 0 {
				fmt.Fprintf(app.Out, "loan %d paid in full: %s repaid %s\n",
					id, row.Loan.Borrower, app.money(row.Loan.PrincipalMinor))
			} else {
				fmt.Fprintf(app.Out, "loan %d: %s repaid %s, %s still owed\n",
					id, row.Loan.Borrower, app.money(paid), app.money(remaining))
			}

			return app.project(cmd.Context(), id)
		},
	}

	cmd.Flags().StringVar(&amount, "amount", "", "amount repaid, in whole rupees")
	cmd.Flags().StringVar(&paidOn, "on", "", "date the money came back (default today)")
	cmd.Flags().StringVar(&note, "note", "", "anything worth recording about this repayment")

	_ = cmd.MarkFlagRequired("amount")

	return cmd
}
