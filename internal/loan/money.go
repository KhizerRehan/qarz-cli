package loan

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Minor int64

func ParseAmount(s string) (Minor, error) {
	raw := strings.TrimSpace(s)
	raw = strings.ReplaceAll(raw, ",", "")
	raw = strings.ReplaceAll(raw, "_", "")
	if raw == "" {
		return 0, errors.New("amount is empty")
	}
	raw = strings.TrimPrefix(raw, "+")

	whole, frac := raw, ""
	if i := strings.IndexByte(raw, '.'); i >= 0 {
		whole, frac = raw[:i], raw[i+1:]
	}
	if whole == "" {
		whole = "0"
	}
	if len(frac) > 2 {
		return 0, fmt.Errorf("amount %s has more than two decimal places", s)
	}

	if !allDigits(whole) || !allDigits(frac) {
		return 0, fmt.Errorf("invalid amount %q", s)
	}
	for len(frac) < 2 {
		frac += "0"
	}

	major, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("amount %s is too large", s)
	}
	minor, err := strconv.ParseInt(frac, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q", s)
	}
	if major > (math.MaxInt64-minor)/100 {
		return 0, fmt.Errorf("amount %s is too large", s)
	}

	total := Minor(major*100 + minor)
	if total <= 0 {
		return 0, fmt.Errorf("amount must be positive, got %s", s)
	}
	return total, nil
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func (m Minor) String() string {
	value := int64(m)
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}

	out := sign + groupDigits(strconv.FormatInt(value/100, 10))
	if rem := value % 100; rem != 0 {
		out += fmt.Sprintf(".%02d", rem)
	}
	return out
}

func (m Minor) WithCurrency(currency string) string {
	return fmt.Sprintf("%s %s", m, currency)
}

func groupDigits(digits string) string {
	if len(digits) <= 3 {
		return digits
	}

	lead := len(digits) % 3
	var b strings.Builder
	if lead > 0 {
		b.WriteString(digits[:lead])
	}
	for i := lead; i < len(digits); i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(digits[i : i+3])
	}
	return b.String()
}
