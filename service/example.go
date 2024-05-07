package service

import (
	"fmt"

	"github.com/Zwlin98/moon/lua"
)

type ExampleService struct{}

func NewExampleService() Service {
	return &ExampleService{}
}

func (s *ExampleService) Execute(args []lua.Value) ([]lua.Value, error) {
	fmt.Printf("ExampleService.Execute called with args: %v\n", args)

	return []lua.Value{
		lua.Boolean(true),
		lua.String("hello world"),
		lua.Table{
			Array: []lua.Value{
				lua.Integer(1),
				lua.Real(3.14),
				lua.String("hello"),
				lua.Boolean(true),
			},
			Hash: map[lua.Value]lua.Value{
				lua.Boolean(true):    lua.String("true"),
				lua.Boolean(false):   lua.String("false"),
				lua.Integer(100):     lua.String("hello world"),
				lua.String("number"): lua.Integer(200),
				lua.String("string"): lua.Boolean(true),
				lua.String("msg"):    lua.String("hello world"),
			},
		},
	}, nil
}
