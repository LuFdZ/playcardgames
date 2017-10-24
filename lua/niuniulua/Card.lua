
--表示牌的KEY
CardName = {"A","2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}

Card = {
    --  1,2,3,4 对应：方块，梅花，红心，黑桃
    _type,
    --  牛牛从1开始，13张从2开始,小王为20， 大王为21
    _value
}
Card.__index = Card

function Card:new(t, v)
    self = {}
    setmetatable(self, Card)
    self._type = t
    self._value = v
    return self
end

CardGroup = {
    --  所有的牌
    _cards = {},
    --  赢的倍数
    _winN = 1,
    --  0,1,2,...10,11,12,13对应：无牛，牛1，牛2...牛牛，四炸，五花，五小
    _type = 0,
    --组成牛的三张牌
    _niuCards = {},
}

CardGroup.__index = CardGroup

function CardGroup:new(cards)
    self = {}
    setmetatable(self, CardGroup)

    --  先从大到小排序
    table.sort(cards, function(c1, c2)
        return CompareCard(c1, c2)
    end)

    self._cards = cards
    self:CalculateType()
    return self
end

--计算牌型，module为房间模式为1，2，3整数
function CardGroup:CalculateType()
    local res = 0
    if(self:IsWuXiao())then
        res = 13
    elseif(self:IsWuHua())then
        res = 12
    elseif(self:IsSiZha())then
        res = 11
    else
        res = self:IsNiu()
    end

    self._type = res
end

--  五张牌大于10并且
function CardGroup:IsWuHua()
    for i, v in pairs(self._cards) do
        if (v._value < 11) then
            return false
        end
    end

    --把组成牛的三张牌记录起来
    for i = 1, 3 do
        table.insert(self._niuCards, self._cards[i])
    end
    return true
end

--  五张牌有四张相同
function CardGroup:IsSiZha()
    local tempList = {}
    for k, v in pairs(self._cards) do
        --取得点数和v._value相等的牌
        local tempNumber = self:GetEqualsNumberOfA(self._cards, v._value)
        local temp = {Num = tempNumber, card = v}
        --记录牌的相同数量
        table.insert(tempList, temp)
    end

    table.sort(tempList, function(temp1, temp2)
        if temp1.Num > temp2.Num then
            return true
        else
            return false
        end
    end)

    if #tempList <= 0 then
        MyLog("sb了，数据有错:")
        dump(cards)
    end

    local res = false

    if tempList[1].Num == 4 then
        --把组成牛的三张牌记录起来
        for k, v in pairs(self._cards) do
            if v._value == tempList[1].card._value then
                table.insert(self._niuCards, v)
            end
        end
        res = true
    end

    return res
end

function CardGroup:IsWuXiao()
    local sum = 0
    for i, v in pairs(self._cards) do
        if v._value >= 5 then
            return false
        end
        sum = sum + v._value
    end
    if (sum <= 10) then
        --把组成牛的三张牌记录起来
        for i = 1, 5 do
            table.insert(self._niuCards, self._cards[i])
        end
        return true
    else
        return false
    end
end

function CardGroup:IsNiu()
    local num = 0
    local res = {}
    local combs = {}
    --  从五张牌中取出任意三张，组合
    self:Combine_increase(combs, self._cards, res, 1, 3, 3)

    --  所有的组合，计算三张牌的组合是否有牛
    for i, cards in pairs(combs) do
        num = 0
        for k, c in pairs(cards) do
            if (c._value < 10) then --
                num = num + c._value
            end
        end

        if (num % 10 == 0) then
            --有牛，记录组成牛的牌
            self._niuCards = {}
            for k, v in pairs(cards) do
                table.insert(self._niuCards, v)
            end

            num = 0
            for idx = 1, 5 do   --  排除掉组成牛的牌，取最后两张算牛几
                if (not self:CardEqual(self._cards[idx], cards[1]) and not self:CardEqual(self._cards[idx], cards[2]) and not self:CardEqual(self._cards[idx], cards[3]))then
                    if (self._cards[idx]._value < 10) then
                        num = num + self._cards[idx]._value
                    end
                end
            end
            num = num % 10
            if (num == 0) then
                num = 10
            end
            break
        else
            num = 0
        end
    end

    return num
end

--  查找所有组合,combs:结果， cards:所有的牌， res：一组结果的索引, start:开始的索引, count:查找到第几个, num：组合中元素个数
function CardGroup:Combine_increase(combs, cards, res, start, count, num)
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


--取得跟A相等的数
function CardGroup:GetEqualsNumberOfA(cards, A)
    local number = 0
    for k, v in pairs(cards) do
        if v._value == A then
            number = number + 1
        end
    end

    return number
end



--  结比两张牌是否相等，包括花色
function CardGroup:CardEqual(card1, card2)
    if (card1._value == card2._value and card1._type == card2._type) then
        return true
    else
        return false
    end
end

--  对比两张牌谁比较大，注意，同时对比花色
function CompareCard(card1, card2)
    if(card1._value > card2._value)then
        return true
    elseif (card1._value == card2._value)then
        if (card1._type > card2._type)then
            return true
        else
            return false
        end
    else
        return false
    end

end
