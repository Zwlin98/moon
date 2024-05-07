package service

import (
	"fmt"
	"moon/lua"
)

type ExampleService struct {
}

func (s *ExampleService) Execute(args []lua.Value) ([]lua.Value, error) {
	fmt.Printf("ExampleService.Execute called with args: %v\n", args)

	return []lua.Value{
		lua.String("ok"),
		lua.String("not ok"),
		lua.String("hello world"),
		lua.String("skynet"),
		lua.Table{
			Array: []lua.Value{
				lua.Integer(1000000001),
				lua.String("username"),
				lua.Real(3.1415926),
			},
			Hash: map[lua.Value]lua.Value{
				lua.Integer(1000000001): lua.String("uid"),
				lua.String("title"):     lua.Integer(55),
				lua.String("isOK"):      lua.Boolean(true),
				lua.String("msg"):       lua.String("hello world"),
			},
		},
	}, nil
}
