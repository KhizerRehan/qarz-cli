package cli

import (
	"fmt"

	"qarz/internal/loan"
	"qarz/internal/store"
)

func (a *App) describeStatus(row store.LoanRow) string {
	if row.Status == loan.StatusOverdue {
		return fmt.Sprintf("overdue %dd", loan.DaysOverdue(row.Loan.DueOn, a.today()))
	}
	return string(row.Status)
}
