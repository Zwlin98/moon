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

### main.go

```go
func main() {
	// initialize services
	httpService := service.NewHttpService()
	exampleService := service.NewExampleService()

	// initialize cluster
	clusterd := cluster.GetClusterd()
	clusterd.Reload(cluster.DefaultConfig{
		"moon": "127.0.0.1:3345",
		"db":   "127.0.0.1:2528",
	})

	// register services
	clusterd.Register("http", httpService)
	clusterd.Register("example", exampleService)

	// start cluster
	clusterd.Open("moon")
 	
  select{}
}
```

### moon.lua

```lua
local skynet = require("skynet")
local cluster = require("skynet.cluster")

skynet.start(function()
	cluster.reload({
		db = "127.0.0.1:2528",
		moon = "127.0.0.1:3345",
	})

	local sdb = skynet.newservice("simpledb")
	skynet.call(sdb, "lua", "SET", "ping", "pong")

	cluster.register("sdb", sdb)

	cluster.open("db")

	local ok, s, t = cluster.call("moon", "example", "CMD", "arg1", { a = 1, b = 2, c = 3 })
	if ok then
		skynet.error("call moon example CMD success")
		skynet.error(s)
		for k, v in pairs(t) do
			skynet.error(k, v)
		end
	else
		skynet.error("call moon example CMD failed")
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
