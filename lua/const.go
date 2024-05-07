package lua

// For serialization and deserialization
const (
	TYPE_NIL          uint8 = 0
	TYPE_BOOLEAN      uint8 = 1
	TYPE_NUMBER       uint8 = 2
	TYPE_USERDATA     uint8 = 3
	TYPE_SHORT_STRING uint8 = 4
	TYPE_LONG_STRING  uint8 = 5
	TYPE_TABLE        uint8 = 6
)

// For serialization and deserialization
const (
	TYPE_NUMBER_ZERO  uint8 = 0 // 0
	TYPE_NUMBER_BYTE  uint8 = 1 // 8 bit
	TYPE_NUMBER_WORD  uint8 = 2 // 16 bit
	TYPE_NUMBER_DWORD uint8 = 4 // 32 bit
	TYPE_NUMBER_QWORD uint8 = 6 // 64 bit

	TYPE_NUMBER_REAL uint8 = 8 // real
)

const (
	LUA_NIL     uint8 = 0
	LUA_BOOLEAN uint8 = 1
	LUA_INTEGER uint8 = 2
	LUA_REAL    uint8 = 3
	LUA_STRING  uint8 = 4
	LUA_TABLE   uint8 = 5
)
