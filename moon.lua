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
