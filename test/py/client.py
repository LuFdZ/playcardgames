# -*-coding:utf-8 -*-
import httplib2
import json
import thread
import websocket
import time

Host = "http://127.0.0.1:8080"


# Host = "http://game.372124.com:8888"

class Client(object):
    def __init__(self, name, opid, unid, host=Host):
        self.token = ""
        self.user_id = 0
        self.user = None
        self.account = self.GenAccount(name, opid, unid)
        self.host = host

    def GenAccount(self, name, opid, unid):
        return {
            "Username": name,
            "Password": name,
            "Nickname": name,
            "Email": name + "@x.com",
            "Icon": "12345567",
            "Sex": 1,
            "OpenID": opid,
            "UnionID": unid,
        }

    def Request(self, path, data):
        req = httplib2.Http()
        path = self.host + path
        data = json.dumps(data)
        _, rsp = req.request(path, "POST", data, headers={"Token": self.token})
        print "req: {0} -d '{1}' -H Token:{2}".format(path, data, self.token)
        print "rsp:", rsp
        return json.loads(rsp)

    def Register(self):
        """
        注册
        :return: 用户信息
        """
        u = self.Request("/user/UserSrv/Register", self.account)
        self.user_id = u["UserID"]
        return u

    def Login(self):
        """
        登录
        :return: Token
        """
        u = self.Request("/user/UserSrv/Login", self.account)
        self.token = u["Token"]
        return u

    def Stream(self):
        def on_open(ws):
            """
            websocket 连接上后，首先发送用户 Token 认证
            """
            print "websocket connected"
            # send Token to login
            print(self.token)
            ws.send(self.token)

        def on_message(ws, msg):
            print "websocket msg", msg

        def on_error(ws, error):
            print "websocket error:", error

        def on_close():
            print "websocket closed"

        websocket.enableTrace(True)
        self.ws = websocket.WebSocketApp(
            "ws://localhost:8999/stream",
            # "ws://game.372124.com:8888/stream",
            on_message=on_message,
            on_error=on_error,
            on_close=on_close,
            on_open=on_open,
        )

        thread.start_new_thread(self.ws.run_forever, ())

    def StreamSend(self, method, args):
        data = {
            "Method": method,
            "Args": args
        }

        b = json.dumps(data)
        self.ws.send(b)

    def AddZodiacConfig(self):
        """
        活动配置添加
        """
        ul = self.Request("/activity/ActivitySrv/AddConfig", {
            "ConfigID": 10001,
            "Description": "十二星座json活动配置",
            "Parameter": json.dumps({
                "StartDurationInday": "0s",
                "EndEDurationInDay": "1439m",
                "LifeDuration": "1m",
                "InitialGoldPool": 60000,
                "Rate": 0.1
            })
        })

    def GetConfigByID(self, configid):
        ul = self.Request("/activity/ActivitySrv/GetConfigByID", {
            "ID": configid
        })

    def Heartbeat(self):
        ul = self.Request("/room/roomSrv/Heartbeat", {

        })
        return ul

    def SubscribeRoom(self, Password):
        self.StreamSend("SubscribeRoomMessage", Password)

    def SubscribeThirteen(self):
        self.StreamSend("SubscribeThirteenMessage", {})

    def SubscribeBill(self):
        self.StreamSend("SubscribeBillMessage", {})

    def SubscribeNiu(self):
        self.StreamSend("SubscribeNiuniuMessage", {})

    def SubscribeClub(self):
        self.StreamSend("SubscribeClubMessage", {})

    def UnSubscribeClub(self):
        self.StreamSend("UnSubscribeClubMessage", {})

    def ClientHeartbeat(self):
        self.StreamSend("ClientHeartbeat", {})

    def WebScoket(self):
        self.StreamSend("UserWebScoketList", {})

    """
    房间相关操作
    """

    def CreateTRoom(self, roomType, maxNumber, roundNumber):
        ul = self.Request("/room/roomSrv/CreateRoom", {
            "RoundNumber": roundNumber,
            "MaxNumber": maxNumber,
            "GameType": 1001,
            "RoomType": roomType,
            "GameParam": "{\"BankerAddScore\":2,\"Time\":30,\"Joke\":0,\"Times\":2}"
        })
        return ul

    def EnterRoom(self, Password):
        ul = self.Request("/room/roomSrv/EnterRoom", {
            "Password": Password
        })
        return ul

    def LeaveRoom(self):
        ul = self.Request("/room/roomSrv/LeaveRoom", {
        })
        return ul

    def SetReady(self, Password):
        ul = self.Request("/room/roomSrv/SetReady", {
            "Password": Password
        })
        return ul

    def SubmitCardT(self):
        ul = self.Request("/thirteen/thirteenSrv/SubmitCard", {
            "Head": ['2_8', '2_5', '2_3'],
            "Middle": ['3_10', '3_11', '3_12', '3_13', '3_14'],
            "Tail": ['1_10', '1_11', '1_12', '1_13', '1_14']
        })
        return ul

    def SubmitCardT2(self):
        ul = self.Request("/thirteen/thirteenSrv/SubmitCard", {
            "Head": ['4_14', '4_13', '2_12'],
            "Middle": ['4_9', '3_9', '3_7', '2_7', '4_3'],
            "Tail": ['1_6', '2_6', '1_6', '1_4', '1_4']
        })
        return ul

    def GiveUpGame(self, pwd):
        ul = self.Request("/room/roomSrv/GiveUpGame", {
            "Password": pwd,
        })
        return ul

    def GiveUpVote(self, pwd, status):
        ul = self.Request("/room/roomSrv/GiveUpVote", {
            "Password": pwd,
            "AgreeOrNot": status,
        })
        return ul

    def CreateFeedback(self):
        ul = self.Request("/room/roomSrv/CreateFeedback", {
            "UserID": 100000,
            "Channel": "test",
            "Version": "1.0.1",
            "Content": "测试",
            "MobileModel": "123123123",
            "MobileNetWork": "123123123",
            "MobileOs": "1501655847",
            "LoginIP": "1601655847",
        })
        return ul

    def Renewal(self, pwd):
        ul = self.Request("/room/roomSrv/Renewal", {
            "Password": pwd,
        })
        return ul

    def PageFeedbackList(self, page, pagesize):
        ul = self.Request("/room/roomSrv/PageFeedbackList", {
            "Page": {
                "Page": page,
                "PageSize": pagesize,
                "Time": {
                    "Start": 0,
                    "End": 0,
                },
                "Sum": False,
            },
            "Feedback": {}
        })
        return ul

    def PageSGL(self, page, pagesize):
        ul = self.Request("/room/roomSrv/PageSpecialGameList", {
            "Page": {
                "Page": page,
                "PageSize": pagesize,
                "Time": {
                    "Start": 0,
                    "End": 0,
                },
                "Sum": False,
            },
            "GameRecord": {}
        })
        return ul

    def RoomResultList(self, gtype):
        ul = self.Request("/room/roomSrv/RoomResultList", {
            "GameType": gtype,
        })
        return ul

    def GameResultList(self, rid):
        ul = self.Request("/thirteen/thirteenSrv/GameResultList", {
            "RoomID": rid,
        })
        return ul

    def AgentRoomList(self, page, gametyp):
        ul = self.Request("/room/roomSrv/GetAgentRoomList", {
            "Page": page,
            "GameType": gametyp,

        })
        return ul

    def CheckRoomExist(self):
        ul = self.Request("/room/roomSrv/CheckRoomExist", {
        })
        return ul

    def Shock(self, uid):
        ul = self.Request("/room/roomSrv/Shock", {
            "UserID": uid,
        })
        return ul

    def GetRoomRecovery(self, rid, gtype):
        ul = self.Request("/room/roomSrv/GetRoomRecovery", {
            "RoomID": rid,
            "GameType": gtype,
        })
        return ul

    def GetRoomRecoveryA(self):
        ul = self.Request("/room/roomSrv/GetRoomRecovery", {
        })
        return ul

    def ShuffleCard(self, rid, gtype):
        ul = self.Request("/room/roomSrv/ShuffleCard", {
        })
        return ul

    def SetBankerList(self, pwd):
        ul = self.Request("/room/roomSrv/SetBankerList", {
            "Password": pwd,
        })
        return ul

    def OutBankerList(self, pwd):
        ul = self.Request("/room/roomSrv/OutBankerList", {
            "Password": pwd,
        })
        return ul

    def GetRoomResult(self, rid):
        ul = self.Request("/room/roomSrv/GetRoomResultByID", {
            "RoomID": rid,
        })
        return ul

    def GetRoomRoundNow(self, rtype):
        ul = self.Request("/room/roomSrv/GetRoomRoundNow", {
            "RoomType": rtype,
        })
        return ul

    """
    充值相关操作
    """

    def Recharge(self, uid, diamond, orderCode, ctype):
        ul = self.Request("/bill/billSrv/Recharge", {
            "UserID": uid,
            "Diamond": diamond,
            "OrderID": orderCode,
            "CoinType": ctype,
        })
        return ul

    """
    用户相关操作
    """

    def RefreshUser(self):
        ul = self.Request("/user/UserSrv/RefreshUserCount", {
        })
        return ul

    def GetToken(self, uid):
        ul = self.Request("/user/UserSrv/GetToken", {
            "UserID": uid
        })
        return ul

    def CheckUser(self, UserID, Token):
        ul = self.Request("/user/UserSrv/CheckUser", {
            "UserID": UserID,
            "Token": Token
        })
        return ul

    def PageUserList(self, page, pagesize):
        now = int(time.time())
        ul = self.Request("/user/UserSrv/PageUserList", {
            "Page": {
                "Page": page,
                "PageSize": pagesize,
                "Time": {
                    "Start": now - 60 * 60 * 24,
                    "End": now,
                },
                "Sum": False,
            },
        })
        return ul

    def RRobot(self, count):
        ul = self.Request("/user/UserSrv/RegisterRobot", {
            "Count": count
        })
        return ul

    def TestClean(self):
        ul = self.Request("/room/roomSrv/TestClean", {
        })
        return ul

    def GetNotice(self, version):
        ul = self.Request("/notice/NoticeSrv/GetNotice", {
            "Versions": version,
        })
        return ul

    def CreateNotice(self, noticetype):
        ul = self.Request("/notice/NoticeSrv/CreateNotice", {
            "NoticeType": noticetype,
            "NoticeContent": "测试测试测试测试测试测试测试测试测试",
            "Status": 1,
            "StartAt": 1501655847,
            "EndAt": 1601655847,
        })
        return ul

    def UpdateNotice(self, noticetype):
        ul = self.Request("/notice/NoticeSrv/UpdateNotice", {
            "NoticeType": noticetype,
            "NoticeContent": "测试测试测试测试测试测试测试测试测试",
            "Status": 2,
            "StartAt": 1501655847,
            "EndAt": 1600655847,
            "Channel": "",
            "Version": "",
        })
        return ul

    def AllNotice(self):
        ul = self.Request("/notice/NoticeSrv/AllNotice", {
        })
        return ul

    def PageNoticeList(self, page, pagesize):
        ul = self.Request("/notice/noticeSrv/PageNoticeList", {
            "Page": {
                "Page": page,
                "PageSize": pagesize,
                "Time": {
                    "Start": 0,
                    "End": 0,
                },
                "Sum": False,
            },
            "Notice": {}
        })
        return ul

    def Share(self):
        ul = self.Request("/activity/ActivitySrv/Share", {
        })
        return ul

    def Invite(self, uid):
        ul = self.Request("/activity/ActivitySrv/Invite", {
            "UserID": uid,
        })
        return ul

    def ThirteenRecovery(self, rid):
        ul = self.Request("/thirteen/thirteenSrv/ThirteenRecovery", {
            "RoomID": rid,
        })
        return ul

    def WXLogin(self, core):
        ul = self.Request("/user/userSrv/WXLogin", {
            "Code": core,
            "Channel": "test",
            "MobileUuID": "23123123",
            "MobileModel": "dfsfsdfs",
            "MobileNetWork": "dwdqdwd",
            "MobileOs": "qqweqweeesd",
        })
        return ul

    def CreateNRoom(self, maxNumber, roundNumber, bType):
        ul = self.Request("/room/roomSrv/CreateRoom", {
            "RoundNumber": roundNumber,
            "MaxNumber": maxNumber,
            "GameType": 1002,
            "GameParam": "{\"Times\":1,\"BankerType\":%d}" % (bType)
        })
        return ul

    def CreateNRoomClub(self, maxNumber, roundNumber):
        ul = self.Request("/room/roomSrv/CreateRoom", {
            "RoundNumber": roundNumber,
            "MaxNumber": maxNumber,
            "GameType": 1002,
            "SubRoomType": 301,
            "RoomType": 3,
            "GameParam": "{\"Times\":3,\"BankerType\":2}",
            "SettingParam": "{\"ClubCoinRate\":2}",
        })
        return ul

    def GetBankerN(self, value):
        ul = self.Request("/niuniu/niuniuSrv/GetBanker", {
            "Key": value,
        })
        return ul

    def SetBetN(self, value):
        ul = self.Request("/niuniu/niuniuSrv/SetBet", {
            "Key": value,
        })
        return ul

    def SubmitCardN(self):
        ul = self.Request("/niuniu/niuniuSrv/SubmitCard", {
        })
        return ul

    def NiuniuResultListRequest(self, rid):
        ul = self.Request("/niuniu/niuniuSrv/GameResultList", {
            "RoomID": rid,
        })
        return ul

    def NiuniuRecovery(self, rid):
        ul = self.Request("/niuniu/niuniuSrv/NiuniuRecovery", {
            "RoomID": rid,
        })
        return ul

    def InviteUserInfo(self):
        ul = self.Request("/activity/ActivitySrv/InviteUserInfo", {
        })
        return ul

    def DayActive(self, page):
        ul = self.Request("/user/userSrv/DayActiveUserList", {
            "Page": page,
        })
        return ul

    def OnlineCount(self):
        ul = self.Request("/user/userSrv/GetUserOnlineCount", {
        })
        return ul

    def GetConfigs(self):
        ul = self.Request("/config/configSrv/GetConfigs", {
        })
        return ul

    def SetLocation(self, json):
        ul = self.Request("/user/userSrv/SetLocation", {
            "Json": json,
        })
        return ul

    def GetLocation(self):
        ul = self.Request("/room/roomSrv/GetRoomUserLocation", {
        })
        return ul

    def GetConfigsBeforeLogin(self, channel, version, mobileOs):
        ul = self.Request("/config/configSrv/GetConfigsBeforeLogin", {
            "Channel": channel,
            "Version": version,
            "MobileOs": mobileOs,
        })
        return ul

    def SetRegisterChannel(self, uiid, channel):
        ul = self.Request("/user/userSrv/SetRegisterChannel", {
            "UnionID": uiid,
            "RegistChannel": channel,
        })
        return ul

    """
    俱乐部相关操作
    """

    def PageClub(self, page, pagesize):
        ul = self.Request("/club/clubSrv/PageClub", {
            "Page": {
                "Page": page,
                "PageSize": pagesize,
                "Time": {
                    "Start": 0,
                    "End": 0,
                },
                "Sum": False,
            },
            "Club": {}
        })
        return ul

    def PageClubMember(self, page, pagesize):
        ul = self.Request("/club/clubSrv/PageClubMember", {
            "Page": {
                "Page": page,
                "PageSize": pagesize,
                "Time": {
                    "Start": 0,
                    "End": 0,
                },
                "Sum": False,
            },
            "ClubMember": {}
        })
        return ul

    def PageClubRoom(self, page, pagesize):
        ul = self.Request("/club/clubSrv/PageClubRoom", {
            "Page": {
                "Page": page,
                "PageSize": pagesize,
                "Time": {
                    "Start": 0,
                    "End": 0,
                },
                "Sum": False,
            },
            "Room": {}
        })
        return ul

    def UpdateClub(self, clubid):
        ul = self.Request("/club/clubSrv/UpdateClub", {
            "ClubID": clubid,
            "SettingParam": {
                "CostType": 2,
                "CostValue": 500,
                "ClubCoinBaseScore": 1000,
            },
            "Notice": "测试测试123",
        })
        return ul

    def UpdateClubB(self, clubid):
        ul = self.Request("/club/clubSrv/UpdateClub", {
            "ClubID": clubid,
            "Notice": "测试测试",
            "Status": 0,
        })
        return ul

    def RemoveClubMember(self, clubid, uid):
        ul = self.Request("/club/clubSrv/RemoveClubMember", {
            "ClubID": clubid,
            "UserID": uid,
        })
        return ul

    def ClubRecharge(self, amount, clubid, type):
        ul = self.Request("/club/clubSrv/ClubRecharge", {
            "ClubID": clubid,
            "Amount": amount,
            "AmountType": type,
        })
        return ul

    def CreateClub(self, clubname, creatorid, creatorproxy):
        ul = self.Request("/club/clubSrv/CreateClub", {
            "ClubName": clubname,
            "CreatorID": creatorid,
            "CreatorProxy": creatorproxy,
        })
        return ul

    def CreateClubMember(self, clubid, userid):
        ul = self.Request("/club/clubSrv/CreateClubMember", {
            "ClubID": clubid,
            "UserID": userid,
        })
        return ul

    def JoinClub(self, clubid):
        ul = self.Request("/club/clubSrv/JoinClub", {
            "ClubID": clubid,
        })
        return ul

    def LeaveClub(self, clubid):
        ul = self.Request("/club/clubSrv/LeaveClub", {
            "ClubID": clubid,
        })
        return ul

    def GetClub(self):
        ul = self.Request("/club/clubSrv/GetClub", {
        })
        return ul

    def GetClubByClubID(self,cid):
        ul = self.Request("/club/clubSrv/GetClubByClubID", {
            "ClubID": cid,
        })
        return ul

    def SetBlackList(self, clubid, uid):
        ul = self.Request("/club/clubSrv/SetBlackList", {
            "ClubID": clubid,
            "UserID": uid,
        })
        return ul

    def CancelBlackList(self, originID, targetID):
        ul = self.Request("/common/commonSrv/CancelBlackList", {
            "Type": 1,
            "OriginID": originID,
            "TargetID": targetID,
        })
        return ul

    def AddClubMemberClubCoin(self, clubid, uid, amount):
        ul = self.Request("/club/clubSrv/AddClubMemberClubCoin", {
            "ClubID": clubid,
            "UserID": uid,
            "Amount": amount,
        })
        return ul

    def ClubMemberOfferUpClubCoin(self, clubid, amount):
        ul = self.Request("/club/clubSrv/ClubMemberOfferUpClubCoin", {
            "ClubID": clubid,
            "Amount": amount,
        })
        return ul

    def PageClubJournal(self, page, pagesize, clubid, status):
        ul = self.Request("/club/clubSrv/PageClubJournal", {
            "Page": {
                "Page": page,
                "PageSize": pagesize,
                "Time": {
                    "Start": 0,
                    "End": 0,
                },
                "Sum": False,
            },
            "ClubID": clubid,
            "Status": status,
        })
        return ul

    def PageClubMemberJournal(self, page, pagesize):
        ul = self.Request("/club/clubSrv/PageClubMemberJournal", {
            "Page": {
                "Page": page,
                "PageSize": pagesize,
                "Time": {
                    "Start": 0,
                    "End": 0,
                },
                "Sum": False,
            },
        })
        return ul

    def UpdateClubJournal(self, cjid, clubid):
        ul = self.Request("/club/clubSrv/UpdateClubJournal", {
            "ClubJournalID": cjid,
            "ClubID": clubid,
        })
        return ul

    def GetClubMemberCoinRank(self, page, pagesize):
        ul = self.Request("/club/clubSrv/GetClubMemberCoinRank", {
            "Page": {
                "Page": page,
                "PageSize": pagesize,
                "Time": {
                    "Start": 0,
                    "End": 0,
                },
                "Sum": False,
            },
        })
        return ul

    def UpdateClubMemberStatus(self, clubid, uid, status):
        ul = self.Request("/club/clubSrv/UpdateClubMemberStatus", {
            "ClubID": clubid,
            "UserID": uid,
            "Status": status,
        })
        return ul

    def UpdateClubProxyID(self, proxyid, uid):
        ul = self.Request("/club/clubSrv/UpdateClubProxyID", {
            "CreatorID": uid,
            "CreatorProxy": proxyid,
        })
        return ul

    def GetClubsByMemberID(self):
        ul = self.Request("/club/clubSrv/GetClubsByMemberID", {
        })
        return ul

    def GetClubRoomLog(self, clubid):
        ul = self.Request("/club/clubSrv/GetClubRoomLog", {
            "ClubID": clubid,
        })
        return ul

    def PageBlackListMember(self, page, pagesize, clubid):
        ul = self.Request("/club/clubSrv/PageBlackListMember", {
            "Page": {
                "Page": page,
                "PageSize": pagesize,
                "Time": {
                    "Start": 0,
                    "End": 0,
                },
                "Sum": False,
            },
            "ClubMember": {
                "ClubID": clubid,
            },
        })
        return ul

    def CancelBlackList(self, clubid, uid):
        ul = self.Request("/club/clubSrv/CancelBlackList", {
            "ClubID": clubid,
            "UserID": uid,
        })
        return ul

    def CreateClubExamine(self, clubid, uid):
        ul = self.Request("/club/clubSrv/CreateClubExamine", {
            "ClubID": clubid,
            "UserID": uid,
        })
        return ul

    def UpdateClubExamine(self, clubid, uid, status):
        ul = self.Request("/club/clubSrv/UpdateClubExamine", {
            "ClubID": clubid,
            "UserID": uid,
            "Status": status,
        })
        return ul

    """
    斗地主相关操作
    """

    def CreateDRoom(self, roomType, maxNumber, roundNumber):
        ul = self.Request("/room/roomSrv/CreateRoom", {
            "RoundNumber": roundNumber,
            "MaxNumber": maxNumber,
            "GameType": 1003,
            "RoomType": roomType,
            "GameParam": "{\"BaseScore\":0}"
        })
        return ul

    def GetBanker(self, gid, type):
        ul = self.Request("/doudizhu/doudizhuSrv/GetBanker", {
            "GameID": gid,
            "GetBanker": type,
        })
        return ul

    def SubmitCardD(self, gid):
        ul = self.Request("/doudizhu/doudizhuSrv/SubmitCard", {
            "GameID": gid,
            "CardList": ['4_14'],
        })
        return ul

    def DoudizhuList(self, rid):
        ul = self.Request("/doudizhu/doudizhuSrv/GameResultList", {
            "RoomID": rid,
        })
        return ul

    """
    四张相关操作
    """

    def CRFour(self, roomType, maxNumber, roundNumber):  # , scoreType, betType
        ul = self.Request("/room/roomSrv/CreateRoom", {
            "RoundNumber": roundNumber,
            "MaxNumber": maxNumber,
            "GameType": 1004,
            "RoomType": roomType,
            "GameParam": '{\"ScoreType\":2,\"BetType\":2}'  # % (scoreType) % (betType)
        })
        return ul

    def SBFour(self, value):
        ul = self.Request("/fourcard/FourCardSrv/SetBet", {
            "Key": value,
        })
        return ul

    def SCFourA(self):
        ul = self.Request("/fourcard/FourCardSrv/SubmitCard", {
            "Head": ['4_7', '3_12'],
            "Tail": ['3_3', '5_21']
        })
        return ul

    def SCFourB(self):
        ul = self.Request("/fourcard/FourCardSrv/SubmitCard", {
            "Head": ['3_6', '4_10'],
            "Tail": ['1_6', '3_10']
        })
        return ul

    """
    邮件相关操作
    """

    def SendMail(self, mailid, uid):
        ul = self.Request("/mail/MailSrv/SendMail", {
            "MailSend": {
                "MailID": mailid,
                "SendID": 100000,
                "MailType": 1,
                "MailInfo": None,
            },
            "SendAll": 0,
            "Ids": [uid]
        })
        return ul

    def PagePlayerMail(self, page):
        ul = self.Request("/mail/mailSrv/PagePlayerMail", {
            "Page": page,
        })
        return ul

    def GetMailItems(self, logID):
        ul = self.Request("/mail/mailSrv/GetMailItems", {
            "LogID": logID,
        })
        return ul

    def GetAllMailItems(self):
        ul = self.Request("/mail/mailSrv/GetAllMailItems", {
        })
        return ul

    def PagePlayerMail(self, page):
        ul = self.Request("/mail/MailSrv/PagePlayerMail", {
            "Page": page,
        })
        return ul

    def ReadMail(self, logid):
        ul = self.Request("/mail/MailSrv/ReadMail", {
            "LogID": logid,
        })
        return ul

    def CMI(self):
        ul = self.Request("/mail/MailSrv/CreateMailInfo", {
            "MailInfo": {
                "MailID": 2001,
                "MailName": "测试物品邮件模板",
                "MailTitle": "测试物品邮件模板",
                "MailContent": "小吴是傻逼，快点领东西",
            },
            "ItemModelA": {
                "MainType": 100,
                "SubType": 1,
                "Count": 1000,
            },
            "ItemModelB": {
                "MainType": 100,
                "SubType": 2,
                "Count": 10,
            },
            "ItemModelC": {
                "MainType": 200,
                "SubType": 1,
                "ItemID": 1023,
                "Count": 1,
            },
        })
        return ul

    """
    金币场相关操作
    """

    def EnterGRoom(self, level, gtype):
        ul = self.Request("/goldroom/GoldRoomSrv/EnterRoom", {
            "Level": level,
            "GameType": gtype,
        })
        return ul

    def LeaveGRoom(self):
        ul = self.Request("/goldroom/GoldRoomSrv/LeaveRoom", {
        })
        return ul

    def SetGReady(self, pwd):
        ul = self.Request("/goldroom/GoldRoomSrv/SetReady", {
            "Password": pwd
        })
        return ul

    """
    两张相关操作
    """

    def CRTow(self, roomType, maxNumber, roundNumber):  # , scoreType, betType
        ul = self.Request("/room/roomSrv/CreateRoom", {
            "RoundNumber": roundNumber,
            "MaxNumber": maxNumber,
            "GameType": 1005,
            "RoomType": roomType,
            "GameParam": '{\"ScoreType\":2,\"BetType\":1}'  # % (scoreType) % (betType)
        })
        return ul

    def SBTow(self, value):
        ul = self.Request("/twocard/TwoCardSrv/SetBet", {
            "Key": value,
        })
        return ul

    def SCTow(self):
        ul = self.Request("/twocard/TwoCardSrv/SubmitCard", {
        })
        return ul

    def GetTwoRecovery(self, uid, rid):
        ul = self.Request("/twocard/TwoCardSrv/TwoCardRecovery", {
            "UserID": uid,
            "RoomID": rid,
        })
        return ul

    def TowGameResultList(self, rid):
        ul = self.Request("/twocard/TwoCardSrv/GameResultList", {
            "RoomID": rid,
        })
        return ul

    def CRTest(self, roomType, maxNumber, roundNumber):  # , scoreType, betType
        ul = self.Request("/room/roomSrv/CreateRoom", {
            "RoundNumber": roundNumber,
            "MaxNumber": maxNumber,
            "GameType": 1002,
            "RoomType": roomType,
            "GameParam": '{\"ScoreType\":2,\"BetType\":2}'  # % (scoreType) % (betType)
        })
        return ul
