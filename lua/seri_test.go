package lua

import (
	"bytes"
	"testing"
)

func TestSerilizeIntegerZero(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	serilizeInteger(buf, 0)
	packed := buf.Bytes()
	unpacked, err := Deserialize(packed)
	if err != nil {
		t.Errorf("deserialize failed: %v", err)
	}
	if unpacked[0] != Integer(0) {
		t.Errorf("serilize integer failed")
	}
}

func TestSerilizeIntegerBig(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	serilizeInteger(buf, 0x7FFFFFFFFF)
	packed := buf.Bytes()
	unpacked, err := Deserialize(packed)
	if err != nil {
		t.Errorf("deserialize failed: %v", err)
	}
	if unpacked[0] != Integer(0x7FFFFFFFFF) {
		t.Errorf("serilize integer failed")
	}
}

func TestSerilizeNil(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	serilizeNil(buf)
	packed := buf.Bytes()
	unpacked, err := Deserialize(packed)
	if err != nil {
		t.Errorf("deserialize failed: %v", err)
	}
	if unpacked[0].LuaType() != LUA_NIL {
		t.Errorf("serilize nil failed")
	}
}

func TestSerilizeBoolean(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	serilizeBoolean(buf, true)
	packed := buf.Bytes()
	unpacked, err := Deserialize(packed)
	if err != nil {
		t.Errorf("deserialize failed: %v", err)
	}
	if unpacked[0] != Boolean(true) {
		t.Errorf("serilize boolean failed")
	}
	buf = bytes.NewBuffer(nil)
	serilizeBoolean(buf, false)
	packed = buf.Bytes()
	unpacked, err = Deserialize(packed)
	if err != nil {
		t.Errorf("deserialize failed: %v", err)
	}
	if unpacked[0] != Boolean(false) {
		t.Errorf("serilize boolean failed")
	}
}

func TestSerilizeReal(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	serilizeReal(buf, 3.1415926)
	packed := buf.Bytes()
	unpacked, err := Deserialize(packed)
	if err != nil {
		t.Errorf("deserialize failed: %v", err)
	}
	if unpacked[0] != Real(3.1415926) {
		t.Errorf("serilize real failed")
	}
}

func TestSerilizeShortString(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	serilizeString(buf, "hello")
	packed := buf.Bytes()
	unpacked, err := Deserialize(packed)
	if err != nil {
		t.Errorf("deserialize failed: %v", err)
	}
	if unpacked[0] != String("hello") {
		t.Errorf("serilize short string failed")
	}
}

func TestSerilizeLongString(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	serilizeString(buf, "hellohellohellohellohellohellohellohello")
	packed := buf.Bytes()
	unpacked, err := Deserialize(packed)
	if err != nil {
		t.Errorf("deserialize failed: %v", err)
	}
	if unpacked[0] != String("hellohellohellohellohellohellohellohello") {
		t.Errorf("serilize long string failed")
	}
}

func TestSerilizeTableArray(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	arr := Table{
		Array: []Value{
			Integer(1000000001),
			String("username"),
			Real(3.1415926),
		},
	}
	err := serilizeTable(buf, arr, 0)
	if err != nil {
		t.Errorf("serilize table failed: %v", err)
	}
	packed := buf.Bytes()
	unpacked, err := Deserialize(packed)
	if err != nil {
		t.Errorf("deserialize failed: %v", err)
	}
	unpackedTable := unpacked[0].(Table)
	for i, v := range arr.Array {
		if v != unpackedTable.Array[i] {
			t.Errorf("value not match: %v != %v", v, unpackedTable.Array[i])
		}
	}
}

func TestSerilizeTableHash(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	table := Table{
		Hash: map[Value]Value{
			Integer(1000000001): String("uid"),
			String("title"):     Integer(55),
			String("isOK"):      Boolean(true),
			String("msg"):       String("hello world"),
		},
	}
	err := serilizeTable(buf, table, 0)
	if err != nil {
		t.Errorf("serilize table failed: %v", err)
	}
	packed := buf.Bytes()
	unpacked, err := Deserialize(packed)
	if err != nil {
		t.Errorf("deserialize failed: %v", err)
	}
	unpackedTable := unpacked[0].(Table)
	for k, v := range table.Hash {
		unpackedTableValue, ok := unpackedTable.Hash[k]
		if !ok {
			t.Errorf("key not found: %v", k)
		}
		if v != unpackedTableValue {
			t.Errorf("value not match: %v != %v", v, unpackedTableValue)
		}
	}
}

func TestSerilizeMixedTable(t *testing.T) {
	tableMixed := Table{
		Array: []Value{
			Integer(1000000001),
			String("username"),
			Real(3.1415926),
		},
		Hash: map[Value]Value{
			Integer(1000000001): String("uid"),
			String("title"):     Integer(55),
			String("isOK"):      Boolean(true),
			String("msg"):       String("hello world"),
		},
	}
	buf := bytes.NewBuffer(nil)
	err := serilizeTable(buf, tableMixed, 0)
	if err != nil {
		t.Errorf("serilize table failed: %v", err)
	}
	packed := buf.Bytes()
	unpacked, err := Deserialize(packed)
	if err != nil {
		t.Errorf("deserialize failed: %v", err)
	}
	unpackedTable := unpacked[0].(Table)
	for i, v := range tableMixed.Array {
		if v != unpackedTable.Array[i] {
			t.Errorf("value not match: %v != %v", v, unpackedTable.Array[i])
		}
	}
	for k, v := range tableMixed.Hash {
		unpackedTableValue, ok := unpackedTable.Hash[k]
		if !ok {
			t.Errorf("key not found: %v", k)
		}
		if v != unpackedTableValue {
			t.Errorf("value not match: %v != %v", v, unpackedTableValue)
		}
	}
}
