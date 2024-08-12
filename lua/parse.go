package lua

func MustNil(v Value) {
	_, ok := v.(Nil)
	if !ok {
		panic("not a nil")
	}
}

func MustString(v Value) string {
	s, ok := v.(String)
	if !ok {
		panic("not a string")
	}
	return string(s)
}

func MustBoolean(v Value) bool {
	b, ok := v.(Boolean)
	if !ok {
		panic("not a boolean")
	}
	return bool(b)
}

func MustInteger(v Value) int64 {
	i, ok := v.(Integer)
	if !ok {
		panic("not an integer")
	}
	return int64(i)
}

func MustReal(v Value) float64 {
	r, ok := v.(Real)
	if !ok {
		panic("not a real")
	}
	return float64(r)
}

func MustTable(v Value) Table {
	t, ok := v.(Table)
	if !ok {
		panic("not a table")
	}
	return t
}
