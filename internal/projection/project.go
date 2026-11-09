package projection

import (
	"fmt"
	"sort"
	"strings"

	"qarz/internal/loan"
	"qarz/internal/store"
)

func Desired(row store.LoanRow, today loan.Date, currency string) (Event, bool) {
	if row.Loan.Voided() {
		return Event{}, false
	}

	event := Event{
		ID:          row.Loan.CalendarEventID,
		Tag:         row.Loan.UID,
		Date:        row.Loan.DueOn,
		Description: describe(row, currency),
	}

	switch row.Status {
	case loan.StatusPaid:
		event.Title = fmt.Sprintf("[PAID] Qarz: %s — %s",
			row.Loan.Borrower, row.Loan.PrincipalMinor.WithCurrency(currency))

	case loan.StatusOverdue:
		event.Title = fmt.Sprintf("[OVERDUE %dd] Qarz: %s — %s",
			loan.DaysOverdue(row.Loan.DueOn, today),
			row.Loan.Borrower, row.BalanceMinor.WithCurrency(currency))
		event.Date = row.Loan.DueOn.Later(today)
		event.Reminders = []int{0}

	default:
		event.Title = fmt.Sprintf("Qarz: %s — %s",
			row.Loan.Borrower, row.BalanceMinor.WithCurrency(currency))
		event.Reminders = []int{RemindDayBefore, RemindShortlyBefore}
	}

	return event, true
}

func describe(row store.LoanRow, currency string) string {
	var lines []string
	if row.Loan.Note != "" {
		lines = append(lines, row.Loan.Note, "")
	}

	lines = append(lines,
		"principal: "+row.Loan.PrincipalMinor.WithCurrency(currency),
		"repaid: "+row.RepaidMinor.WithCurrency(currency),
		"balance: "+row.BalanceMinor.WithCurrency(currency),
		"lent on: "+row.Loan.LentOn.String(),
		"due on: "+row.Loan.DueOn.String(),
		fmt.Sprintf("loan: %d", row.Loan.ID),
	)

	if len(row.Loan.Metadata) > 0 {
		keys := make([]string, 0, len(row.Loan.Metadata))
		for key := range row.Loan.Metadata {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		lines = append(lines, "")
		for _, key := range keys {
			lines = append(lines, key+": "+row.Loan.Metadata[key])
		}
	}

	return strings.Join(lines, "\n")
}
