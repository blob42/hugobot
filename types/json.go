package types

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type JsonMap map[string]interface{}

func (m *JsonMap) Scan(src interface{}) error {

	val, ok := src.([]byte)
	if !ok {
		return errors.New("not []byte")
	}

	jsonDecoder := json.NewDecoder(bytes.NewBuffer(val))

	err := jsonDecoder.Decode(m)
	if err != nil {
		return err
	}

	return nil
}

func (m JsonMap) Value() (driver.Value, error) {

	val, err := json.Marshal(m)
	if err != nil {
		return driver.Value(nil), err
	}

	return driver.Value(val), nil

}
