
local json = require "cjson"  
local t= '{"name":"sssss"}'
local st = json.decode(t)  
print(st)
