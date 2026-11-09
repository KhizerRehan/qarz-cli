package loan_test

import (
	"errors"
	"testing"

	"qarz/internal/loan"
)

func mustDate(t *testing.T, s string) loan.Date {
	t.Helper()
	d, err := loan.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

func TestParseAmount(t *testing.T) {
	cases := []struct {
		in      string
		want    loan.Minor
		wantErr bool
	}{
		{in: "50000", want: 5_000_000},
		{in: "50,000", want: 5_000_000},
		{in: "0.50", want: 50},
		{in: "1.5", want: 150},
		{in: "1.05", want: 105},
		{in: "0", wantErr: true},
		{in: "-1", wantErr: true},
		{in: "1.005", wantErr: true},
		{in: "abc", wantErr: true},
		{in: "", wantErr: true},
	}

	for _, c := range cases {
		got, err := loan.ParseAmount(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("ParseAmount(%q) = %d, want an error", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseAmount(%q): %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseAmount(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestMinorString(t *testing.T) {
	cases := []struct {
		in   loan.Minor
		want string
	}{
		{in: 5_000_000, want: "50,000"},
		{in: 100_000_000, want: "1,000,000"},
		{in: 100_000, want: "1,000"},
		{in: 99_999, want: "999.99"},
		{in: 50, want: "0.50"},
		{in: 0, want: "0"},
	}

	for _, c := range cases {
		if got := c.in.String(); got != c.want {
			t.Errorf("Minor(%d).String() = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDeriveStatus(t *testing.T) {
	today := mustDate(t, "2026-12-15")

	cases := []struct {
		name    string
		voided  bool
		balance loan.Minor
		dueOn   string
		want    loan.Status
	}{
		{name: "not yet due", balance: 5_000_000, dueOn: "2026-12-20", want: loan.StatusOpen},
		{name: "due today", balance: 5_000_000, dueOn: "2026-12-15", want: loan.StatusOpen},
		{name: "one day past due", balance: 5_000_000, dueOn: "2026-12-14", want: loan.StatusOverdue},
		{name: "fully repaid", balance: 0, dueOn: "2026-12-14", want: loan.StatusPaid},
		{name: "voided beats everything", voided: true, balance: 5_000_000, dueOn: "2026-12-14", want: loan.StatusVoid},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := loan.DeriveStatus(c.voided, c.balance, mustDate(t, c.dueOn), today)
			if got != c.want {
				t.Errorf("DeriveStatus = %q, want %q", got, c.want)
			}
		})
	}
}

func TestDaysOverdue(t *testing.T) {
	today := mustDate(t, "2026-12-15")

	cases := []struct {
		dueOn string
		want  int
	}{
		{dueOn: "2026-12-15", want: 0},
		{dueOn: "2026-12-14", want: 1},
		{dueOn: "2026-12-03", want: 12},
		{dueOn: "2026-12-20", want: 0},
	}

	for _, c := range cases {
		if got := loan.DaysOverdue(mustDate(t, c.dueOn), today); got != c.want {
			t.Errorf("DaysOverdue(due %s) = %d, want %d", c.dueOn, got, c.want)
		}
	}
}

func TestValidateRepayment(t *testing.T) {
	const balance loan.Minor = 3_000_000

	cases := []struct {
		name    string
		amount  loan.Minor
		wantErr error
	}{
		{name: "well under the balance", amount: 1_000_000},
		{name: "exactly the balance", amount: 3_000_000},
		{name: "one paisa over the balance", amount: 3_000_001, wantErr: loan.ErrOverpayment},
		{name: "zero", amount: 0, wantErr: loan.ErrNonPositiveAmount},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := loan.ValidateRepayment(c.amount, balance)
			if c.wantErr == nil {
				if err != nil {
					t.Fatalf("ValidateRepayment(%d, %d): %v", c.amount, balance, err)
				}
				return
			}
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("ValidateRepayment(%d, %d) = %v, want %v", c.amount, balance, err, c.wantErr)
			}
		})
	}
}

func TestValidateNewLoan(t *testing.T) {
	lentOn := mustDate(t, "2026-11-10")

	if err := loan.ValidateNewLoan(5_000_000, lentOn, mustDate(t, "2026-11-10")); err != nil {
		t.Errorf("a loan due on the day it was lent should be valid: %v", err)
	}
	if err := loan.ValidateNewLoan(5_000_000, lentOn, mustDate(t, "2026-11-09")); !errors.Is(err, loan.ErrDueBeforeLent) {
		t.Errorf("due before lent = %v, want ErrDueBeforeLent", err)
	}
	if err := loan.ValidateNewLoan(0, lentOn, mustDate(t, "2026-12-15")); !errors.Is(err, loan.ErrNonPositiveAmount) {
		t.Errorf("zero principal = %v, want ErrNonPositiveAmount", err)
	}
}

func TestBalance(t *testing.T) {
	principal := loan.Minor(5_000_000)

	steps := []struct {
		repaid loan.Minor
		want   loan.Minor
	}{
		{repaid: 0, want: 5_000_000},
		{repaid: 2_000_000, want: 3_000_000},
		{repaid: 5_000_000, want: 0},
	}

	for _, step := range steps {
		if got := loan.Balance(principal, step.repaid); got != step.want {
			t.Errorf("Balance(%d, %d) = %d, want %d", principal, step.repaid, got, step.want)
		}
	}
}
