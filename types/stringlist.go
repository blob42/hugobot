package types

import (
	"database/sql/driver"
	"errors"
	"strings"
)

type StringList []string

func (sl *StringList) Scan(src interface{}) error {

	switch v := src.(type) {
	case string:
		*sl = strings.Split(v, ",")
	case []byte:
		*sl = strings.Split(string(v), ",")
	default:
		return errors.New("Could not scan to []string")
	}

	return nil
}

func (sl StringList) Value() (driver.Value, error) {
	result := strings.Join(sl, ",")
	return driver.Value(result), nil
}
