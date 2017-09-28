
--表示牌的KEY
CardName = {"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}

Card = {
    --  1,2,3,4,5 对应：方块，梅花，红心，黑桃, 大小王
    _type,
    --  牛牛从1开始，13张从2开始,小王为20，大王为21
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

--  结比两张牌是否相等，包括花色
function CardEqual(card1, card2)
    if (card1._value == card2._value and card1._type == card2._type) then
        return true
    else
        return false
    end
end

--  对比两张牌谁比较大，注意，不对比花色
function CompareCard(card1, card2)
    if(card1._value > card2._value)then
        return 1
    elseif (card1._value == card2._value)then
        return 0
    else
        return -1
    end

end
