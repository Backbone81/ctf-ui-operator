package ctfdapi

import (
	"encoding/json"
	"time"
)

// DateOnly behaves like a time.Time but marshals/unmarshals only the date portion.
type DateOnly struct {
	time.Time
}

func NewDateOnly(t time.Time) DateOnly {
	return DateOnly{
		Time: t,
	}
}

var _ json.Marshaler = (*DateOnly)(nil)

func (d DateOnly) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	dateStr := d.Format("2006-01-02")
	return json.Marshal(dateStr)
}

var _ json.Unmarshaler = (*DateOnly)(nil)

func (d *DateOnly) UnmarshalJSON(data []byte) error {
	var dateStr string
	if err := json.Unmarshal(data, &dateStr); err != nil {
		return err
	}

	if dateStr == "" {
		*d = DateOnly{}
		return nil
	}

	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return err
	}

	*d = DateOnly{t}
	return nil
}
