--  定义牌组的类型
GroupTypeName = {"Single", "Couple", "TwoCouple", "Three", "Straight", "Flush", "ThreeCouple", "Four", "FlushStraight"}



CardGroup = {
    --  所有的牌
    _cards = {},
    --  对应GroupTypeName
    _type = 1,
    --  关键牌
    _keyCard = {}
}

CardGroup.__index = CardGroup

function CardGroup:new(cards, t, key)
    self = {}
    setmetatable(self, CardGroup)

    self._cards = cards
    self._type = t
    self._keyValue = key
    return self
end
