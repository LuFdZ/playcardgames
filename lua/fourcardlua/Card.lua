--表示牌的KEY
CardName = { "A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "JB" }

--
--牌型大小
--至尊 > 天牌对 > 地牌对 > 人牌对 > 和牌对 > 长牌对 > 短牌对 > 杂牌对 > 天王 > 天杠 > 地杠 > 点数
local CardGroupKV = {
    ["1_3-5_21"] = 110, --至尊：大王+红桃3
    ["3_3-5_21"] = 110,
    ["1_12-3_12"] = 100, --天牌对：2张红Q

    ["1_2-3_2"] = 90, --地牌对：2张红2

    ["1_8-3_8"] = 80, --人牌对：2张红8

    ["1_4-3_4"] = 70, --和牌对：2张红4

    ["2_10-4_10"] = 60, --长牌对：2张黑10、黑6、黑4
    ["2_6-4_6"] = 60,
    ["2_4-4_4"] = 60,
    ["1_11-3_11"] = 50, --短牌对：2张红J、红10、红7、红6
    ["1_10-3_10"] = 50,
    ["1_7-3_7"] = 50,
    ["1_6-3_6"] = 50,
    ["1_9-3_9"] = 40, --杂牌对：2张红9、黑8、黑7、红5
    ["1_5-3_5"] = 40,
    ["2_7-4_7"] = 40,
    ["2_8-4_8"] = 40,
    ["1_9-1_12"] = 30, --天王：红Q+红9
    ["3_9-3_12"] = 30,
    ["1_9-3_12"] = 30,
    ["1_12-3_9"] = 30,
    ["1_12-8"] = 20, --天杠：红Q+任意8
    ["3_12-8"] = 20,
    ["1_2-8"] = 10, --地杠：红2+任意8
    ["3_2-8"] = 10,
}
--同样点数，根据2张中较大的一张牌的牌型比大小
--天牌>地牌>人牌>和牌>长牌>短牌>杂牌>大王/红桃3
local CardKV = {
    ["1_12"] = 8; --天牌 红Q
    ["3_12"] = 8;

    ["1_2"] = 7; --地牌 红2
    ["3_2"] = 7;

    ["1_8"] = 6; --人牌 红8
    ["3_8"] = 6;

    ["1_4"] = 5; --和牌 红4
    ["3_4"] = 5;

    ["2_10"] = 4; --长牌 黑10、黑6、黑4
    ["2_6"] = 4;
    ["2_4"] = 4;
    ["4_10"] = 4;
    ["4_6"] = 4;
    ["4_4"] = 4;

    ["1_11"] = 3; --短牌 红J、红10、红7、红6
    ["1_10"] = 3;
    ["1_7"] = 3;
    ["1_6"] = 3;
    ["3_11"] = 3;
    ["3_10"] = 3;
    ["3_7"] = 3;
    ["3_6"] = 3;

    ["1_9"] = 2; --杂牌 红9、黑8、黑7、红5
    ["1_5"] = 2;
    ["2_8"] = 2;
    ["2_7"] = 2;
    ["3_9"] = 2;
    ["3_5"] = 2;
    ["4_8"] = 2;
    ["4_7"] = 2;

    ["3_3"] = 1; --红桃3

    ["5_21"] = 1; --大王
}
local ScoreKV = {
    [110] = 13, --至尊13分
    [100] = 12, --对子12分
    [90] = 12,
    [80] = 12,
    [70] = 12,
    [60] = 12,
    [50] = 12,
    [40] = 11, --天王11分
    [30] = 11,
    [20] = 10, --天杠10分
    [10] = 10, --地杠10分
}
--}
Card = {
    --  1,2,3,4,5 对应：方块，梅花，红心，黑桃，大王
    _type,
    --【牌型规则】
    --牌型(共32张)
    --1、天牌对：2张红Q
    --2、地牌对：2张红2
    --3、人牌对：2张红8
    --4、和牌对：2张红4
    --5、长牌对：2张黑10、黑6、黑4
    --6、短牌对：2张红J、红10、红7、红6
    --7、杂牌对：2张红9、黑8、黑7、红5
    --8、至尊：大王+红桃3
    --9、天王：红Q+红9
    --10、天杠：红Q+任意8
    --11、地杠：红2+任意8
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

function GetUserResult(userCards)
    local cardA = userCards.CardList[1]
    local cardB = userCards.CardList[2]
    local typeKey = cardA .. "-" .. cardB
    local cardType = CardGroupKV[typeKey]
    local score = 0
    if cardType == nil then
        cardA = CardKV[cardA]
        cardB = CardKV[cardB]
        if cardA > cardB then
            cardType = cardA
        else
            cardType = cardB
        end
        local arrA = split(cardA, "_")
        local cardAScore = tonumber(arrA[2])
        local arrB = split(cardB, "_")
        local cardBScore = tonumber(arrB[2])
        score = (cardAScore + cardBScore) % 10
    else
        score = ScoreKV[cardType]
    end
    userCards.Score = score
    userCards.CardType = cardType
    return userCards.Score, userCards.CardType
end

function ChekcUserCardList(head, tail)
    local headScore,headType = GetUserResult(head)
    local tailScore,tailType = GetUserResult(tail)
    if headType > tailType then
        return false
    end
    if headScore > tailScore then
        return false
    end
end