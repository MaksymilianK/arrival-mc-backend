package validator

import "testing"

func TestValid(t *testing.T) {
	type TestCase struct {
		constraints []bool
		result      error
	}

	tests := []TestCase{
		{[]bool{true, true, false}, ErrValidation},
		{[]bool{true, true, true}, nil},
		{[]bool{false, false, false}, ErrValidation},
	}

	for i, test := range tests {
		if r := Validate(test.constraints...); r != test.result {
			t.Errorf("%d: Expected %t, got %t", i, test.result, r)
		}
	}
}

func TestInSlice(t *testing.T) {
	type TestCase struct {
		val    interface{}
		slice  []interface{}
		result bool
	}

	tests := []TestCase{
		{"val2", []interface{}{"val1", "val2", "val3"}, true},
		{5.0, []interface{}{2.3, 7.8, 1.5, 5.0}, true},
		{2, []interface{}{1, 3, 5, 7, 9}, false},
	}

	for i, test := range tests {
		if r := InSlice(test.val, test.slice...); r != test.result {
			t.Errorf("%d: Expected %t, got %t", i, test.result, r)
		}
	}
}

func TestInSet(t *testing.T) {
	type TestCase struct {
		val    interface{}
		set    map[interface{}]struct{}
		result bool
	}

	tests := []TestCase{
		{"val2", map[interface{}]struct{}{"val1": {}, "val2": {}, "val3": {}}, true},
		{5.0, map[interface{}]struct{}{2.3: {}, 7.8: {}, 1.5: {}, 5.0: {}}, true},
		{2, map[interface{}]struct{}{1: {}, 3: {}, 5: {}, 7: {}, 9: {}}, false},
	}

	for i, test := range tests {
		if r := InSet(test.val, test.set); r != test.result {
			t.Errorf("%d: Expected %t, got %t", i, test.result, r)
		}
	}
}
