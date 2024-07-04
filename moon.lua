local skynet = require("skynet")
local cluster = require("skynet.cluster")

---@class HttpOpts
---@field method string
---@field headers table | nil
---@field body string | nil
---@field noBody boolean | nil
---@field noHeader boolean | nil

---@class HttpResponse
---@field headers table
---@field body string

---comment
---@param url string
---@param opts HttpOpts
---@return boolean ok
---@return number code
---@return HttpResponse | nil response
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
