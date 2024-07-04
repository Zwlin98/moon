# Moon —— a sidecar for Skynet

## Feature

1. 与 Skynet Cluster 互相调用，支持 Skynet Cluster 的 RPC 格式，无缝接入 Skynet 的 Cluster 模式 
2. 不侵入 Skynet ，避免了使用 C 对 Skynet 进行扩展时的编程风险
3. 由 Go 语言编写，能方便的集成 Go 语言所带来的各种生态优势
4. API 设计参考 Skynet  Cluster，没有集成难度
5. 简单直观的 Lua 对象封装，可以方便的对 Skynet 集群调用中使用的 Lua 对象进行序列化和反序列化
6. 性能持平 Skynet Cluster 集群方案

## Examples

service 文件夹中出给出了两个简单的示例: `http.go` 和 `example.go`, 展示了如何处理 Skynet 传递来的 Lua 对象。
下面的例子展示了 Skynet 节点调用 Moon 节点上 HTTP 服务(call) 的过程，HTTP 是 Moon 自带的一个用于处理 http 请求的服务,
可以用于替换 Skynet 自带的 http 服务。

### example.go

```go
func main() {
	// initialize services
	httpService := service.NewHttpService()
	pingService := service.NewPingService()

	// initialize cluster
	clusterd := cluster.GetClusterd()
	clusterd.Reload(cluster.DefaultConfig{
		"moon": "0.0.0.0:3345",
	})

	// register services
	clusterd.Register("http", httpService)
	clusterd.Register("ping", pingService)

	// start cluster
	clusterd.Open("moon")

	log.Printf("moon start")

	term := make(chan os.Signal, 1)

	signal.Notify(term, os.Interrupt)

	<-term
}
```

### moon.lua

```lua
local function moonHttp(url, opts)
	opts = opts or {}
	opts.method = opts.method or "GET"
	opts.headers = opts.headers or {}
	opts.body = opts.body or ""

	opts.noBody = opts.noBody or false
	if opts.noHeader == nil then
		opts.noHeader = true
	end
	local ok, msg, code, resp = pcall(cluster.call, "moon", "http", "request", url, opts)
	if ok then
		return msg, code, resp
	else
		return false, msg
	end
end

skynet.start(function()
	cluster.reload({
		moon = "127.0.0.1:3345",
	})

	local ok, code, resp = moonHttp("https://www..com", { method = "GET" })

	if ok and resp then
		skynet.error(resp.body)
	else
		skynet.error(code)
	end
end)

```

### config.moon

```lua
thread = 8
logger = nil
harbor = 0
start = "moon"
bootstrap = "snlua bootstrap"	-- The service for bootstrap
luaservice = "./service/?.lua;./test/?.lua;./examples/?.lua"
lualoader = "lualib/loader.lua"
cpath = "./cservice/?.so"
snax = "./test/?.lua"
```
