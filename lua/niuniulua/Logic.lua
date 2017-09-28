require "lua/niuniulua/CardGroup"
require "lua/niuniulua/Card"
require "lua/niuniulua/Tools"
require "../../lua/niuniulua/json"



Logic = {cards = {}, cardNum = 52, existKing = false}
Logic.__index = Logic
GroupType = {}
--表示牌的KEY
CardValueData = {"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
--牌的值
CardValue = {}

function Logic:new()
    self = {}
    math.randomseed(os.time())
    setmetatable(self, Logic)
    --创建牌组型枚举
    GroupType = CreateEnumTable(GroupTypeName, -1)
    --dump(GroupType)

    --创建牌型枚举
    CardValue = CreateEnumTable(CardValueData, 1)
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
        for v = 2, 14 do
            local card = Card:new(t, v)
            table.insert(self.cards, card)
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
        table.insert(cardvec, i, self.cards[rn])
        self.cards[rn], self.cards[self.cardNum] = self.cards[self.cardNum], self.cards[rn]
        self.cardNum = self.cardNum - 1

    end

    return cardvec
end

function Logic:GetNiuniuCards()
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

--比牌，group1大于group2时返回1 * 牌型，否则返回-1 * 牌型，相等返回0
function Logic:CompareGroup(group1, group2)
    if #group1 ~= #group2 then
        print("比牌错误，牌的数量不同！")
        return 0
    end

    if group1._type > group2._type then
        return 1 * group1._type
    elseif group1._type < group2._type then
        return -1 * group2._type
    end

    --如果类型相等，则比关键牌
    local res = 0
    for k, keycard in pairs(group1._keyCard) do
        res = CompareCard(keycard, group2._keyCard[k])
        if res > 0 then
            res = res * group1._type
            break
        elseif res < 0 then
            res = res * group2._type * -1
            break
        end
    end

    return res
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
function Logic:GetResult( data, roomData )
    print(data)
    print(roomData)

    local datas = json.decode(data)
    roomData = json.decode(roomData)
    -- 计算所有人的牌组
    --ResGroup
    local ThirteenResultList = {}
    for key, value in pairs(datas) do
        local Result = {}
        local ThirteenResult = {}
        ThirteenResult.UserID = value.UserID
        if value.Role == 1 then
            ThirteenResult.Banker = true
        else
            ThirteenResult.Banker = false
        end

        Result.Head = self:GetGroup(String2Cards(value.Head))
        Result.Middle = self:GetGroup(String2Cards(value.Middle))
        Result.Tail = self:GetGroup(String2Cards(value.Tail))
        ThirteenResult.Result = Result
        ThirteenResult.Result.Shoot = {}
        --记录赢每一个人的数分
        ThirteenResult.winScore = {}
        ThirteenResult.Settle = {Head = 0, Middle = 0, Tail = 0, AddScore = 0, TotalScore = 0}
        table.insert( ThirteenResultList,ThirteenResult )
    end

    --取一个和其他的做对比
    for key, value in pairs(ThirteenResultList) do
        for key1, value1 in pairs(ThirteenResultList) do
            --value和temp对比
            local Settle = {}
            local addScore = 0
            if value.UserID ~= value1.UserID then
                --对比大小
                local resHead = self:CompareGroup(value.Result.Head, value1.Result.Head)
                --大于0的标记
                local temp = 1
                if resHead < 0 then
                --小于0
                    temp = -1
                    addScore = roomData.BankerAddScore * -1 + addScore
                else
                    addScore = roomData.BankerAddScore + addScore
                end
                --计分
                if resHead == GroupTypeName.Three then
                    Settle.Head = 3
                elseif resHead ~= 0 then
                    Settle.Head = 1
                else
                    Settle.Head = 0
                end
                --乘于大于0的标记，如果标记小于0，value输，分数为负
                Settle.Head = Settle.Head * temp

                --对比大小
                local resMiddle = self:CompareGroup(value.Result.Middle, value1.Result.Middle)
                --大于0的标记
                temp = 1
                if resMiddle < 0 then
                    temp = -1
                    resMiddle = resMiddle * temp
                end

                if resMiddle == GroupType.Four then
                    Settle.Middle = 7

                elseif resMiddle == GroupType.FlushStraight then
                    Settle.Middle = 8

                elseif resMiddle ~= 0 then
                    Settle.Middle = 2
                else
                    Settle.Middle = 0
                end

                Settle.Middle = Settle.Middle * temp

                local resTail = self:CompareGroup(value.Result.Tail, value1.Result.Tail)
                temp = 1
                if resTail < 0 then
                    temp = -1
                    resTail = resTail * temp
                end
                if resTail == GroupType.Four then
                    Settle.Tail = 7

                elseif resTail == GroupType.FlushStraight then
                    Settle.Tail = 8

                elseif resTail ~= 0 then
                    Settle.Tail = 3
                else
                    Settle.Tail = 0
                end

                Settle.Tail = Settle.Tail * temp

                if Settle.Head == 0 then--头道相等
                    if Settle.Tail == 0 then --根据尾道大小
                        if Settle.Middle == 0 then--尾道相等，根据中道大小
                            --三道一样，平分
                            Settle.Head = 0
                            Settle.Middle = 0
                            Settle.Tail = 0
                        elseif Settle.Middle > 0 then
                        --中道大,计算得分，中道大，尾送也大，尾道大，头道大
                            if value.Tail._type == GroupType.FlushStraight then
                            --同花顺，8分
                                Settle.Tail = 8

                            else
                                Settle.Tail = 3
                            end
                            Settle.Head = 1
                        else
                        --中道大，计算得分，中道小，尾道小，头道小
                            if value1.Tail._type == GroupType.FlushStraight then
                            --同花顺，8分
                                Settle.Tail = -8

                            else
                                Settle.Tail = -3
                            end
                            Settle.Head = -1
                        end
                    elseif Settle.Tail > 0 then --尾道大，头道也大
                        Settle.Head = 1
                    else--尾道小，头道也小
                        Settle.Head = -1
                    end
                end
                if Settle.Middle == 0 then--中道等
                    if Settle.Tail == 0 then--按尾道比大小，尾道等，接头道比大小
                        if Settle.Head > 0 then--头道大，尾道也大，尾道大，中道也大
                            --尾
                            if value.Tail._type == GroupType.FlushStraight then
                                Settle.Tail = 8

                            else
                                Settle.Tail = 3
                            end
                            --中
                            if value.Middle._type == GroupType.FlushStraight then
                                Settle.Middle = 8

                            else
                                Settle.Middle = 3
                            end
                        else --如果头道小，尾道和中道都小
                            --尾道
                            if value1.Result.Tail._type == GroupType.FlushStraight then
                                    Settle.Tail = -8

                                else
                                    Settle.Tail = -3
                            end
                            --中道
                            if value1.Result.Middle._type == GroupType.FlushStraight then
                                Settle.Middle = -8

                            else
                                Settle.Middle = -3
                            end
                        end
                    elseif Settle.Tail > 0 then--尾道大，中道大
                        --中
                        if value.Result.Middle._type == GroupType.FlushStraight then
                            Settle.Middle = 8

                        else
                            Settle.Middle = 3
                        end
                    else --尾道小，中道小
                         --中道
                        if value1.Result.Middle._type == GroupType.FlushStraight then
                            Settle.Middle = -8

                        else
                            Settle.Middle = -3
                        end
                    end
                end


                if Settle.Tail == 0 then
                --尾道相等，按中道比
                    if Settle.Middle == 0 then
                        --中道相等，按头道
                        if Settle.Head == 0 then
                            --三个都相待，平局
                            Settle.Head = 0
                            Settle.Middle = 0
                            Settle.Tail = 0
                        elseif Settle.Head > 0 then
                            --尾
                            if value.Result.Tail._type == GroupType.FlushStraight then
                                Settle.Tail = 8

                            else
                                Settle.Tail = 3
                            end
                            --中
                            if value.Result.Middle._type == GroupType.FlushStraight then
                                Settle.Middle = 8

                            else
                                Settle.Middle = 3
                            end
                        else
                            --尾
                            if value1.Result.Tail._type == GroupType.FlushStraight then
                                Settle.Tail = -8

                            else
                                Settle.Tail = -3
                            end
                            --中
                            if value1.Result.Middle._type == GroupType.FlushStraight then
                                Settle.Middle = -8

                            else
                                Settle.Middle = -3
                            end
                        end
                    elseif Settle.Middle > 0 then
                        --中道大，尾道也大
                        --尾
                        if value.Result.Tail._type == GroupType.FlushStraight then
                            Settle.Tail = 8

                        else
                            Settle.Tail = 3
                        end
                    else
                        --中道小，尾道也小
                        --尾
                        if value1.Result.Tail._type == GroupType.FlushStraight then
                            Settle.Tail = -8

                        else
                            Settle.Tail = -3
                        end
                    end
                end


                --计算结果完成
                if Settle.Head > 0 and Settle.Middle > 0 and Settle.Tail > 0 then
                    --打枪，value 打 value1

                    table.insert( value.Result.shoot, value1.UserID )
                end

                if Settle.Head < 0 and Settle.Middle < 0 and Settle.Tail < 0 then

                    --打枪
                    table.insert( value1.Result.shoot, value.UserID )
                end

                value.Settle.Head = value.Settle.Head + Settle.Head
                value.Settle.Middle = value.Settle.Middle + Settle.Middle
                value.Settle.Tail = value.Settle.Tail + Settle.Tail
                value.Settle.AddScore = addScore
                local winScore = {}
                winScore.winID = value1.UserID
                winScore.Score = (value.Settle.Head + value.Settle.Middle + value.Settle.Tail)
                table.insert( value.winScore, winScore )
            end
        end
    end

    for k, value in pairs(ThirteenResultList) do
        for kk, value1 in pairs(value.winScore) do
            local totalScore = 0
            local times = 1
            for i, shootid in pairs(value.Result.Shoot)do
                if shootid == value1.winID then
                    --打枪
                    times = 2
                    break
                end
            end

            if roomData.Times == 3 then
            --3翻模式-没有全垒打
                times = math.pow(2,#value.Result.Shoot)
            elseif #value.Result.Shoot == #ThirteenResultList then
                times = 4
            end

            value.Settle.TotalScore = value.Settle.TotalScore + times * value1.Score

            --如果一方是庄家，则+庄家分
            if roomData.BankerAddScore ~= 0 then
                --取到跟value对比的玩家结果信息
                local winUser = self:GetUserResult(value1.winID, ThirteenResultList)
                if value.Banker or winUser.Banker then
                    --当前比牌有一个是庄家
                    local bankerAddScore = 0
                    if value1.Score > 0 then
                        bankerAddScore = roomData.BankerAddScore
                    elseif value1.Score < 0 then
                        bankerAddScore = roomData.BankerAddScore * -1
                    end

                    value.Settle.TotalScore = value.Settle.TotalScore + bankerAddScore
                end
            end
        end
    end

    local results = {}
    for k, value in pairs(ThirteenResultList) do
        local res = {Head = {}, Middle = {}, Tail = {}}
        res.Head.CardList = Cards2String(value.Result.Head._cards)
        res.Head.GroupType = GroupTypeName[value.Result.Head._type+1]

        res.Middle.CardList = Cards2String(value.Result.Middle._cards)
        res.Middle.GroupType = GroupTypeName[value.Result.Middle._type+1]

        res.Tail.CardList = Cards2String(value.Result.Tail._cards)
        res.Tail.GroupType = GroupTypeName[value.Result.Tail._type+1]

        res.Shoot = value.Result.Shoot

        local settle = {}
        settle.Head = tostring(value.Settle.Head)
        settle.Middle = tostring(value.Settle.Middle)
        settle.Tail = tostring(value.Settle.Tail)
        settle.AddScore = tostring(value.Settle.AddScore)
        settle.TotalScore = tostring(value.Settle.TotalScore)

        local result = {Result = res, Settle = settle, UserID = value.UserID}
        table.insert(results, result)
    end
    --dump(results, "results", 100)
    local str = json.encode(results)
    print(str)
    --dump(ThirteenResultList, "ThirteenResultList", 100)
    return str
end

--根据id取得玩家的结果
function Logic:GetUserResult(id, list)
    for k, v in pairs(list) do
        if v.UserID == id then
            return v
        end
    end
    print("找不到玩家结果")
    return nil
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

    if #cards < 4 and #cards > 0 then
        --对子，三个或乌龙
        local res, keycards = self:IsThree(cards)
        if res then
        --三条
            local group = CardGroup:new(cards, GroupType.Three, keycards)
            return group
        else
            local combs = {}
            local result = {}
            self:Combine_increase(combs, cards, result, 1, 2, 2)
            local keycards = {}
            local res = false
            for k, v in pairs(combs) do
                res, keycards = self:IsCouple(v)
                if res then
                    --是对子牌
                    --取出跟keycard不同的牌，此处为对子，只可能有一张不同的
                    local tempcards = self:GetNotEqualsInTableCards(cards, keycards[1])
                    if #tempcards ~= 0 then
                        --tempcards数量不为0，找到了
                        table.insert(keycards, tempcards[1])
                    end
                    break
                end
            end

            if res then
                --有对子，创建牌组
                local group = CardGroup:new(cards, GroupType.Couple, keycards)
                return group
            end
            --如果不是三张，也不是对子，那就是乌龙
            --留到最后统一回复
        end
    elseif #cards == 5 then
        local res = false
        local keycards = {}
        --先判断是否同花
        res, keycards = self:AllTheSameType(cards)
        --判断是否为顺子
        local res1, keycards1 = self:IsStraight(cards)
        if res and res1 then
            --同花顺
            local group = CardGroup:new(cards, GroupType.FlushStraight, keycards)
            return group
        end

        if res then
            --同花
            local group = CardGroup:new(cards, GroupType.Flush, keycards)
            return group
        end

        if res1 then
        --顺子
            local group = CardGroup:new(cards, GroupType.Straight, keycards)
            return group
        end
        --判断是否铁支
        local combs = {}
        local result = {}
        self:Combine_increase(combs, cards, result, 1, 4, 4)
        res = false
        keycards = {}
        for k, v in pairs(combs) do
            res, keycards = self:IsFour(v)
            if res then
                break
            end
        end
        if res then
            for k, v in cards do
                if v._value ~= keycards[1]._value then
                    table.insert(keycards, v)
                    break
                end
            end

            local group = CardGroup:new(cards, GroupType.Four, keycards)
            return group
        end

        --判断是否葫芦
        res = false
        keycards = {}
        res, keycards = self:IsThreeWithCouple(cards)
        if res then
            local group = CardGroup:new(cards, GroupType.ThreeCouple, keycards)
            return group
        end
        --判断是否两对
        res = false
        keycards = {}
        --上面判断是否铁支已经做过4张牌的组合操作，此处直接拿来用即可
        for k, v in pairs(combs) do
            res, keycards = self:IsTwoCouple(v)
            if res then
                --是两对，找到最后一张牌加到关键组
                for key, vcard in pairs(cards) do
                    for kk, keycard in pairs(keycards) do
                        if vcard._value ~= keycard._value then
                            table.insert( keycards, vcard )
                            --找到了那张牌
                            local group = CardGroup:new(cards, GroupType.TwoCouple, keycards)
                            return group
                        end
                    end
                end
                --如果运行到此，说明代码sb了，请检查代码
                print("明代码sb了，请检查代码 a")
            end
        end

        --判断是否对子
        res = false
        keycards = {}
        local combs = {}
        self:Combine_increase(combs, cards, res, 1, 2, 2)
        for k, v in pairs(combs) do
            res, keycards = self:IsCouple(v)
            if res then
                break
            end
        end

        if res then
            --找到了对子，再找出其他的牌
            for k, v in pairs(cards) do
                if v._value ~= keycards[1]._value then
                    table.insert( keycards, v )
                end
            end
            local group = CardGroup:new(cards, GroupType.Couple, keycards)
            return group
        end

        print("sb了，请检查代码")
    else
        --不符合配牌牌型
        print("出错了，请检查传过来的牌是否正常:")
        dump(cards)
    end

    --运行到此处，只可能是乌龙
    local group = CardGroup:new(cards, GroupType.Single, {cards})
    return group
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

--取出在table里跟card不同的牌
function Logic:GetNotEqualsInTableCards(cards, card)
    local res = {}
    for k, v in pairs(cards) do
        if v._value ~= card._value and v._type ~= card._value then
            table.insert( res, v )
        end
    end

    return res
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
    local temp = self:GetGroupInCombs(combs, GroupType.Couple)
    TableInsert2Table(self.groups, temp)
    --检查两对
    combs = {}
    self:Combine_increase(combs, cards, res, 1, 4, 4)
    temp = self:GetGroupInCombs(combs, GroupType.TwoCouple)
    TableInsert2Table(self.groups, temp)
    --检查线三
    combs = {}
    self:Combine_increase(combs, cards, res, 1, 3, 3)
    temp = self:GetGroupInCombs(combs, GroupType.Three)
    TableInsert2Table(self.groups, temp)
    --检查四条
    combs = {}
    self:Combine_increase(combs, cards, res, 1, 4, 4)
    temp = self:GetGroupInCombs(combs, GroupType.Four)
    TableInsert2Table(self.groups, temp)
    --检查五张的牌组
    combs = {}
    temp = self:Combine_increase(combs, cards, res, 1, 5, 5)
    TableInsert2Table(self.groups, temp)
    --顺子
    temp = self:GetGroupInCombs(combs, GroupType.Straight)
    TableInsert2Table(self.groups, temp)
    --同花
    temp = self:GetGroupInCombs(combs, GroupType.Flush)
    TableInsert2Table(self.groups, temp)
    --同花顺子
    temp = self:GetGroupInCombs(combs, GroupType.FlushStright)
    TableInsert2Table(self.groups, temp)
    --葫芦
    temp = self:GetGroupInCombs(combs, GroupType.ThreeCouple)
    TableInsert2Table(self.groups, temp)

    --for k, v in pairs(self.groups) do
    --    print("========================================group: " , v._type)
    --end
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
        return false, nil
    end

    local keycards = {}
    local sameCard = self:EqualsNumber(cards)
    if sameCard.Num == 3 then
        table.insert(keycards, sameCard.card)
        return true, keycards
    end
    return false, nil
end

--判断是否一对
function Logic:IsCouple(cards)

    --对子
    if #cards ~= 2 then
        return false, 0
    end

    local sameCard = self:EqualsNumber(cards)
    if sameCard.Num == 2 then
        return true, {sameCard.card}
    end

    return false, 0
end

--判断是否四条
function Logic:IsFour(cards)
    if #cards ~= 4 then
        return false, 0
    end

    local sameCard = self:EqualsNumber(cards)
    --关键牌
    local keycards = {}

    if sameCard.Num == 4 then

        --添加一张关键牌
        table.insert(keycards, sameCard.card)
        return true, keycards
    end

    return false, 0
end

--判断有多少个牌相等
function Logic:EqualsNumber(cards)
    local number = 0

    local tempList = {}
    for k, v in pairs(cards) do
        --取得点数和v._value相等的牌
        local tempNumber = self:GetEqualsNumberOfA(cards, v._value)
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
        print("sb了，数据有错:")
        dump(cards)
    end

    return tempList[1]
end

--取得跟A相等的数
function Logic:GetEqualsNumberOfA(cards, A)
    local number = 0
    for k, v in pairs(cards) do
        if v._value == A then
            number = number + 1
        end
    end

    return number
end
--判断是否所有的类型都相等(同花)
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


    return true, cards
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


    return true, cards
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
    local three = nil
    --用于存储关键牌
    local keycards = {}

    for k, vcards in pairs(combs) do
        local sameCard = self:EqualsNumber(vcards)
        if sameCard.Num == 3 then
            --三张相等的牌
            three = sameCard.card
            --添加一张关键牌
            table.insert(keycards, three)
            break
        end
    end


    --判断有三张
    if three == nil then
        return false, 0
    end

    local tempcards = {}

    for k, v in pairs(cards) do
        local same = false
        --检查牌是否在三张牌里面
        for k1, v1 in pairs(three) do
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
    local sameCard = self:EqualsNumber(tempcards)
    if sameCard.Num == 2 then
        --两张牌相等，是对子，符合三带对的规则
        --把对子的牌加入为关键牌
        table.insert(keycards, temp)
        return true, keycards
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
    for k, vcards in pairs(combs) do
        local sameCard = self:EqualsNumber(vcards)
        if sameCard.Num == 2 then
            table.insert(tempCombs, vcards)
        end
    end

    if #tempCombs == 2 then
        if CompareCard(tempCombs[1][1], tempCombs[2][1]) == 1 then
            return true, {tempCombs[1][1], tempCombs[2][1]}
        else
            return true, {tempCombs[2][1], tempCombs[1][1]}
        end
    end

    return false, 0
end
