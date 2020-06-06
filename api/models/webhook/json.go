package webhook

import (
	"bytes"
	"database/sql/driver"
	"errors"
)

//JSON represents a JSON Object.
type JSON []byte

//Value implements a JSON interface
func (j JSON) Value() (driver.Value, error) {
	if j.IsNull() {
		return nil, nil
	}
	return string(j), nil
}

//Scan implements a JSON interface
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	s, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source")
	}
	*j = append((*j)[0:0], s...)
	return nil
}

//MarshalJSON implements a JSON interface
func (j JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return j, nil
}

//UnmarshalJSON implements a JSON interface
func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("null point exception")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

//IsNull implements a JSON interface
func (j JSON) IsNull() bool {
	return len(j) == 0 || string(j) == "null"
}

//Equals implements a JSON interface
func (j JSON) Equals(j1 JSON) bool {
	return bytes.Equal([]byte(j), []byte(j1))
}
