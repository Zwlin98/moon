package service

import (
	"moon/lua"
)

type Service interface {
	Execute([]lua.Value) ([]lua.Value, error)
}

type LuaFunction func([]lua.Value) ([]lua.Value, error)
