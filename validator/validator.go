package validator

func Valid(constraints ...bool) bool {
	for _, c := range constraints {
		if !c {
			return false
		}
	}
	return true
}

func InSlice(val interface{}, allowed ...interface{}) bool {
	for _, a := range allowed {
		if val == a {
			return true
		}
	}
	return false
}

func InSet(val interface{}, allowed map[interface{}]bool) bool {
	return allowed[val]
}