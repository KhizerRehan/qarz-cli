package projection

import (
	"context"

	"github.com/KhizerRehan/qarz-cli/internal/loan"
)

const (
	TagKey      = "qarz_uid"
	MarkerKey   = "qarz"
	MarkerValue = "1"

	SlotHour            = 9
	SlotMinutes         = 15
	RemindDayBefore     = 24 * 60
	RemindShortlyBefore = 10
)

type Event struct {
	ID          string
	Tag         string
	Title       string
	Description string
	Date        loan.Date
	Reminders   []int
}

type Calendar interface {
	Create(ctx context.Context, event Event) (string, error)
	Update(ctx context.Context, event Event) error
	Delete(ctx context.Context, id string) error
	ListTagged(ctx context.Context) ([]Event, error)
	FindByTag(ctx context.Context, tag string) (Event, bool, error)
}
