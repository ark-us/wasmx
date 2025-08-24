package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type StringUint64 uint64

func (s StringUint64) ToString() string {
	return strconv.FormatUint(uint64(s), 10)
}

// MarshalJSON makes it encode as a JSON string
func (s StringUint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(s), 10))
}

// UnmarshalJSON accepts either a string or a number
func (s *StringUint64) UnmarshalJSON(b []byte) error {
	// Try as string
	var str string
	if err := json.Unmarshal(b, &str); err == nil {
		v, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}
		*s = StringUint64(v)
		return nil
	}

	// Try as number
	var num uint64
	if err := json.Unmarshal(b, &num); err == nil {
		*s = StringUint64(num)
		return nil
	}

	return fmt.Errorf("invalid value for StringUint64: %s", string(b))
}
