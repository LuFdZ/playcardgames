require "lua/CardGroup"
require "lua/Card"
require "lua/Global"

Logic = {cards = {}, cardNum = 54, existKing = false}
Logic.__index = Logic
GroupType = {}
--表示牌的KEY
CardValueData = {"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A","B", "R"}
--牌的值
CardValue = {}

function Logic:test1234(a, b)
    return a + b
end

function Logic:new()
    self = {}
    math.randomseed(tostring(os.time()):reverse():sub(1, 6))
    setmetatable(self, Logic)
    --创建牌组型枚举
    GroupType = CreateEnumTable(GroupTypeName, -1)
    --创建牌型枚举
    CardValue = CreateEnumTable(CardValueData, 1)

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
        return 0
    end

    --如果类型相等，则比关键牌
    --顺子，单牌，同花，同花顺，比最大的一张牌
    if group1._type == GroupType.Single or
        group1._type == GroupType.Straight or
        group1._type == GroupType.Flush or
    group1._type == GroupType.FlushStright then
        return CompareCard(group1._cards[1], group2._cards[1])
    end

    --对子，三条，四条，葫芦， 两对比关键牌


    return CompareCard(group1._keyCard, group2_keyCard)
end


--取出所有的牌组
function Logic:CalculateAllGroupTypes(cards)
    --没有牌，直接返回nil
    if #cards == 0 then
        return nil
    end

    local num = 0
    local res = {}
    local combs = {}

    local groups = {}
    --先判断牌如果只有三张
    if #cards == 3 then

        --在三张牌中选两张，看看是否有对
        self:Combine_increase(combs, self._cards, res, 1, 2, 2)
        for k, v in pairs(combs) do
            if self:IsCouple(v) then
                local group = CardGroup:new(v, CardType.Couple)
                table.insert(groups, group)

            end
        end

        --判断是否三张
        if self:IsThree(cards) then
            local group = CardGroup:new(v, CardType.Three)
            table.insert(groups, group)

        end
    end

    combs = {}
    --牌大于五张
    --从任意牌中随意选出五张
    self:Combine_increase(combs, self._cards, res, 1, 5, 5)

    for k, v in pairs(combs) do
        --先从大到小排序
        table.sort(v, function(c1, c2)
                       return CompareCard(c1, c2)
        end)
        repeat
            local res, key = self:IsTwoCouple(v)
            if res then    --两对
                local group = CardGroup:new(v, CardType.TwoCople)
                table.insert(groups, group, key)
                break
            end

            res, key = self:IsCouple(cards)
            if res then --对子
                local group = CardGroup:new(v, CardType.Couple, key)
                table.insert(groups, group)
                break
            end

            res, key = self:IsFour(v)
            if res then  --四条
                local group = CardGroup:new(v, CardType.Four, key)
                table.insert(groups, group)
                break
            end

            res, key = self:IsThree(cards)
            if res then --三条
                local group = CardGroup:new(v, CardType.Three, key)
                table.insert(groups, group)
                break
            end

            res, key = self:IsStraight(v)
            if res and self:AllTheSameType(v) then--同花顺
                local group = CardGroup:new(v, CardType.FlushStright, key)
                table.insert(groups, group)
                break
            end

            res, key = self:IsStraight(v)
            if res then --顺子
                local group = CardGroup:new(v, CardType.Straight, key)
                table.insert(groups, group)
                break
            end

            res, key = self:AllTheSameType(v)
            if res then --同花
                local group = CardGroup:new(v, CardType.Flush, key)
                table.insert(groups, group)
                break
            end

            res, key = self:IsThreeWithCouple(v)
            if res then  --葫芦
                local group = CardGroup:new(v, CardType.ThreeCouple, key)
                table.insert(groups, group)
                break
            end

            --都 不是这几种 牌型 ，那么就是单张
            local group = CardGroup:new(v, CardType.Single, v._cards[1]._value)
            table.insert(groups, group)
            break
        until true
    end

    return groups
end

--查找所有组合,combs:结果， cards:所有的牌， res：一组结果的索引, start:开始的索引, count:查找到第几个, num：组合中元素个数
function Logic:Combine_increase(combs, cards, res, start, count, num)
    local i = 1
    for i = start, #cards + 1 - count do
        res[count] = i
        if (count == 1) then
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

    local num = 0
    local res = {}
    local combs = {}
    self:Combine_increase(combs, self._cards, res, 1, 3, 3)

    --用来保存关键牌
    local temp = 0
    for k, v in pairs(combs) do
        local number = self:EqualsNumber(v)
        if number == 3 then
            num = num + 1
            temp = v[1]._value
        end
    end

    if num == 1 then
        return true, temp
    end
    return false, temp
end

--判断是否一对
function Logic:IsCouple(cards)

    --对子
    local num = 0
    local res = {}
    local combs = {}
    self:Combine_increase(combs, self._cards, res, 1, 2, 2)

    local temp = 0
    for k, v in pairs(combs) do
        local number = self:EqualsNumber(v)
        if number == 2 then
            num = num + 1
            temp = v[1]._value
        end
    end

    if num == 1 then
        return true, temp
    end

    return false, temp
end

--判断是否四条带一个
function Logic:IsFour(cards)
    if #cards ~= 5 then
        return false
    end

    local number = self:EqualsNumber(cards)
    local temp = 0
    if number == 4 then
        temp = v[1]._value
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

    --要排除temp本身，所以要-1
    return number - 1
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

    for i = 1, #cards do
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
    self:Combine_increase(combs, self._cards, res, 1, 3, 3)
    local temp = 0
    local three = {}
    for k, v in pairs(combs) do
        --先从大到小排序
        table.sort(v, function(c1, c2)
                       return CompareCard(c1, c2)
        end)

        if self:IsThree(v) then
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
        for k1, v1 in pairs(three) do
            if v._type ~= v1._type and v._value ~= v1._value then
                table.insert(tempcards, v)
            end
        end
    end

    --判断有对
    if self:IsCouple(tempcards) then
        return true, temp
    end

    return false, 0
end

--判断是否两对
function Logic:IsTwoCouple(cards)
    local num = 0
    local res = {}
    local combs = {}
    self:Combine_increase(combs, self._cards, res, 1, 2, 2)

    if #combs == 2 then
        if CompareCard(combs[1]._cards[1], combs[2]._cards[1]) then
            return true, combs[1]._cards[1]._value
        else
            return true, combs[2]._cards[1]._value
        end
    end

    return false, 0
end
