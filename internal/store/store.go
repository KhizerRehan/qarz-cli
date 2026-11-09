package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"qarz/internal/loan"
)

var ErrNotFound = errors.New("no loan with that number")

type LoanRow struct {
	Loan         loan.Loan
	RepaidMinor  loan.Minor
	BalanceMinor loan.Minor
	Status       loan.Status
}

type Filter string

const (
	FilterOpen    Filter = "open"
	FilterPaid    Filter = "paid"
	FilterOverdue Filter = "overdue"
	FilterVoid    Filter = "void"
	FilterAll     Filter = "all"
)

func Open(path string) (*sql.DB, error) {
	dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open ledger at %s: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}

const loanColumns = `
  l.id, l.qarz_uid, l.borrower, l.principal_minor, l.lent_on, l.due_on,
  COALESCE(l.note, ''), COALESCE(l.metadata, ''), COALESCE(l.voided_at, ''),
  COALESCE(l.calendar_event_id, ''), l.sync_state,
  COALESCE(SUM(r.amount_minor), 0)`

func InsertLoan(ctx context.Context, db *sql.DB, l loan.Loan, now time.Time) (int64, error) {
	metadata, err := encodeMetadata(l.Metadata)
	if err != nil {
		return 0, err
	}
	state := l.SyncState
	if state == "" {
		state = loan.SyncPendingCreate
	}
	stamp := timestamp(now)

	res, err := db.ExecContext(ctx, `
		INSERT INTO loans (qarz_uid, borrower, principal_minor, lent_on, due_on,
		                   note, metadata, sync_state, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		l.UID, l.Borrower, int64(l.PrincipalMinor), l.LentOn.String(), l.DueOn.String(),
		nullable(l.Note), nullable(metadata), string(state), stamp, stamp)
	if err != nil {
		return 0, fmt.Errorf("record loan: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read back loan number: %w", err)
	}
	return id, nil
}

func GetLoan(ctx context.Context, db *sql.DB, id int64, today loan.Date) (LoanRow, error) {
	query := `SELECT` + loanColumns + `
		FROM loans l LEFT JOIN repayments r ON r.loan_id = l.id
		WHERE l.id = ?
		GROUP BY l.id`

	row, err := scanLoanRow(db.QueryRowContext(ctx, query, id), today)
	if errors.Is(err, sql.ErrNoRows) {
		return LoanRow{}, fmt.Errorf("%w: %d", ErrNotFound, id)
	}
	if err != nil {
		return LoanRow{}, fmt.Errorf("read loan %d: %w", id, err)
	}
	return row, nil
}

func ListLoans(ctx context.Context, db *sql.DB, filter Filter, today loan.Date) ([]LoanRow, error) {
	query := `SELECT` + loanColumns + `
		FROM loans l LEFT JOIN repayments r ON r.loan_id = l.id
		GROUP BY l.id
		ORDER BY l.due_on ASC, l.id ASC`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list loans: %w", err)
	}
	defer rows.Close()

	var out []LoanRow
	for rows.Next() {
		row, err := scanLoanRow(rows, today)
		if err != nil {
			return nil, fmt.Errorf("list loans: %w", err)
		}
		if matches(filter, row.Status) {
			out = append(out, row)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list loans: %w", err)
	}
	return out, nil
}

func InsertRepayment(ctx context.Context, db *sql.DB, r loan.Repayment) (int64, error) {
	res, err := db.ExecContext(ctx, `
		INSERT INTO repayments (loan_id, amount_minor, paid_on, note)
		VALUES (?, ?, ?, ?)`,
		r.LoanID, int64(r.AmountMinor), r.PaidOn.String(), nullable(r.Note))
	if err != nil {
		return 0, fmt.Errorf("record repayment: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read back repayment number: %w", err)
	}
	return id, nil
}

func ListRepayments(ctx context.Context, db *sql.DB, loanID int64) ([]loan.Repayment, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, loan_id, amount_minor, paid_on, COALESCE(note, '')
		FROM repayments WHERE loan_id = ?
		ORDER BY paid_on ASC, id ASC`, loanID)
	if err != nil {
		return nil, fmt.Errorf("list repayments for loan %d: %w", loanID, err)
	}
	defer rows.Close()

	var out []loan.Repayment
	for rows.Next() {
		var (
			r      loan.Repayment
			amount int64
			paidOn string
		)
		if err := rows.Scan(&r.ID, &r.LoanID, &amount, &paidOn, &r.Note); err != nil {
			return nil, fmt.Errorf("list repayments for loan %d: %w", loanID, err)
		}
		r.AmountMinor = loan.Minor(amount)
		if r.PaidOn, err = loan.ParseDate(paidOn); err != nil {
			return nil, fmt.Errorf("repayment %d: %w", r.ID, err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list repayments for loan %d: %w", loanID, err)
	}
	return out, nil
}

func MarkSyncState(ctx context.Context, db *sql.DB, id int64, state loan.SyncState, eventID string, now time.Time) error {
	_, err := db.ExecContext(ctx, `
		UPDATE loans SET sync_state = ?, calendar_event_id = ?, updated_at = ?
		WHERE id = ?`, string(state), nullable(eventID), timestamp(now), id)
	if err != nil {
		return fmt.Errorf("update sync state of loan %d: %w", id, err)
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanLoanRow(src scanner, today loan.Date) (LoanRow, error) {
	var (
		row       LoanRow
		principal int64
		repaid    int64
		lentOn    string
		dueOn     string
		metadata  string
		state     string
	)

	err := src.Scan(&row.Loan.ID, &row.Loan.UID, &row.Loan.Borrower, &principal,
		&lentOn, &dueOn, &row.Loan.Note, &metadata, &row.Loan.VoidedAt,
		&row.Loan.CalendarEventID, &state, &repaid)
	if err != nil {
		return LoanRow{}, err
	}

	row.Loan.PrincipalMinor = loan.Minor(principal)
	row.Loan.SyncState = loan.SyncState(state)
	if row.Loan.LentOn, err = loan.ParseDate(lentOn); err != nil {
		return LoanRow{}, err
	}
	if row.Loan.DueOn, err = loan.ParseDate(dueOn); err != nil {
		return LoanRow{}, err
	}
	if row.Loan.Metadata, err = decodeMetadata(metadata); err != nil {
		return LoanRow{}, err
	}

	row.RepaidMinor = loan.Minor(repaid)
	row.BalanceMinor = loan.Balance(row.Loan.PrincipalMinor, row.RepaidMinor)
	row.Status = loan.DeriveStatus(row.Loan.Voided(), row.BalanceMinor, row.Loan.DueOn, today)
	return row, nil
}

func matches(filter Filter, status loan.Status) bool {
	switch filter {
	case FilterAll:
		return true
	case FilterOpen:
		return status == loan.StatusOpen || status == loan.StatusOverdue
	case FilterPaid:
		return status == loan.StatusPaid
	case FilterOverdue:
		return status == loan.StatusOverdue
	case FilterVoid:
		return status == loan.StatusVoid
	default:
		return false
	}
}

func encodeMetadata(m map[string]string) (string, error) {
	if len(m) == 0 {
		return "", nil
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("encode loan details: %w", err)
	}
	return string(raw), nil
}

func decodeMetadata(raw string) (map[string]string, error) {
	if raw == "" {
		return nil, nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, fmt.Errorf("decode loan details: %w", err)
	}
	return m, nil
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func timestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func PendingLoanIDs(ctx context.Context, db *sql.DB) ([]int64, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id FROM loans WHERE sync_state <> ? ORDER BY id ASC`, string(loan.SyncSynced))
	if err != nil {
		return nil, fmt.Errorf("list loans awaiting the calendar: %w", err)
	}
	defer rows.Close()

	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("list loans awaiting the calendar: %w", err)
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list loans awaiting the calendar: %w", err)
	}
	return out, nil
}
