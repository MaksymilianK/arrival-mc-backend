package validator

import (
	"errors"
)

var ErrNotFound = errors.New("resource not found")
var ErrValidation = errors.New("invalid data")

func Validate(constraints ...bool) error {
	for _, c := range constraints {
		if !c {
			return ErrValidation
		}
	}
	return nil
}

func InSlice(val interface{}, allowed ...interface{}) bool {
	for _, a := range allowed {
		if val == a {
			return true
		}
	}
	return false
}

func InSet(val interface{}, allowed map[interface{}]struct{}) bool {
	_, ok := allowed[val]
	return ok
}

func EveryKey(vals map[interface{}]interface{}, validator func(interface{}) bool) bool {
	for i := range vals {
		if !validator(i) {
			return false
		}
	}
	return true
}

func EveryVal(vals map[interface{}]interface{}, validator func(interface{}) bool) bool {
	for _, v := range vals {
		if !validator(v) {
			return false
		}
	}
	return true
}
