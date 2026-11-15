package cli_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/KhizerRehan/qarz-cli/internal/cli"
	"github.com/KhizerRehan/qarz-cli/internal/loan"
	"github.com/KhizerRehan/qarz-cli/internal/store"
)

var update = flag.Bool("update", false, "rewrite the golden files from the current output")

const fixedToday = "2026-12-15"

type harness struct {
	t      *testing.T
	db     *sql.DB
	out    *bytes.Buffer
	errOut *bytes.Buffer
	now    time.Time
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	path := filepath.Join(t.TempDir(), "qarz.db")
	db, err := store.Open(path)
	if err != nil {
		t.Fatalf("open ledger: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := store.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	now, err := time.Parse(loan.DateLayout, fixedToday)
	if err != nil {
		t.Fatalf("parse fixed date: %v", err)
	}

	return &harness{
		t:      t,
		db:     db,
		out:    &bytes.Buffer{},
		errOut: &bytes.Buffer{},
		now:    now,
	}
}

func (h *harness) run(args ...string) (string, error) {
	h.t.Helper()
	h.out.Reset()
	h.errOut.Reset()

	root := cli.NewRootCommand(&cli.App{
		DB:       h.db,
		Now:      func() time.Time { return h.now },
		Currency: "PKR",
		Out:      h.out,
		Err:      h.errOut,
	})
	root.SetArgs(args)

	err := root.ExecuteContext(context.Background())
	return h.out.String(), err
}

func (h *harness) mustRun(args ...string) string {
	h.t.Helper()
	out, err := h.run(args...)
	if err != nil {
		h.t.Fatalf("qarz %s: %v", strings.Join(args, " "), err)
	}
	return out
}

func (h *harness) wantFailure(args ...string) string {
	h.t.Helper()
	out, err := h.run(args...)
	if err == nil {
		h.t.Fatalf("qarz %s succeeded, want a failure; output was:\n%s", strings.Join(args, " "), out)
	}
	return err.Error()
}

func assertGolden(t *testing.T, name, got string) {
	t.Helper()

	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("create testdata: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s (run the tests with -update to create it): %v", path, err)
	}
	if got != string(want) {
		t.Errorf("output does not match %s\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
	}
}

func TestAddRecordsLoanAndPrintsItsNumber(t *testing.T) {
	h := newHarness(t)

	out := h.mustRun("add", "--to", "Ali", "--amount", "50000", "--due", "2026-12-20")
	if want := "loan 1 recorded: Ali owes 50,000 PKR by 2026-12-20\n"; out != want {
		t.Errorf("add printed %q, want %q", out, want)
	}

	row, err := store.GetLoan(context.Background(), h.db, 1, loan.DateOf(h.now))
	if err != nil {
		t.Fatalf("read back loan 1: %v", err)
	}
	if row.Loan.Borrower != "Ali" {
		t.Errorf("borrower = %q, want %q", row.Loan.Borrower, "Ali")
	}
	if row.Loan.PrincipalMinor != 5_000_000 {
		t.Errorf("principal = %d, want 5000000", row.Loan.PrincipalMinor)
	}
	if row.Loan.LentOn.String() != fixedToday {
		t.Errorf("lent on = %s, want today (%s)", row.Loan.LentOn, fixedToday)
	}
	if row.Loan.UID == "" {
		t.Error("loan was recorded without a unique tag")
	}
}

func TestEachLoanGetsItsOwnTag(t *testing.T) {
	h := newHarness(t)
	h.mustRun("add", "--to", "Ali", "--amount", "50000", "--due", "2026-12-20")
	h.mustRun("add", "--to", "Sara", "--amount", "80000", "--due", "2027-01-20")

	today := loan.DateOf(h.now)
	first, err := store.GetLoan(context.Background(), h.db, 1, today)
	if err != nil {
		t.Fatalf("read loan 1: %v", err)
	}
	second, err := store.GetLoan(context.Background(), h.db, 2, today)
	if err != nil {
		t.Fatalf("read loan 2: %v", err)
	}
	if first.Loan.UID == second.Loan.UID {
		t.Errorf("two loans share the tag %q", first.Loan.UID)
	}
}

func TestAddRejectsBadInput(t *testing.T) {
	h := newHarness(t)

	if got := h.wantFailure("add", "--to", "Ali", "--amount", "50000",
		"--due", "2026-11-09", "--lent-on", "2026-11-10"); !strings.Contains(got, "before the date of lending") {
		t.Errorf("error = %q, want it to mention the due date being before the date of lending", got)
	}
	if got := h.wantFailure("add", "--to", "Ali", "--amount", "0", "--due", "2026-12-20"); !strings.Contains(got, "positive") {
		t.Errorf("error = %q, want it to mention a positive amount", got)
	}
	if got := h.wantFailure("add", "--to", "  ", "--amount", "500", "--due", "2026-12-20"); !strings.Contains(got, "borrower") {
		t.Errorf("error = %q, want it to mention the missing borrower", got)
	}
}

func TestRepaymentsReduceTheBalance(t *testing.T) {
	h := newHarness(t)
	h.mustRun("add", "--to", "Ali", "--amount", "50000", "--due", "2026-12-20")

	out := h.mustRun("repay", "1", "--amount", "20000")
	if want := "loan 1: Ali repaid 20,000 PKR, 30,000 PKR still owed\n"; out != want {
		t.Errorf("repay printed %q, want %q", out, want)
	}

	out = h.mustRun("list")
	if !strings.Contains(out, "30,000 PKR") {
		t.Errorf("listing after a partial repayment = %q, want it to show the 30,000 balance", out)
	}
}

func TestFinalRepaymentPaysTheLoan(t *testing.T) {
	h := newHarness(t)
	h.mustRun("add", "--to", "Ali", "--amount", "50000", "--due", "2026-12-20")
	h.mustRun("repay", "1", "--amount", "20000")

	out := h.mustRun("repay", "1", "--amount", "30000")
	if want := "loan 1 paid in full: Ali repaid 50,000 PKR\n"; out != want {
		t.Errorf("final repay printed %q, want %q", out, want)
	}

	row, err := store.GetLoan(context.Background(), h.db, 1, loan.DateOf(h.now))
	if err != nil {
		t.Fatalf("read back loan 1: %v", err)
	}
	if row.Status != loan.StatusPaid {
		t.Errorf("status = %q, want %q", row.Status, loan.StatusPaid)
	}
	if row.BalanceMinor != 0 {
		t.Errorf("balance = %d, want 0", row.BalanceMinor)
	}
}

func TestRepaymentAboveTheBalanceIsRejected(t *testing.T) {
	h := newHarness(t)
	h.mustRun("add", "--to", "Ali", "--amount", "50000", "--due", "2026-12-20")
	h.mustRun("repay", "1", "--amount", "20000")

	got := h.wantFailure("repay", "1", "--amount", "30001")
	for _, want := range []string{"30,001", "30,000"} {
		if !strings.Contains(got, want) {
			t.Errorf("error = %q, want it to name %s", got, want)
		}
	}

	repayments, err := store.ListRepayments(context.Background(), h.db, 1)
	if err != nil {
		t.Fatalf("list repayments: %v", err)
	}
	if len(repayments) != 1 {
		t.Errorf("rejected repayment was still recorded: %d rows, want 1", len(repayments))
	}
}

func TestUnknownLoanIsRejected(t *testing.T) {
	h := newHarness(t)

	if got := h.wantFailure("repay", "7", "--amount", "100"); !strings.Contains(got, "7") {
		t.Errorf("error = %q, want it to name the loan number", got)
	}
	if got := h.wantFailure("show", "7"); !strings.Contains(got, "7") {
		t.Errorf("error = %q, want it to name the loan number", got)
	}
	if got := h.wantFailure("show", "banana"); !strings.Contains(got, "banana") {
		t.Errorf("error = %q, want it to quote the bad argument", got)
	}
}

func TestListFilters(t *testing.T) {
	h := newHarness(t)
	h.mustRun("add", "--to", "Ali", "--amount", "50000", "--due", "2026-12-20")
	h.mustRun("add", "--to", "Sara", "--amount", "80000", "--due", "2026-12-03")
	h.mustRun("add", "--to", "Imran", "--amount", "12000", "--due", "2026-12-20")
	h.mustRun("repay", "3", "--amount", "12000")

	open := h.mustRun("list")
	if strings.Contains(open, "Imran") {
		t.Errorf("default listing includes the paid loan:\n%s", open)
	}

	paid := h.mustRun("list", "--paid")
	if !strings.Contains(paid, "Imran") || strings.Contains(paid, "Ali") {
		t.Errorf("--paid listing = %q, want only Imran", paid)
	}

	overdue := h.mustRun("list", "--overdue")
	if !strings.Contains(overdue, "Sara") || strings.Contains(overdue, "Ali") {
		t.Errorf("--overdue listing = %q, want only Sara", overdue)
	}
	if !strings.Contains(overdue, "overdue 12d") {
		t.Errorf("--overdue listing = %q, want it to state the days late", overdue)
	}

	if got := h.wantFailure("list", "--paid", "--overdue"); !strings.Contains(got, "pick one") {
		t.Errorf("error = %q, want it to ask for a single filter", got)
	}
}

func TestEmptyListSaysSo(t *testing.T) {
	h := newHarness(t)

	if out := h.mustRun("list"); out != "no open loans\n" {
		t.Errorf("empty listing = %q, want %q", out, "no open loans\n")
	}
}

func TestListGolden(t *testing.T) {
	h := newHarness(t)
	h.mustRun("add", "--to", "Ali", "--amount", "50000", "--due", "2026-12-20")
	h.mustRun("add", "--to", "Sara", "--amount", "80000", "--due", "2026-12-03")
	h.mustRun("repay", "1", "--amount", "20000")

	assertGolden(t, "list-open", h.mustRun("list"))
}

func TestShowGolden(t *testing.T) {
	h := newHarness(t)
	h.mustRun("add", "--to", "Ali", "--amount", "50000", "--due", "2026-12-20",
		"--lent-on", "2026-11-10", "--note", "bike repair",
		"--meta", "purpose=bike", "--meta", "via=cash")

	assertGolden(t, "show-no-repayments", h.mustRun("show", "1"))

	h.mustRun("repay", "1", "--amount", "20000", "--on", "2026-11-20", "--note", "partial")
	assertGolden(t, "show-with-repayments", h.mustRun("show", "1"))

	h.mustRun("repay", "1", "--amount", "30000", "--on", "2026-12-01")
	assertGolden(t, "show-paid", h.mustRun("show", "1"))
}

func TestAmountAcceptsFormattedInput(t *testing.T) {
	h := newHarness(t)

	out := h.mustRun("add", "--to", "Ali", "--amount", "1,50,000", "--due", "2026-12-20")
	if !strings.Contains(out, "150,000 PKR") {
		t.Errorf("add printed %q, want the amount normalised to 150,000 PKR", out)
	}
}

func TestLoanDetailsRoundTrip(t *testing.T) {
	h := newHarness(t)
	h.mustRun("add", "--to", "Ali", "--amount", "50000", "--due", "2026-12-20",
		"--meta", "purpose=bike", "--meta", "via=cash")

	row, err := store.GetLoan(context.Background(), h.db, 1, loan.DateOf(h.now))
	if err != nil {
		t.Fatalf("read back loan 1: %v", err)
	}
	if row.Loan.Metadata["purpose"] != "bike" || row.Loan.Metadata["via"] != "cash" {
		t.Errorf("details = %v, want purpose=bike and via=cash", row.Loan.Metadata)
	}

	if got := h.wantFailure("add", "--to", "Ali", "--amount", "500",
		"--due", "2026-12-20", "--meta", "nonsense"); !strings.Contains(got, "key=value") {
		t.Errorf("error = %q, want it to explain the key=value form", got)
	}
}

func TestRepaymentAgainstUnknownLoanIsRefusedByTheDatabase(t *testing.T) {
	h := newHarness(t)

	_, err := store.InsertRepayment(context.Background(), h.db, loan.Repayment{
		LoanID:      404,
		AmountMinor: 1000,
		PaidOn:      loan.DateOf(h.now),
	})
	if err == nil {
		t.Fatal("inserting a repayment against a non-existent loan succeeded; the foreign_keys pragma is not in effect")
	}
	if errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected a database constraint failure, got %v", err)
	}
}
