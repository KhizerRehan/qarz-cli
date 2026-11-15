package loan

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
)

type Status string

const (
	StatusOpen    Status = "open"
	StatusPaid    Status = "paid"
	StatusOverdue Status = "overdue"
	StatusVoid    Status = "void"
)

type Loan struct {
	ID             int64
	UID            string
	Borrower       string
	PrincipalMinor Minor
	LentOn         Date
	DueOn          Date
	Note           string
	Metadata       map[string]string
	VoidedAt       string
}

type Repayment struct {
	ID          int64
	LoanID      int64
	AmountMinor Minor
	PaidOn      Date
	Note        string
}

func (l Loan) Voided() bool {
	return l.VoidedAt != ""
}

func NewUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate loan tag: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}

func Balance(principal, repaid Minor) Minor {
	return principal - repaid
}

func DeriveStatus(voided bool, balance Minor, dueOn, today Date) Status {
	switch {
	case voided:
		return StatusVoid
	case balance == 0:
		return StatusPaid
	case dueOn.Before(today):
		return StatusOverdue
	default:
		return StatusOpen
	}
}

func DaysOverdue(dueOn, today Date) int {
	days := today.DaysSince(dueOn)
	if days < 0 {
		return 0
	}
	return days
}

var (
	ErrNonPositiveAmount = errors.New("amount must be positive")
	ErrDueBeforeLent     = errors.New("due date is before the date of lending")
	ErrOverpayment       = errors.New("repayment exceeds the outstanding balance")
)

func ValidateNewLoan(principal Minor, lentOn, dueOn Date) error {
	if principal <= 0 {
		return ErrNonPositiveAmount
	}
	if dueOn.Before(lentOn) {
		return fmt.Errorf("%w: due %s, lent on %s", ErrDueBeforeLent, dueOn, lentOn)
	}
	return nil
}

func ValidateRepayment(amount, balance Minor) error {
	if amount <= 0 {
		return ErrNonPositiveAmount
	}
	if amount > balance {
		return fmt.Errorf("%w: tried to record %s against an outstanding balance of %s",
			ErrOverpayment, amount, balance)
	}
	return nil
}
