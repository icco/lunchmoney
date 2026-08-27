package lunchmoney

import (
	"encoding/json"
	"fmt"
	"time"
)

// Timestamp is a time the API may return as a date-time or a plain date. A
// time.Time alone fails on the date-only form.
type Timestamp struct {
	time.Time
}

var timestampLayouts = []string{time.RFC3339Nano, time.DateOnly}

// UnmarshalJSON accepts an ISO 8601 date-time, a plain date, or null.
func (t *Timestamp) UnmarshalJSON(b []byte) error {
	var s *string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if s == nil || *s == "" {
		t.Time = time.Time{}
		return nil
	}

	for _, layout := range timestampLayouts {
		parsed, err := time.Parse(layout, *s)
		if err == nil {
			t.Time = parsed
			return nil
		}
	}

	return fmt.Errorf("%q is not a valid date or date-time", *s)
}

// MarshalJSON writes the timestamp back out in ISO 8601 extended format.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}

	return json.Marshal(t.Format(time.RFC3339Nano))
}
