package lunchmoney

import (
	"encoding/json"
	"fmt"
	"time"
)

// Timestamp is a time the API documents as a date-time but may also return as
// a plain YYYY-MM-DD date. Decoding either shape into a time.Time directly
// fails on the date-only form, which would sink the whole response.
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
	if t.Time.IsZero() {
		return []byte("null"), nil
	}

	return json.Marshal(t.Time.Format(time.RFC3339Nano))
}
