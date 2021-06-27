package validator

import "testing"

func TestValid(t *testing.T) {
	type TestCase struct {
		constraints []bool
		result bool
	}

	tests := []TestCase{
		{[]bool{true, true, false}, false},
		{[]bool{true, true, true}, true},
		{[]bool{false, false, false}, false},
	}

	for i, test := range tests {
		if r := Valid(test.constraints...); r != test.result {
			t.Errorf("%d: Expected %t, got %t", i, test.result, r)
		}
	}
}

func TestInSlice(t *testing.T) {
	type TestCase struct {
		val interface{}
		slice []interface{}
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
		val interface{}
		set map[interface{}]bool
		result bool
	}

	tests := []TestCase{
		{"val2", map[interface{}]bool{"val1": true, "val2": true, "val3": true}, true},
		{5.0, map[interface{}]bool{2.3: true, 7.8: true, 1.5: true, 5.0: true}, true},
		{2, map[interface{}]bool{1: true, 3: true, 5: true, 7: true, 9: true}, false},
	}

	for i, test := range tests {
		if r := InSet(test.val, test.set); r != test.result {
			t.Errorf("%d: Expected %t, got %t", i, test.result, r)
		}
	}
}
