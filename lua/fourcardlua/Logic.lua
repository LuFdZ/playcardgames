package.path = os.getenv("PWD") .. '/?.lua;'
require "lua/fourcardlua/Card"
require "lua/fourcardlua/Tools"
require "/lua/fourcardlua/json"
Logic = { cards = {}, cardNum = 52, existKing = false, instance = nil }
Logic.__index = Logic

--从无牛到五小牛的赔率
-------------- 无牛 1  2   3   4   5   6   7   8   9   10  四炸  五花  五小
local global_win3 = { 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 3, 3, 3, 3 }
local global_win5 = { 1, 1, 1, 1, 1, 1, 1, 2, 3, 4, 5, 5, 5, 5 }
local global_win10 = { 1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 10, 10, 10 }


local global_card_type = { 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 5 }
local global_card_value = { 2, 4, 5, 6, 7, 8, 9, 10, 11, 12, 4, 6, 7, 8, 10, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 4, 6, 7, 8, 10, 21 }

function G_Init(s)
    math.randomseed(s)
end

function G_Reset()
    Logic:GetInstance():ReSet()
end

function G_GetCards()
    return Logic:GetInstance():GetCards()
end

function G_CalculateRes(playerTbl, roomData)
    return Logic:GetInstance():CalculateRes(playerTbl, roomData)
end

function Logic:GetInstance()
    if Logic.instance == nil then
        Logic.instance = Logic:new()
    end

    return Logic.instance
end

function Logic:new()
    self = {}
    setmetatable(self, Logic)

    self:ReSet()
    self:InitCards()
    return self
end

function Logic:ReSet()
    self.cardNum = 32
end

--初始化牌
function Logic:InitCards()
    for i = 1, 32 do
        local card = Card:new(global_card_type[i], global_card_value[i])
        table.insert(self.cards, card)
    end
end

--  计算结果
function Logic:CalculateRes(playerTbl, roomData)

    --    print(playerTbl)
    --    print(roomData)
    --    playerTbl = json.decode(playerTbl)
    --    roomData = json.decode(roomData)

    --dump(playerTbl, "playerTbl", 100)
    local banker = {}

    --  计算所有玩家牌组分值并找到庄家
    for k, user in pairs(playerTbl.List) do
        GetUserResult(String2Cards(user.HeadCards), roomData.ScoreType)
        GetUserResult(String2Cards(user.TailCards), roomData.ScoreType)
        if user.Role == 1 then
            banker = user
        end
    end
    print(playerTbl.List)
    --  庄家输赢
    local bankerWin = 0
    --  所有人和庄家比牌
    for k, player in pairs(playerTbl.List) do
        if player.Role ~= 1 then
            CalculateCards(player.HeadCards, banker.HeadCards, roomData.ScoreType)
            CalculateCards(player.TailCards, banker.TailCards, roomData.ScoreType)
            player.TotalScore = player.HeadCards + player.TailCards * (player.Bet + 1)
        end
    end

    local str = json.encode(playerTbl)

    --dump(playerTbl, "playerTbl", 100)
    return str
end

function CalculateCards(playerCards, bankerCards, scoreType)
    if playerCards.CardType > bankerCards.CardType and playerCards.Score > bankerCards.Score then
        if scoreType == 1 then
            playerCards.Win = 1
            bankerCards.Win = bankerCards.Win - 1
        elseif scoreType == 2 then
            playerCards.Win = playerCards.Score
            bankerCards.Win = bankerCards.Win - playerCards.Score
        end
    else
        if scoreType == 1 then
            playerCards.Win = -1
            bankerCards.Win = bankerCards.Win + 1
        elseif scoreType == 2 then
            playerCards.Win = -playerCards.Score
            bankerCards.Win = bankerCards.Win + playerCards.Score
        end
    end
end


--取牌，返回4张牌
function Logic:GetCards()
    local cardvec = {}

    if #self.cards < 4 then
        print("没有那么多牌了")
        return cardvec
    end

    for i = 1, 4 do
        local rn = math.random(self.cardNum)
        table.insert(cardvec, i, self.cards[rn])
        self.cards[rn], self.cards[self.cardNum] = self.cards[self.cardNum], self.cards[rn]
        self.cardNum = self.cardNum - 1
    end

    return cardvec
end

