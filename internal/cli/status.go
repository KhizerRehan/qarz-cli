package cli

import (
	"fmt"

	"github.com/KhizerRehan/qarz-cli/internal/loan"
	"github.com/KhizerRehan/qarz-cli/internal/store"
)

func (a *App) describeStatus(row store.LoanRow) string {
	if row.Status == loan.StatusOverdue {
		return fmt.Sprintf("overdue %dd", loan.DaysOverdue(row.Loan.DueOn, a.today()))
	}
	return string(row.Status)
}
