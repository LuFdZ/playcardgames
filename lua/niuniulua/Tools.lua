--
-- Author: KGD
-- Date: 2017-08-13 23:36:02
--
package.path = os.getenv("PWD") .. '/?.lua;'
require("lua/niuniulua/functions")

function CreateEnumTable(tbl, index)
    local enumtbl = {}
    local enumindex = index or 0
    for i, v in ipairs(tbl) do
        enumtbl[v] = enumindex + i
    end
    return enumtbl
end

function split(s, delim)
    if type(delim) ~= "string" or string.len(delim) <= 0 then
        return
    end

    local start = 1
    local t = {}
    while true do
    local pos = string.find (s, delim, start, true) -- plain find
        if not pos then
          break
        end

        table.insert (t, string.sub (s, start, pos - 1))
        start = pos + string.len (delim)
    end
    table.insert (t, string.sub (s, start))

    return t
end

--取得表的key
function getTableKey( tb, value )
    -- body
    for k, v in pairs(tb) do
        if v == value then
            return k
        end
    end

    return nil
end

--删除表中的值，并返回删除后的表
function removeTableValue( tb, value )
    -- body
    local key = getTableKey(tb, value)
    if key ~= nil then
        table.remove(tb, key)
    end

    return tb
end

--字符串转牌
function String2Cards( str )
    -- body
    local cards = {}
    for k, v in pairs(str) do
        local arr = split(v, "_")
        local card = {_type = tonumber(arr[1]), _value = tonumber(arr[2])}
        table.insert(cards, card)
    end
    return cards
end

--牌转字符串
function Cards2String( cards )
    -- body
    local cardstr = {}
    for k, v in pairs(cards ) do
        local card = v._type .. "_" .. v._value
        table.insert(cardstr, card)
    end

    return cardstr
end

function dump(value, desciption, nesting)
    if type(nesting) ~= "number" then nesting = 3 end

    local lookupTable = {}
    local result = {}

    local function _v(v)
        if type(v) == "string" then
            v = "\"" .. v .. "\""
        end
        return tostring(v)
    end

    local traceback = string.split(debug.traceback("", 2), "\n")
    print("dump from: " .. string.trim(traceback[3]))

    local function _dump(value, desciption, indent, nest, keylen)
        desciption = desciption or "<var>"
        local spc = ""
        if type(keylen) == "number" then
            spc = string.rep(" ", keylen - string.len(_v(desciption)))
        end
        if type(value) ~= "table" then
            result[#result +1 ] = string.format("%s%s%s = %s", indent, _v(desciption), spc, _v(value))
        elseif lookupTable[value] then
            result[#result +1 ] = string.format("%s%s%s = *REF*", indent, desciption, spc)
        else
            lookupTable[value] = true
            if nest > nesting then
                result[#result +1 ] = string.format("%s%s = *MAX NESTING*", indent, desciption)
            else
                result[#result +1 ] = string.format("%s%s = {", indent, _v(desciption))
                local indent2 = indent.."    "
                local keys = {}
                local keylen = 0
                local values = {}
                for k, v in pairs(value) do
                    keys[#keys + 1] = k
                    local vk = _v(k)
                    local vkl = string.len(vk)
                    if vkl > keylen then keylen = vkl end
                    values[k] = v
                end
                table.sort(keys, function(a, b)
                    if type(a) == "number" and type(b) == "number" then
                        return a < b
                    else
                        return tostring(a) < tostring(b)
                    end
                end)
                for i, k in ipairs(keys) do
                    _dump(values[k], k, indent2, nest + 1, keylen)
                end
                result[#result +1] = string.format("%s}", indent)
            end
        end
    end
    _dump(value, desciption, "- ", 1)

    for i, line in ipairs(result) do
        print(line)
    end
end


--把table2的内容加到table1里面
function TableInsert2Table(table1, table2)
    if table2 == nil then
        return table1
    end

    for k, v in pairs(table2) do
        table.insert(table1, v)
    end
    return table1
end

--获取无序table的长度
function GetTableLen(tb)
    local count = 0
    for k, v in pairs(tb) do
        count = count + 1
    end
    return count
end
