package orm

import (
	"encoding/json"
	"errors"
	"time"
)

type Duration int64

func (d *Duration) Duration() time.Duration { return time.Duration(*d) }

func (d *Duration) UnmarshalJSON(rawBytes []byte) (err error) {
	var valParsed time.Duration
	if len(rawBytes) > 0 && rawBytes[0] == '"' {
		var valText string
		if err = json.Unmarshal(rawBytes, &valText); err != nil {
			return err
		}
		if valParsed, err = time.ParseDuration(valText); err != nil {
			return err
		}
	} else if err = json.Unmarshal(rawBytes, &valParsed); err != nil {
		return err
	}
	*d = Duration(valParsed)
	return nil
}

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var v interface{}
	if err = unmarshal(&v); err != nil {
		return err
	}
	switch value := v.(type) {
	case int:
		*((*time.Duration)(d)) = time.Millisecond * time.Duration(value)
	case string:
		*((*time.Duration)(d)), err = time.ParseDuration(value)
	default:
		err = errors.New("invalid duration format")
	}
	return err
}

func fallback[T comparable](vals ...T) T {
	var emptyVal = empty[T]()
	for _, val := range vals {
		if val != emptyVal {
			return val
		}
	}
	return emptyVal
}

func empty[T any]() T {
	var val T
	return val
}
