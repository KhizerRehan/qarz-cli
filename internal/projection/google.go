package projection

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"github.com/KhizerRehan/qarz-cli/internal/loan"
)

type Google struct {
	service    *calendar.Service
	calendarID string
	location   *time.Location
}

func NewGoogle(ctx context.Context, client *http.Client, calendarID string, location *time.Location) (*Google, error) {
	service, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("reach Google Calendar: %w", err)
	}
	return &Google{service: service, calendarID: calendarID, location: location}, nil
}

func (g *Google) Create(ctx context.Context, event Event) (string, error) {
	created, err := g.service.Events.Insert(g.calendarID, g.toGoogle(event)).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("create the calendar event: %w", err)
	}
	return created.Id, nil
}

func (g *Google) Update(ctx context.Context, event Event) error {
	if _, err := g.service.Events.Update(g.calendarID, event.ID, g.toGoogle(event)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("update calendar event %s: %w", event.ID, err)
	}
	return nil
}

func (g *Google) Delete(ctx context.Context, id string) error {
	if err := g.service.Events.Delete(g.calendarID, id).Context(ctx).Do(); err != nil {
		if isGone(err) {
			return nil
		}
		return fmt.Errorf("delete calendar event %s: %w", id, err)
	}
	return nil
}

func (g *Google) ListTagged(ctx context.Context) ([]Event, error) {
	var out []Event

	call := g.service.Events.List(g.calendarID).
		PrivateExtendedProperty(MarkerKey + "=" + MarkerValue).
		ShowDeleted(false).
		MaxResults(2500)

	err := call.Pages(ctx, func(page *calendar.Events) error {
		for _, item := range page.Items {
			event, err := g.fromGoogle(item)
			if err != nil {
				return err
			}
			out = append(out, event)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list the calendar events qarz owns: %w", err)
	}

	return out, nil
}

func (g *Google) FindByTag(ctx context.Context, tag string) (Event, bool, error) {
	page, err := g.service.Events.List(g.calendarID).
		PrivateExtendedProperty(TagKey + "=" + tag).
		ShowDeleted(false).
		MaxResults(2).
		Context(ctx).Do()
	if err != nil {
		return Event{}, false, fmt.Errorf("look up the calendar event tagged %s: %w", tag, err)
	}
	if len(page.Items) == 0 {
		return Event{}, false, nil
	}

	event, err := g.fromGoogle(page.Items[0])
	if err != nil {
		return Event{}, false, err
	}
	return event, true, nil
}

func (g *Google) toGoogle(event Event) *calendar.Event {
	start := time.Date(
		time.Time(event.Date).Year(), time.Time(event.Date).Month(), time.Time(event.Date).Day(),
		SlotHour, 0, 0, 0, g.location)
	end := start.Add(SlotMinutes * time.Minute)

	overrides := make([]*calendar.EventReminder, 0, len(event.Reminders))
	for _, minutes := range event.Reminders {
		overrides = append(overrides, &calendar.EventReminder{Method: "popup", Minutes: int64(minutes)})
	}

	return &calendar.Event{
		Summary:     event.Title,
		Description: event.Description,
		Start:       &calendar.EventDateTime{DateTime: start.Format(time.RFC3339), TimeZone: g.location.String()},
		End:         &calendar.EventDateTime{DateTime: end.Format(time.RFC3339), TimeZone: g.location.String()},
		Reminders: &calendar.EventReminders{
			UseDefault:      false,
			Overrides:       overrides,
			ForceSendFields: []string{"UseDefault"},
		},
		ExtendedProperties: &calendar.EventExtendedProperties{
			Private: map[string]string{
				TagKey:    event.Tag,
				MarkerKey: MarkerValue,
			},
		},
	}
}

func (g *Google) fromGoogle(item *calendar.Event) (Event, error) {
	event := Event{
		ID:          item.Id,
		Title:       item.Summary,
		Description: item.Description,
	}

	if item.ExtendedProperties != nil {
		event.Tag = item.ExtendedProperties.Private[TagKey]
	}

	when, err := slotDate(item)
	if err != nil {
		return Event{}, err
	}
	event.Date = when

	if item.Reminders != nil {
		for _, override := range item.Reminders.Overrides {
			event.Reminders = append(event.Reminders, int(override.Minutes))
		}
	}

	return event, nil
}

func slotDate(item *calendar.Event) (loan.Date, error) {
	if item.Start == nil {
		return loan.Date{}, fmt.Errorf("calendar event %s has no start", item.Id)
	}
	if item.Start.Date != "" {
		return loan.ParseDate(item.Start.Date)
	}

	parsed, err := time.Parse(time.RFC3339, item.Start.DateTime)
	if err != nil {
		return loan.Date{}, fmt.Errorf("read the start of calendar event %s: %w", item.Id, err)
	}
	return loan.DateOf(parsed), nil
}

func isGone(err error) bool {
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		return apiErr.Code == http.StatusNotFound || apiErr.Code == http.StatusGone
	}
	return false
}
