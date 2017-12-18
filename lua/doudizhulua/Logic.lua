package.path = os.getenv("PWD") .. '/?.lua;'
require "lua/doudizhulua/Card"
require "lua/doudizhulua/Tools"
require "lua/doudizhulua/json"--""../../lua/doudizhulua/json"

--package.path = os.getenv("PWD") .. '/?.lua;'
--require "Card"
--require "Tools"
--require "json" --""../../lua/doudizhulua/json"

Logic = { cards = {}, cardNum = 108, existKing = false, instance = nil }
Logic.__index = Logic

function G_Init(s)
    print(s)
    math.randomseed(s)
end

function G_Reset()
    Logic:GetInstance():ReSet()
end

function G_GetCards()
    return Logic:GetInstance():GetCards()
end

function G_GetResult(data, roomData)
    return Logic:GetInstance():GetResult(data, roomData)
end

function Logic:new()
    self = {}
    setmetatable(self, Logic)

    self:ReSet()
    self:InitCards()
    return self
end

function Logic:GetInstance()
    if Logic.instance == nil then
        Logic.instance = Logic:new()
    end

    return Logic.instance
end

function Logic:ReSet()
    self.cardNum = 108
end

--初始化牌
function Logic:InitCards()
    for n = 1, 2 do
        for t = 1, 4 do
            for v = 1, 13 do
                local card = Card:new(t, v)
                table.insert(self.cards, card)
            end
        end
        local card = Card:new(5, 20)
        table.insert(self.cards, card)
        local card = Card:new(5, 21)
        table.insert(self.cards, card)
    end
end


--取牌，返回13张牌
function Logic:GetCards()
    --MyLog(roomData)
    local allcardvec = {}
    for n = 1, 5 do
        local cardvec = {}
        cardNumber = 25
        if n == 5 then cardNumber = 8 end
        for i = 1, cardNumber do
            local rn = math.random(self.cardNum)
            print(rn)
            table.insert(cardvec, self.cards[rn]._type .. "_" .. self.cards[rn]._value)
            self.cards[rn], self.cards[self.cardNum] = self.cards[self.cardNum], self.cards[rn]
            self.cardNum = self.cardNum - 1
        end
        table.insert(allcardvec, cardvec)
        --dump(cardvec)
    end
    return allcardvec
end


--取牌，返回13张牌
function Logic:GetCards()

    local allcardvec = {}
    for n = 1, 5 do
        local cardvec = {}
        cardNumber = 25
        if n == 5 then cardNumber = 8 end

        for i = 1, cardNumber do

            local rn = 1
            if self.cardNum ~= 1 then
                rn = math.random(self.cardNum)
            end

            table.insert(cardvec, self.cards[rn]._type .. "_" .. self.cards[rn]._value)
            self.cards[rn], self.cards[self.cardNum] = self.cards[self.cardNum], self.cards[rn]
            self.cardNum = self.cardNum - 1
        end
        table.insert(allcardvec, cardvec)
    end

    --print("当前总牌数: " .. #self.cards .. " 当前剩余牌数: " .. self.cardNum)
    --dump(allcardvec)
    return allcardvec
end
