package cli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/KhizerRehan/qarz-cli/internal/store"
)

func newShowCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "show <loan>",
		Short: "Show one loan and its full repayment history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseLoanNumber(args[0])
			if err != nil {
				return err
			}

			row, err := store.GetLoan(cmd.Context(), app.DB, id, app.today())
			if err != nil {
				return err
			}

			repayments, err := store.ListRepayments(cmd.Context(), app.DB, id)
			if err != nil {
				return err
			}

			fmt.Fprintf(app.Out, "loan %d\n", row.Loan.ID)

			w := tabwriter.NewWriter(app.Out, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "  borrower\t%s\n", row.Loan.Borrower)
			fmt.Fprintf(w, "  principal\t%s\n", app.money(row.Loan.PrincipalMinor))
			fmt.Fprintf(w, "  repaid\t%s\n", app.money(row.RepaidMinor))
			fmt.Fprintf(w, "  balance\t%s\n", app.money(row.BalanceMinor))
			fmt.Fprintf(w, "  status\t%s\n", app.describeStatus(row))
			fmt.Fprintf(w, "  lent on\t%s\n", row.Loan.LentOn)
			fmt.Fprintf(w, "  due on\t%s\n", row.Loan.DueOn)
			if row.Loan.Voided() {
				fmt.Fprintf(w, "  voided at\t%s\n", row.Loan.VoidedAt)
			}
			if row.Loan.Note != "" {
				fmt.Fprintf(w, "  note\t%s\n", row.Loan.Note)
			}
			if len(row.Loan.Metadata) > 0 {
				fmt.Fprintf(w, "  details\t%s\n", formatDetails(row.Loan.Metadata))
			}
			if err := w.Flush(); err != nil {
				return err
			}

			if len(repayments) == 0 {
				fmt.Fprintln(app.Out, "  repayments: none yet")
				return nil
			}

			fmt.Fprintln(app.Out, "  repayments")
			rw := tabwriter.NewWriter(app.Out, 0, 0, 2, ' ', 0)
			for _, r := range repayments {
				fmt.Fprintf(rw, "    %s\t%s\t%s\n", r.PaidOn, app.money(r.AmountMinor), r.Note)
			}
			return rw.Flush()
		},
	}
}

func formatDetails(details map[string]string) string {
	keys := make([]string, 0, len(details))
	for key := range details {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, key+"="+details[key])
	}
	return strings.Join(pairs, " ")
}

func parseLoanNumber(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%q is not a loan number", raw)
	}
	return id, nil
}
