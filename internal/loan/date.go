package loan

import (
	"fmt"
	"time"
)

const DateLayout = "2006-01-02"

type Date time.Time

func ParseDate(s string) (Date, error) {
	t, err := time.Parse(DateLayout, s)
	if err != nil {
		return Date{}, fmt.Errorf("invalid date %q: expected YYYY-MM-DD", s)
	}
	return Date(t), nil
}

func DateOf(t time.Time) Date {
	y, m, d := t.Date()
	return Date(time.Date(y, m, d, 0, 0, 0, 0, time.UTC))
}

func (d Date) String() string {
	return time.Time(d).Format(DateLayout)
}

func (d Date) Before(other Date) bool {
	return time.Time(d).Before(time.Time(other))
}

func (d Date) DaysSince(other Date) int {
	return int(time.Time(d).Sub(time.Time(other)) / (24 * time.Hour))
}

func (d Date) Later(other Date) Date {
	if d.Before(other) {
		return other
	}
	return d
}
