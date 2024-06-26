package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Zwlin98/moon/lua"
)

type request struct {
	method   string
	headers  map[string]string
	body     string
	noBody   bool
	noHeader bool
}

type httpFunc func(url string, req request) ([]lua.Value, error)

type HttpService struct{}

func NewHttpService() Service {
	return &HttpService{}
}

func (s *HttpService) Execute(args []lua.Value) (ret []lua.Value, err error) {
	defer func() {
		if err := recover(); err != nil {
			ret = buildError(fmt.Sprintf("http service panic: %v", err))
			err = nil
		}
	}()

	// 参数检查和处理
	if len(args) < 3 {
		return buildError("args not enough"), nil
	}
	cmd, ok := args[0].(lua.String)
	if !ok || cmd != "request" {
		return buildError("command not found"), nil
	}

	url, ok := args[1].(lua.String)
	if !ok {
		return buildError("url parse error"), nil
	}

	opts, ok := args[2].(lua.Table)
	if !ok {
		return buildError("opts parse error"), nil
	}

	method := opts.Hash[lua.String("method")].(lua.String)
	headers := opts.Hash[lua.String("headers")].(lua.Table)
	body := opts.Hash[lua.String("body")].(lua.String)
	noBody := opts.Hash[lua.String("noBody")].(lua.Boolean)
	noHeader := opts.Hash[lua.String("noHeader")].(lua.Boolean)

	if !checkMethod(string(method)) {
		return buildError("method not allowed"), nil
	}

	reqHeaders := make(map[string]string)
	for k, v := range headers.Hash {
		reqHeaders[string(k.(lua.String))] = string(v.(lua.String))
	}

	return httpRequest(string(url), request{
		method:   string(method),
		headers:  reqHeaders,
		body:     string(body),
		noBody:   bool(noBody),
		noHeader: bool(noHeader),
	})
}

func checkMethod(method string) bool {
	switch method {
	case "GET", "POST", "PUT", "DELETE", "PATCH":
		return true
	default:
		return false
	}
}

func httpRequest(url string, req request) ([]lua.Value, error) {
	r, err := http.NewRequest(req.method, url, strings.NewReader(req.body))
	if err != nil {
		return buildError(err.Error()), nil
	}
	for k, v := range req.headers {
		r.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return buildError(err.Error()), nil
	}
	return buildResponse(resp, req.noHeader, req.noBody), nil
}

func buildResponse(resp *http.Response, noHeader bool, noBody bool) []lua.Value {

	statusCode := lua.Integer(resp.StatusCode)

	ret := make([]lua.Value, 0, 4)
	ret = append(ret, lua.Boolean(true))
	ret = append(ret, statusCode)

	hash := make(map[lua.Value]lua.Value)
	if !noHeader {
		headers := make(map[lua.Value]lua.Value)
		for k, v := range resp.Header {
			if len(v) == 1 {
				headers[lua.String(k)] = lua.String(v[0])
			} else {
				arr := make([]lua.Value, 0, len(v))
				for _, vv := range v {
					arr = append(arr, lua.String(vv))
				}
				headers[lua.String(k)] = lua.Table{
					Array: arr,
				}
			}
		}
		hash[lua.String("headers")] = lua.Table{
			Hash: headers,
		}
	}

	if !noBody {
		defer resp.Body.Close()
		buf := bytes.NewBuffer(nil)
		io.Copy(buf, resp.Body)
		hash[lua.String("body")] = lua.String(buf.String())
	}

	ret = append(ret, lua.Table{Hash: hash})

	return ret
}

func buildError(msg string) []lua.Value {
	return []lua.Value{
		lua.Boolean(false),
		lua.String(msg),
	}
}
