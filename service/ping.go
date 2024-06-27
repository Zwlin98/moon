package service

import (
	"fmt"

	"github.com/Zwlin98/moon/lua"
)

type PingService struct{}

func NewPingService() Service {
	return &PingService{}
}

func (s *PingService) Execute(args []lua.Value) (ret []lua.Value, err error) {
	defer func() {
		if err := recover(); err != nil {
			ret = []lua.Value{lua.String("error panic")}
			err = fmt.Errorf("panic: %v", err)
		}
	}()

	if len(args) < 1 {
		return []lua.Value{lua.String("error args")}, nil
	}

	method, ok := args[0].(lua.String)
	if !ok {
		return []lua.Value{lua.String("error args")}, nil
	}

	if method == "ping" {
		return []lua.Value{lua.String("pong")}, nil
	}

	return []lua.Value{lua.String("error method")}, nil
}
