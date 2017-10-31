require "lua/niuniulua/Card"
require "lua/niuniulua/Tools"
require "../../lua/niuniulua/json"
Logic = {cards = {}, cardNum = 52, existKing = false}
Logic.__index = Logic

--从无牛到五小牛的赔率
--------------无牛 1  2   3   4   5   6   7   8   9   10  四炸  五花  五小
local global_win3 = {1, 1, 1,  1,  1,  1,  1,  1,  2,  2,  3,  3,     3,   3}
local global_win5 = {1, 1, 1,  1,  1,  1,  1,  2,  3,  4,  5,  5,     5,   5}
local global_win10= {1, 1, 2,  3,  4,  5,  6,  7,  8,  9, 10, 10,     10, 10}


function Logic:new()
    self = {}
    math.randomseed(os.time())
    setmetatable(self, Logic)

    self:ReSet()
    self:InitCards()
    return self
end

function Logic:ReSet()
    self.cardNum = 52
end

--初始化牌
function Logic:InitCards()
    for t = 1, 4 do
        for v = 1, 13 do
            local card = Card:new(t, v)
            table.insert(self.cards, card)
        end
    end
end


--[[

playerTbl ={
    result = {
        UserID = 100000,
        Cards = {},
        Info = {Role = 1, BankerScore = 10, BetScore = 10}
    }
}

BankerNoNiu   = 1 //无牛下庄
BankerTurns   = 2 //轮流上庄
BankerSeeCard = 3 //看牌上庄
BankerDefault = 4 //固定庄家

roomData = {
    Times = 1,
    BankerType = 1
}

看牌抢庄 结算=下注数*N*庄家倍数*自家抢庄倍数
其他模式 结算=下注数*N
]]--
--  计算结果
function Logic:CalculateRes(playerTbl, roomData)

    print(playerTbl)
    print(roomData)
    playerTbl = json.decode(playerTbl)
    roomData = json.decode(roomData)

    --dump(playerTbl, "playerTbl", 100)
    local banker = {}
    local bankerGroup = {}

    --  先找到庄家是谁
    for k, v in pairs(playerTbl.List)do
        if v.Info.Role == 1 then
            banker = v
            bankerGroup = CardGroup:new(String2Cards(banker.Cards.CardList))
            banker.Cards.CardType = tostring(bankerGroup._type)
            print(banker.Cards.CardType)
            break
        end
    end


    --  庄家输赢
    local bankerWin = 0
    --  所有人和庄家比牌
    for k, player in pairs(playerTbl.List)do
        if player.Info.Role ~= 1 then
            local playerGroup = CardGroup:new(String2Cards(player.Cards.CardList))
            local winN = self:CompareGroup(playerGroup, bankerGroup, roomData.Times) --  player的输赢倍数
            local win = 0
            if roomData.BankerType == 3 then
                --看牌抢庄
                if banker.Info.BankerScore == 0 then
                    banker.Info.BankerScore = 1
                end

                win = player.Info.BetScore * winN * banker.Info.BankerScore
            else
                win = player.Info.BetScore * winN
            end
            player.Cards.CardType = tostring(playerGroup._type)
            print("type: " .. player.Cards.CardType)
            player.Score = win
            banker.Score = banker.Score + win * -1
        end
    end

    local str = json.encode(playerTbl)

    --dump(playerTbl, "playerTbl", 100)
    return str
end


--取牌，返回13张牌
function Logic:GetCards()
    local cardvec = {}

    if #self.cards < 5 then
        print("没有那么多牌了")
        return cardvec
    end

    for i = 1, 5 do
        local rn = math.random(self.cardNum)
        table.insert(cardvec, i, self.cards[rn])
        self.cards[rn], self.cards[self.cardNum] = self.cards[self.cardNum], self.cards[rn]
        self.cardNum = self.cardNum - 1
    end

    return cardvec
end

function Logic:TempCards2Cards( tempcards)
    -- body
    local cards = {}
    for i, cardinfo in pairs(tempcards) do
        local card = Card:new(cardinfo._type, cardinfo._value)
        table.insert(cards, card)
    end

    return cards
end

--取得牌的牌组类型
function Logic:GetGroupType(cards)
    local group = CardGroup:new(cards)
    return group._type
end

function Logic:GetGroupTypeTest()

    local cardstring = {"2_13","3_6","2_2","1_1","4_3"}

    local cards = String2Cards(cardstring)
    local group = CardGroup:new(cards)
    return group._type
end

--取得组成牛的牌
function Logic:GetCardsGroupNiu(cards)
    local group = CardGroup:new(cards)
    return group._niuCards
end



--  对比两个组合的大小，group1大于group2则返回1的倍数,否则返回2的倍数 * -1
function Logic:CompareGroup(group1, group2, module)
    local res = 0
    local win = 1
    if (group1._type > group2._type)then
        res = group1._type
    elseif (group1._type == group2._type)then
        if (CompareCard(group1._cards[1], group2._cards[1]))then
            res = group1._type
        else
            res = group2._type
            win = -1
        end
    else
        res = group2._type
        win = -1
    end

    print("module:" .. module)
    if module == 3 then
        res = global_win3[res+1]
    elseif module == 5 then
        res = global_win5[res+1]
    elseif module == 10 then
        res = global_win10[res+1]
    end

    res = res * win

    return res
end
