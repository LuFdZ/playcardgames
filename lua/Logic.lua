require "lua/CardGroup"
require "lua/Card"
require "lua/Tools"
--require "lua/json"
--json = require("cjson")

Logic = {cards = {}, cardNum = 54, existKing = false}
Logic.__index = Logic
GroupType = {}
--表示牌的KEY
CardValueData = {"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A","B", "R"}
--牌的值
CardValue = {}

function Logic:new()
    self = {}
    math.randomseed(os.time())
    setmetatable(self, Logic)
    --创建牌组型枚举
    GroupType = CreateEnumTable(GroupTypeName, -1)

    --创建牌型枚举
    CardValue = CreateEnumTable(CardValueData, 1)
    self:ReSet()
    self:InitCards()
    return self
end

function Logic:ReSet()
    self.cardNum = 54
end

--初始化牌
function Logic:InitCards()
    local i = 1
    for t = 1, 4 do
        for v = 2, 15 do
            local card = Card:new(t, v)
            table.insert(self.cards, i, card)
            i = i + 1
        end
    end
end

--取牌，返回13张牌
function Logic:GetCards()
    local cardvec = {}

    if #self.cards < 13 then
        print("没有那么多牌了")
        return cardvec
    end

    for i = 1, 13 do
        local rn = math.random(self.cardNum)
        --如果不存在大小王，而且随机到的不是大小王，添加，或者存在大小王，添加
        if (not self.existKing and self.cards[rn]._value ~= 20 and self.cards[rn]._value ~= 21) or (self.existKing) then
            table.insert(cardvec, i, self.cards[rn])
            self.cards[rn], self.cards[self.cardNum] = self.cards[self.cardNum], self.cards[rn]
            self.cardNum = self.cardNum - 1
        end
    end


    return cardvec
end


--比牌，group1大于group2时返回1，否则返回-1，相等返回0， 出错返回-2
function Logic:CompareGroup(group1, group2)
    if #group1 ~= #group2 then
        print("比牌错误，牌的数量不同！")

        return -2
    end

    if group1._type > group2._type then
        return 1
    elseif group1._type < group2._type then
        return -1
    end

    --如果类型相等，则比关键牌
    --顺子，单牌，同花，同花顺，比最大的一张牌
    if group1._type == GroupType.Single or
        group1._type == GroupType.Straight or
        group1._type == GroupType.Flush or
    group1._type == GroupType.FlushStright then

        local res = CompareCard(group1._cards[1], group2._cards[1])
        return res
    end

    --对子，三条，四条，葫芦， 两对比关键牌
    for i = 1, group1._keyValue do
      local res = CompareCard(group1._keyValue[i], group2._keyValue[i])
      if res ~= 0 then
        return res
        end
    end

    return 0
end

-- data结构
-- data =
-- {
--     {ID = "", Head = {"", "", ""}, Middle = {"", "", "","",""}, Tail = {"", "", "", "", ""}},
--     {ID = "", Head = {"", "", ""}, Middle = {"", "", "","",""}, Tail = {"", "", "", "", ""}},
--     {ID = "", Head = {"", "", ""}, Middle = {"", "", "","",""}, Tail = {"", "", "", "", ""}},
-- }

--返回：

-- --根据玩家的牌，取得结果
function Logic:GetResult( data )
    --json.decode(data)
    -- body
    --先把接到的数据转成所需要的数据格子
    -- local datas = {}
    -- for key, value in pairs(data) do

    --     datatemp.ID = value.ID
    --     local headCards = {}
    --     local headCardinfos = String2Cards(value.Head)
    --     datatemp.Head = self:TempCards2Cards(headCardinfos)

    --     local middleCardinfos = String2Cards(value.Head)
    --     datatemp.Middle = self:TempCards2Cards(middleCardinfos)

    --     local tailCardinfos = String2Cards(value.Head)
    --     datatemp.Tail = self:TempCards2Cards(tailCardinfos)

    --     table.insert(datas, datatemp)
    -- end

    -- --取一个和其他的做对比
    -- for key, value in pairs(datas) do
    --     for key1, value1 in pairs(datas) do
    --         --value和temp对比
    --         if value.ID ~= value1.ID then

    --         end
    --     end
    -- end

   local result =
    {
        RoomID = 31,
        ResultList = {{   UserID = 100001,
            Settle = {Head = 2, Middle = 3, Tail = 4, AddScore = 9, TotalScore= 9},
            Result = {Head = {"2_3", "3_3", "4_4"}, Middle = {"1_7", "2_7", "4_1","3_5","4_5"}, Tail = {"2_1", "3_2", "2_9", "1_8", "1_4"}}
        },
        {   ID = 100000,
            Settle = {Head = 5, Middle = 6, Tail = 1, AddScore = 3, TotalScore= 8},
            Result = {Head = {"2_3", "3_3", "4_4"},Middle = {"1_7", "2_7", "4_1","3_5","4_5"}, Tail = {"2_1", "3_2", "2_9", "1_8", "1_4"}}
        }}
     }
    return result
end

function Logic:GetGroup( cards )
    -- body
    table.sort(cards,
        function(card1, card2)
            local res = CompareCard(card1, card2)
            if res == 1 then
                return true
            else
                return false
            end
        end)
end

function Logic:TempCards2Cards( tempcards)
    -- body
    local cards = {}
    for i, cardinfo in pairs(tempcards) do
        local card = Card.new(cardinfo._type, cardinfo._value)
        table.insert(cards, card)
    end

    return cards
end

--在组合中取出指定的类型牌组，找到返回牌组，否则返回nil
function Logic:GetGroupInCombs( combs, type )
    -- body

    local groups = {}
    for key, cards in pairs(combs) do

        table.sort(cards, function(card1, card2)
                local res = CompareCard(card1, card2)
                if res == 1 then
                    return true
                else
                    return false
                end
            end)
        if type == GroupType.TwoCouple then
            local res, keyCard = self:IsTwoCouple(cards)
            if res then
                local group = CardGroup:new(cards, GroupType.TwoCouple, keyCard)
                table.insert(groups, group)
            end
        elseif type == GroupType.Couple then
            local res, keyCard = self:IsCouple(cards)
            if res then
                local group = CardGroup:new(cards, GroupType.Couple, keyCard)

                table.insert(groups, group)
            end
        elseif type == GroupType.Four then
            local res, keyCard = self:IsFour(cards)
            if res then
                local group = CardGroup:new(cards, GroupType.Four, keyCard)
                table.insert(groups, group)
            end
        elseif type == GroupType.Three then
            local res, keyCard = self:IsThree(cards)
            if res then
                local group = CardGroup:new(cards, GroupType.Three, keyCard)
                table.insert(groups, group)
            end
        elseif type == GroupType.FlushStright then
            local res, keyCard = self:IsStraight(cards)
            local res2, keyCard2 = self:AllTheSameType(cards)
            if res and res2 then
                local group = CardGroup:new(cards, GroupType.FlushStright, keyCard)
                table.insert(groups, group)
            end
        elseif type == GroupType.Straight then
            local res, keyCard = self:IsStraight(cards)
            if res then
                local group = CardGroup:new(cards, GroupType.Straight, keyCard)
                table.insert(groups, group)
            end
        elseif type == GroupType.Flush then
            local res, keyCard = self:AllTheSameType(cards)
            if res then
                local group = CardGroup:new(cards, GroupType.Flush, keyCard)
                table.insert(groups, group)
            end
        elseif type == GroupType.ThreeCouple then
            local res, keyCard = self:IsThreeWithCouple(cards)
            if res then
                local group = CardGroup:new(cards, GroupType.ThreeCouple, keyCard)
                table.insert(groups, group)
            end
        end
    end

    return groups
end

--取出所有的牌组
function Logic:CalculateAllGroupTypes(cards)

    --没有牌，直接返回nil
    self.groups = {}
    if #cards == 0 then
        return nil
    end

    local num = 0
    local res = {}
    local combs = {}

    --任意数量牌中取出两张，检查是否对子
    self:Combine_increase(combs, cards, res, 1, 2, 2)
    self:GetGroupInCombs(combs, GroupType.Couple)
    --检查两对
    combs = {}
    self:Combine_increase(combs, cards, res, 1, 4, 4)
    self:GetGroupInCombs(combs, GroupType.TwoCouple)
    --检查线三
    combs = {}
    self:Combine_increase(combs, cards, res, 1, 3, 3)
    self:GetGroupInCombs(combs, GroupType.Three)
    --检查四条
    combs = {}
    self:Combine_increase(combs, cards, res, 1, 4, 4)
    self:GetGroupInCombs(combs, GroupType.Four)
    --检查五张的牌组
    combs = {}
    self:Combine_increase(combs, cards, res, 1, 5, 5)
    --顺子
    self:GetGroupInCombs(combs, GroupType.Straight)
    --同花
    self:GetGroupInCombs(combs, GroupType.Flush)
    --同花顺子
    self:GetGroupInCombs(combs, GroupType.FlushStright)
    --葫芦
    self:GetGroupInCombs(combs, GroupType.ThreeCouple)


end

--查找所有组合,combs:结果， cards:所有的牌， res：一组结果的索引, start:开始的索引, count:查找到第几个, num：组合中元素个数
function Logic:Combine_increase(combs, cards, res, start, count, num)
    local i = 1
    if cards == nil then
        return
    end

    for i = start, #cards + 1 - count do
        res[count] = i
        if count == 1 then
            card = {}
            for j = num, 1, -1 do
                card[j] = cards[res[j]]
            end
            table.insert(combs, card)
        else
            self:Combine_increase(combs, cards, res, i + 1, count - 1, num)
        end
    end
end

--判断是否三条
function Logic:IsThree(cards)
    if #cards ~= 3 then
        return false, 0
    end

    local number = self:EqualsNumber(cards)
    if number == 3 then
        return true, cards[1]._value
    end

    return false, temp
end

--判断是否一对
function Logic:IsCouple(cards)

    --对子
    if #cards ~= 2 then
        return false, 0
    end

    if self:EqualsNumber(cards) == 2 then
        return true, cards[1]._value
    end

    return false, 0
end

--判断是否四条带一个
function Logic:IsFour(cards)
    if #cards ~= 4 then
        return false, 0
    end

    local number = self:EqualsNumber(cards)
    local temp = 0
    if number == 4 then
        temp = cards[1]._value
        return true, temp
    end

    return false, temp
end

--判断有多少个牌相等
function Logic:EqualsNumber(cards)
    local number = 0
    local temp = cards[1]
    for k, v in pairs(cards) do
        if temp._value == v._value then
            number = number + 1
        end
    end

    return number
end

--判断是否所有的类型都相等
function Logic:AllTheSameType(cards)
    if #cards ~= 5 then
        return false, 0
    end

    local temp = cards[1]
    for k, v in pairs(cards) do
        if temp._type ~= v._type then
            return false, 0
        end
    end


    return true, temp._value
end

--判断是否顺子
function Logic:IsStraight(cards)
    if #cards ~= 5 then
        return false, 0
    end

    local temp = cards[1]


    for i = 2, #cards do
        if temp._value - cards[i]._value ~= 1 then
            return false, 0
        else
            temp = cards[i]
        end
    end

    return true, cards[1]._value
end

--是否葫芦
function Logic:IsThreeWithCouple(cards)
    if #cards ~= 5 then
        return false, 0
    end


    --检查三张里面是否有三张相同，如果没有，返回false，如果有，再检查余下的两张是否相同
    local num = 0
    local res = {}
    local combs = {}
    self:Combine_increase(combs, cards, res, 1, 3, 3)
    local temp = 0
    local three = {}
    for k, v in pairs(combs) do

        if self:EqualsNumber(v) == 3 then
            table.insert(three, v)
            temp = three[1]._value
            break
        end
    end

    --判断有三张
    if #three == 0 then
        return false, 0
    end

    local tempcards = {}

    for k, v in pairs(cards) do
        local same = false
        --检查牌是否在三张牌里面
        for k1, v1 in pairs(three[1]) do
            if v._type == v1._type and v._value == v1._value then
                same = true
            end
            --print("v: ", v._type, v._value, "v1: ", v1._type, v1._value)
        end
        --如果不在，那么把这张牌加到列表里面
        if not same then
            --print("not same : ", v._type, v._value)
            table.insert(tempcards, v)
        end
    end

    --判断有对
    if self:EqualsNumber(tempcards) == 2 then
        return true, temp
    end

    -- print("........................................................>> 没有对子 tempcards:")
    -- for k, v in pairs(tempcards) do
    --     print("card ", v._type , v._value )
    -- end
    return false, 0
end

--判断是否两对
function Logic:IsTwoCouple(cards)
    if #cards ~= 4 then
        return false, 0
    end

    local num = 0
    local res = {}
    local combs = {}
    self:Combine_increase(combs, cards, res, 1, 2, 2)
    local tempCombs = {}
    for k, v in pairs(combs) do

        if self:EqualsNumber(v) == 2 then
            table.insert(tempCombs, v)
        end
    end

    if #tempCombs == 2 then
        if CompareCard(tempCombs[1][1], tempCombs[2][1]) == 1 then
            return true, tempCombs[1][1]._value
        else
            return true, tempCombs[2][1]._value
        end
    end

    return false, 0
end
