package lua

type Value interface {
	LuaType() uint8
}

type Nil struct{}

func (v Nil) LuaType() uint8 {
	return LUA_NIL
}

type Boolean bool

func (v Boolean) LuaType() uint8 {
	return LUA_BOOLEAN
}

type Integer int64

func (v Integer) LuaType() uint8 {
	return LUA_INTEGER
}

type Real float64

func (v Real) LuaType() uint8 {
	return LUA_REAL
}

type String string

func (v String) LuaType() uint8 {
	return LUA_STRING
}

type Table struct {
	Array []Value
	Hash  map[Value]Value
}

func (v Table) LuaType() uint8 {
	return LUA_TABLE
}
