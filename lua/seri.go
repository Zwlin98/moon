package lua

// Lua 对象在 Go 中的简单表示，用于序列化和反序列化
// 1. 不支持元表的序列化/反序列化
// 2. 不支持函数, 线程, 用户数据, 轻量用户数据的序列化/反序列化
//    (LUA_TFUNCTION, LUA_TTHREAD, LUA_TUSERDATA, LUA_TLIGHTUSERDATA)
// 3. 支持纯 Lua Table 数组与 Go Slice 的序列化/反序列化
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const MAX_COOKIE = 32

func combineType(t uint8, v uint8) uint8 {
	return t | (v << 3)
}

func Serialize(values []Value) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	for _, v := range values {
		if err := serilizeLuaValue(buf, v, 0); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func Deserialize(data []byte) ([]Value, error) {
	buf := bytes.NewBuffer(data)
	values := make([]Value, 0)
	for {
		value, err := deserilizeLuaValue(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

func serilizeLuaValue(buf *bytes.Buffer, v Value, depth int) error {
	if depth > 32 {
		return fmt.Errorf("serialize can't pack too deep")
	}
	switch v.LuaType() {
	case LUA_NIL:
		serilizeNil(buf)
	case LUA_BOOLEAN:
		serilizeBoolean(buf, v.(Boolean))
	case LUA_INTEGER:
		serilizeInteger(buf, v.(Integer))
	case LUA_REAL:
		serilizeReal(buf, v.(Real))
	case LUA_STRING:
		serilizeString(buf, v.(String))
	case LUA_TABLE:
		err := serilizeTable(buf, v.(Table), depth+1)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported lua value type: %v", v.LuaType())
	}
	return nil
}

func serilizeNil(buf *bytes.Buffer) {
	buf.WriteByte(TYPE_NIL)
}

func serilizeBoolean(buf *bytes.Buffer, b Boolean) {
	if b {
		buf.WriteByte(combineType(TYPE_BOOLEAN, 1))
	} else {
		buf.WriteByte(combineType(TYPE_BOOLEAN, 0))
	}
}

func serilizeReal(buf *bytes.Buffer, v Real) {
	buf.WriteByte(combineType(TYPE_NUMBER, TYPE_NUMBER_REAL))
	binary.Write(buf, binary.LittleEndian, v)
}

func serilizeTableArray(buf *bytes.Buffer, array []Value, depth int) error {
	arrLen := len(array)
	if arrLen >= MAX_COOKIE-1 {
		buf.WriteByte(combineType(TYPE_TABLE, MAX_COOKIE-1))
		serilizeInteger(buf, Integer(arrLen))
	} else {
		buf.WriteByte(combineType(TYPE_TABLE, uint8(arrLen)))
	}
	for _, v := range array {
		if v.LuaType() == LUA_NIL {
			return fmt.Errorf("array value can't be nil")
		}
		err := serilizeLuaValue(buf, v, depth+1)
		if err != nil {
			return err
		}
	}
	return nil
}

func serilizeTable(buf *bytes.Buffer, table Table, depth int) error {
	if table.Array != nil && len(table.Array) > 0 {
		serilizeTableArray(buf, table.Array, depth+1)
	} else {
		buf.WriteByte(combineType(TYPE_TABLE, 0))
	}
	if table.Hash != nil {
		for k, v := range table.Hash {
			if k.LuaType() == LUA_TABLE || k.LuaType() == LUA_NIL {
				return fmt.Errorf("table key can't be table, array or nil")
			}
			if v.LuaType() == LUA_NIL {
				return fmt.Errorf("table value can't be nil")
			}
			err := serilizeLuaValue(buf, k, depth+1)
			if err != nil {
				return err
			}
			err = serilizeLuaValue(buf, v, depth+1)
			if err != nil {
				return err
			}
		}
	}
	serilizeNil(buf) // end of table
	return nil
}

func serilizeString(buf *bytes.Buffer, luaString String) {
	sz := len(luaString)
	if sz < MAX_COOKIE {
		buf.WriteByte(combineType(TYPE_SHORT_STRING, uint8(sz)))
		if sz > 0 {
			buf.Write([]byte(luaString))
		}
		return
	}
	if sz < 0x10000 {
		buf.WriteByte(combineType(TYPE_LONG_STRING, 2))
		binary.Write(buf, binary.LittleEndian, uint16(sz))
	} else {
		buf.WriteByte(combineType(TYPE_LONG_STRING, 4))
		binary.Write(buf, binary.LittleEndian, uint32(sz))
	}
	buf.Write([]byte(luaString))
}

func serilizeInteger(buf *bytes.Buffer, v Integer) {
	if v == 0 {
		buf.WriteByte(combineType(TYPE_NUMBER, TYPE_NUMBER_ZERO))
		return
	}
	if (Integer)(int32(v)) != v {
		buf.WriteByte(combineType(TYPE_NUMBER, TYPE_NUMBER_QWORD))
		binary.Write(buf, binary.LittleEndian, int64(v))
		return
	}
	if v < 0 {
		buf.WriteByte(combineType(TYPE_NUMBER, TYPE_NUMBER_DWORD))
		binary.Write(buf, binary.LittleEndian, int32(v))
		return
	}
	if v < 0x100 {
		buf.WriteByte(combineType(TYPE_NUMBER, TYPE_NUMBER_BYTE))
		buf.WriteByte(uint8(v))
		return
	}
	if v < 0x10000 {
		buf.WriteByte(combineType(TYPE_NUMBER, TYPE_NUMBER_WORD))
		binary.Write(buf, binary.LittleEndian, uint16(v))
		return
	}
	buf.WriteByte(combineType(TYPE_NUMBER, TYPE_NUMBER_DWORD))
	binary.Write(buf, binary.LittleEndian, uint32(v))
}

func deserilizeLuaValue(buf *bytes.Buffer) (Value, error) {
	head, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	typ, cookie := head&0x07, head>>3
	switch typ {
	case TYPE_NIL:
		return Nil{}, nil
	case TYPE_BOOLEAN:
		if cookie == 0 {
			return Boolean(false), nil
		} else {
			return Boolean(true), nil
		}
	case TYPE_NUMBER:
		if cookie == TYPE_NUMBER_REAL {
			return deserilizeReal(buf)
		} else {
			return deserilizeInteger(buf, cookie)
		}
	case TYPE_USERDATA:
		var pointer uint64 // skip
		err := binary.Read(buf, binary.LittleEndian, &pointer)
		return Nil{}, err
	case TYPE_SHORT_STRING:
		b := make([]byte, cookie)
		_, err := buf.Read(b)
		return String(string(b)), err
	case TYPE_LONG_STRING:
		return deserilizeLongString(buf, cookie)
	case TYPE_TABLE:
		return deserilizeTable(buf, cookie)
	default:
		return nil, fmt.Errorf("unsupported lua value type: %v", typ)
	}
}

func deserilizeTable(buf *bytes.Buffer, arrLen uint8) (Value, error) {
	var arrSize Integer
	if arrLen == MAX_COOKIE-1 {
		var head uint8
		err := binary.Read(buf, binary.LittleEndian, &head)
		if err != nil {
			return nil, err
		}
		typ, cookie := head&0x07, head>>3
		if typ != TYPE_NUMBER || cookie == TYPE_NUMBER_REAL {
			return nil, fmt.Errorf("unsupported table cookie: %v", head)
		}
		v, err := deserilizeInteger(buf, cookie)
		if err != nil {
			return nil, err
		}
		arrSize = v.(Integer)
	} else {
		arrSize = Integer(arrLen)
	}
	table := Table{
		Array: make([]Value, arrSize),
		Hash:  make(map[Value]Value),
	}
	for i := 0; i < int(arrSize); i++ {
		v, err := deserilizeLuaValue(buf)
		if err != nil {
			return nil, err
		}
		table.Array[i] = v
	}
	for {
		key, err := deserilizeLuaValue(buf)
		if err != nil {
			return nil, err
		}
		if key.LuaType() == LUA_NIL {
			break
		}
		value, err := deserilizeLuaValue(buf)
		if err != nil {
			return nil, err
		}
		table.Hash[key] = value

	}
	return table, nil
}

func deserilizeLongString(buf *bytes.Buffer, cookie uint8) (Value, error) {
	if cookie == 2 {
		var sz uint16
		err := binary.Read(buf, binary.LittleEndian, &sz)
		if err != nil {
			return nil, err
		}
		b := make([]byte, sz)
		_, err = buf.Read(b)
		return String(string(b)), err
	}
	if cookie == 4 {
		var sz uint32
		err := binary.Read(buf, binary.LittleEndian, &sz)
		if err != nil {
			return nil, err
		}
		b := make([]byte, sz)
		_, err = buf.Read(b)
		return String(string(b)), err
	}
	return nil, fmt.Errorf("unsupported long string cookie: %v", cookie)
}

func deserilizeReal(buf *bytes.Buffer) (Value, error) {
	var v float64
	err := binary.Read(buf, binary.LittleEndian, &v)
	return Real(v), err
}

func deserilizeInteger(buf *bytes.Buffer, cookie uint8) (Value, error) {
	switch cookie {
	case TYPE_NUMBER_ZERO:
		return Integer(0), nil
	case TYPE_NUMBER_BYTE:
		var v uint8
		err := binary.Read(buf, binary.LittleEndian, &v)
		return Integer(v), err
	case TYPE_NUMBER_WORD:
		var v uint16
		err := binary.Read(buf, binary.LittleEndian, &v)
		return Integer(v), err
	case TYPE_NUMBER_DWORD:
		var v int32
		err := binary.Read(buf, binary.LittleEndian, &v)
		return Integer(v), err
	case TYPE_NUMBER_QWORD:
		var v int64
		err := binary.Read(buf, binary.LittleEndian, &v)
		return Integer(v), err
	default:
		return nil, fmt.Errorf("unsupported integer cookie: %v", cookie)
	}
}
