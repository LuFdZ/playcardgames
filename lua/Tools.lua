--
-- Author: KGD
-- Date: 2017-08-13 23:36:02
--
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
    local cardstr = ""
    for k, v in pairs(cards ) do
        local card = v._type .. "_" .. v._value
        table.insert(cardstr, card)
    end
    
    return cardstr
end
