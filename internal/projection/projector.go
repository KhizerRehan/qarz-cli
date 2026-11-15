package projection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/KhizerRehan/qarz-cli/internal/loan"
	"github.com/KhizerRehan/qarz-cli/internal/store"
)

var ErrPending = errors.New("the calendar was not updated")

type Projector struct {
	DB       *sql.DB
	Calendar Calendar
	Now      func() time.Time
	Currency string
}

func (p *Projector) Project(ctx context.Context, loanID int64) error {
	today := loan.DateOf(p.Now())

	row, err := store.GetLoan(ctx, p.DB, loanID, today)
	if err != nil {
		return err
	}

	desired, wanted := Desired(row, today, p.Currency)

	if !wanted {
		return p.remove(ctx, row)
	}
	if row.Loan.CalendarEventID == "" {
		return p.create(ctx, row, desired)
	}
	return p.update(ctx, row, desired)
}

func (p *Projector) create(ctx context.Context, row store.LoanRow, desired Event) error {
	if err := p.mark(ctx, row, loan.SyncPendingCreate, ""); err != nil {
		return err
	}

	existing, found, err := p.Calendar.FindByTag(ctx, row.Loan.UID)
	if err != nil {
		return pending(err)
	}
	if found {
		desired.ID = existing.ID
		if err := p.Calendar.Update(ctx, desired); err != nil {
			return pending(err)
		}
		return p.mark(ctx, row, loan.SyncSynced, existing.ID)
	}

	id, err := p.Calendar.Create(ctx, desired)
	if err != nil {
		return pending(err)
	}

	return p.mark(ctx, row, loan.SyncSynced, id)
}

func (p *Projector) update(ctx context.Context, row store.LoanRow, desired Event) error {
	if err := p.mark(ctx, row, loan.SyncPendingUpdate, row.Loan.CalendarEventID); err != nil {
		return err
	}

	if err := p.Calendar.Update(ctx, desired); err != nil {
		return pending(err)
	}

	return p.mark(ctx, row, loan.SyncSynced, row.Loan.CalendarEventID)
}

func (p *Projector) remove(ctx context.Context, row store.LoanRow) error {
	if row.Loan.CalendarEventID == "" {
		return p.mark(ctx, row, loan.SyncSynced, "")
	}

	if err := p.mark(ctx, row, loan.SyncPendingDelete, row.Loan.CalendarEventID); err != nil {
		return err
	}

	if err := p.Calendar.Delete(ctx, row.Loan.CalendarEventID); err != nil {
		return pending(err)
	}

	return p.mark(ctx, row, loan.SyncSynced, "")
}

func (p *Projector) mark(ctx context.Context, row store.LoanRow, state loan.SyncState, eventID string) error {
	return store.MarkSyncState(ctx, p.DB, row.Loan.ID, state, eventID, p.Now())
}

func pending(err error) error {
	return fmt.Errorf("%w: %v", ErrPending, err)
}
