package projectiontest

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/KhizerRehan/qarz-cli/internal/projection"
)

type Fake struct {
	mu      sync.Mutex
	events  map[string]projection.Event
	next    int
	Failure error
	Creates int
	Updates int
	Deletes int
}

func NewFake() *Fake {
	return &Fake{events: map[string]projection.Event{}}
}

func (f *Fake) Create(_ context.Context, event projection.Event) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.Failure != nil {
		return "", f.Failure
	}

	f.next++
	id := fmt.Sprintf("event-%d", f.next)
	event.ID = id
	f.events[id] = event
	f.Creates++
	return id, nil
}

func (f *Fake) Update(_ context.Context, event projection.Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.Failure != nil {
		return f.Failure
	}
	if _, ok := f.events[event.ID]; !ok {
		return fmt.Errorf("no event %s", event.ID)
	}

	f.events[event.ID] = event
	f.Updates++
	return nil
}

func (f *Fake) Delete(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.Failure != nil {
		return f.Failure
	}

	delete(f.events, id)
	f.Deletes++
	return nil
}

func (f *Fake) ListTagged(_ context.Context) ([]projection.Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.Failure != nil {
		return nil, f.Failure
	}
	return f.snapshot(), nil
}

func (f *Fake) FindByTag(_ context.Context, tag string) (projection.Event, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.Failure != nil {
		return projection.Event{}, false, f.Failure
	}
	for _, event := range f.events {
		if event.Tag == tag {
			return event, true, nil
		}
	}
	return projection.Event{}, false, nil
}

func (f *Fake) Events() []projection.Event {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.snapshot()
}

func (f *Fake) snapshot() []projection.Event {
	out := make([]projection.Event, 0, len(f.events))
	for _, event := range f.events {
		out = append(out, event)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
