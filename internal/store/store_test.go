package store_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/KhizerRehan/qarz-cli/internal/loan"
	"github.com/KhizerRehan/qarz-cli/internal/store"
)

const calendarEraSchema = `
CREATE TABLE loans (
  id                INTEGER PRIMARY KEY AUTOINCREMENT,
  qarz_uid          TEXT    NOT NULL UNIQUE,
  borrower          TEXT    NOT NULL,
  principal_minor   INTEGER NOT NULL CHECK (principal_minor > 0),
  lent_on           TEXT    NOT NULL,
  due_on            TEXT    NOT NULL,
  note              TEXT,
  metadata          TEXT,
  voided_at         TEXT,
  calendar_event_id TEXT,
  sync_state        TEXT    NOT NULL CHECK (sync_state IN
                      ('synced','pending_create','pending_update','pending_delete')),
  created_at        TEXT    NOT NULL,
  updated_at        TEXT    NOT NULL,
  CHECK (due_on >= lent_on)
);

CREATE TABLE repayments (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  loan_id      INTEGER NOT NULL REFERENCES loans(id),
  amount_minor INTEGER NOT NULL CHECK (amount_minor > 0),
  paid_on      TEXT    NOT NULL,
  note         TEXT
);

INSERT INTO loans (id, qarz_uid, borrower, principal_minor, lent_on, due_on, note,
                   calendar_event_id, sync_state, created_at, updated_at)
VALUES (3, 'tag-3', 'Ali', 5000000, '2026-11-10', '2026-12-20', 'bike repair',
        'event-1', 'synced', '2026-11-10T00:00:00Z', '2026-11-10T00:00:00Z');

INSERT INTO repayments (loan_id, amount_minor, paid_on) VALUES (3, 2000000, '2026-11-20');
`

func TestMigrateDropsTheCalendarColumnsAndKeepsTheLedger(t *testing.T) {
	ctx := context.Background()

	db, err := store.Open(filepath.Join(t.TempDir(), "qarz.db"))
	if err != nil {
		t.Fatalf("open ledger: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.ExecContext(ctx, calendarEraSchema); err != nil {
		t.Fatalf("build a calendar-era ledger: %v", err)
	}

	for range 2 {
		if err := store.Migrate(ctx, db); err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}

	today, err := loan.ParseDate("2026-12-15")
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}

	row, err := store.GetLoan(ctx, db, 3, today)
	if err != nil {
		t.Fatalf("read back loan 3: %v", err)
	}
	if row.Loan.Borrower != "Ali" || row.Loan.Note != "bike repair" || row.BalanceMinor != 3_000_000 {
		t.Errorf("loan 3 after migrating = %+v, want Ali, the note, and a 30,000 balance", row)
	}

	id, err := store.InsertLoan(ctx, db, loan.Loan{
		UID:            "tag-new",
		Borrower:       "Sara",
		PrincipalMinor: 100,
		LentOn:         today,
		DueOn:          today,
	}, time.Now())
	if err != nil {
		t.Fatalf("record a loan after migrating: %v", err)
	}
	if id != 4 {
		t.Errorf("new loan number = %d, want 4 so numbering carries on", id)
	}

	if _, err := store.InsertRepayment(ctx, db, loan.Repayment{
		LoanID: 404, AmountMinor: 1, PaidOn: today,
	}); err == nil {
		t.Error("a repayment against a missing loan was accepted; foreign keys are off after migrating")
	}
}
