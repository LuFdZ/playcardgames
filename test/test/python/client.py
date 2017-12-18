# -*-coding:utf-8 -*-
import httplib2
import json
import thread
import websocket
import threading
import time

Host = "http://192.168.1.76:8080"
WS_URL = "ws://192.168.1.76:8999/stream"

allLock = threading.RLock()


class Client(object):
    def __init__(self, name, opid, unid, host,wshost):
        self.token = ""
        self.user_id = 0
        self.user = None
        self.account = self.GenAccount(name, opid, unid)
        self.host = host
        self.pwd = "------"
        self.owner = 0  # 1房主，2非房主
        self.status = 0  # 1可加入
        self.lock = threading.RLock()
        self.wshost = wshost

    def Pwd(self):
        return self.pwd

    def Status(self):
        return self.status

    def GenAccount(self, name, opid, unid):
        return {
            "Username": name,
            "Password": "123456",
            "Nickname": name,
            "Email": name + "@x.com",
            "Icon": "12345567",
            "Sex": 1,
            "OpenID": opid,
            "UnionID": unid,
            "pwd": ""
        }

    def Request(self, path, data):
        req = httplib2.Http()
        path = self.host + path
        data = json.dumps(data)
        if self.lock.acquire():
            _, rsp = req.request(path, "POST", data, headers={"Token": self.token})
            #print("req: {%s} -d '{%s}' -H Token:{%s}" % (path, data, self.token))
            print("req: {%s} -H Token:{%s}" % (path, self.token))
            print("rsp:", rsp)
            self.lock.release()
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
            # """
            # websocket 连接上后，首先发送用户 Token 认证
            # """
            print("websocket connected")
            # send Token to login
            ws.send(self.token)

        def on_message(ws, msg):
            info = json.loads(msg)
            if allLock.acquire():
                print("websocket msg------", info["Type"])
                allLock.release()

            if self.lock.acquire():

                #info = json.loads(msg)
                # 创建房间
                if info["Type"] == "RoomCreate":
                    self.pwd = info["Data"]["Password"]
                    #print("create room:------------", self.pwd)
                    time.sleep(5.0)
                    self.SetReady(self.pwd)
                    self.status = 1

                if info["Type"] == "RoomExist":
                    self.pwd = info["Data"]["Room"]["Password"]
                    #print("RoomExist:------------", self.pwd)
                    time.sleep(5.0)
                    self.SetReady(self.pwd)
                if info["Type"] == "ThirteenGameStart":
                    self.status = 0
                    time.sleep(5.0)
                    self.SubmitCardT()
                    #print("SubmitCardT:------------", self.pwd)
                if info["Type"] == "RoomResult":
                    RN1 = info["Data"]["RoundNow"]
                    RN2 = info["Data"]["RoundNumber"]
                    ST = info["Data"]["Status"]
                    #print("Round:", RN1, "/", RN2)
                    if ST == 5:
                        if self.owner == 1:
                            time.sleep(5.0)
                            self.CreateTRoom(1, 2, 10)
                        #print("game over..")
                    else:
                        self.SetReady(self.pwd)
                        # print("RoomResult:------------",self.pwd)
                        # self.SetReady(self.pwd)

                self.lock.release()
        def on_error(ws, error):
            print("websocket error:", error)

        def on_close():
            print("websocket closed")

        websocket.enableTrace(True)
        self.ws = websocket.WebSocketApp(
            self.wshost,
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

    def CreateTRoom(self, roomType, maxNumber, roundNumber):
        self.owner = 1
        ul = self.Request("/room/roomSrv/CreateRoom", {
            "RoundNumber": roundNumber,
            "MaxNumber": maxNumber,
            "GameType": 1001,
            "RoomType": roomType,
            "GameParam": "{\"BankerAddScore\":2,\"Time\":30,\"Joke\":0,\"Times\":2,\"BankerType\":1}"
        })
        return ul

    def EnterRoom(self, Password):
        self.owner = 2
        self.pwd = Password
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

    def OutReady(self, Password):
        ul = self.Request("/room/roomSrv/OutReady", {
            "Password": Password
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

    def ClinetHearbeat(self):
        self.StreamSend("ClinetHearbeat", {})

    def Recharge(self, uid, diamond, orderCode):
        ul = self.Request("/bill/billSrv/Recharge", {
            "UserID": uid,
            "Diamond": diamond,
            "OrderID": orderCode,
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
        ul = self.Request("/user/UserSrv/PageUserList", {
            "Page": {
                "Page": page,
                "PageSize": pagesize,
                "Time": {
                    "Start": 0,
                    "End": 0,
                },
                "Sum": False,
            },
            "OpenID": "123123"
        })
        return ul

    def SubmitCardT(self):
        ul = self.Request("/thirteen/thirteenSrv/SubmitCard", {
            "Head": ['2_8', '2_5', '2_3'],
            "Middle": ['3_13', '1_13', '3_12', '3_11', '3_9'],
            "Tail": ['4_14', '4_5', '1_4', '4_3', '4_2']
        })
        return ul

    def SubmitCardT2(self):
        ul = self.Request("/thirteen/thirteenSrv/SubmitCard", {
            "Head": ['4_14', '4_13', '2_12'],
            "Middle": ['4_9', '3_9', '3_7', '2_7', '4_3'],
            "Tail": ['1_6', '2_6', '1_6', '1_4', '1_4']
        })
        return ul

    def TestClean(self):
        ul = self.Request("/room/roomSrv/TestClean", {
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

    def Renewal(self, pwd):
        ul = self.Request("/room/roomSrv/Renewal", {
            "Password": pwd,
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

    def Share(self):
        ul = self.Request("/activity/ActivitySrv/Share", {
        })
        return ul

    def Invite(self, uid):
        ul = self.Request("/activity/ActivitySrv/Invite", {
            "UserID": uid,
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
            "GameParam": "{\"Times\":3,\"BankerType\":%d}" % (bType)
        })
        return ul

    def GetBanker(self, value):
        ul = self.Request("/niuniu/niuniuSrv/GetBanker", {
            "Key": value,
        })
        return ul

    def SetBet(self, value):
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

    def AgentRoomList(self, page, gametyp):
        ul = self.Request("/room/roomSrv/GetAgentRoomList", {
            "Page": page,
            "GameType": gametyp,

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

    def UpdateClub(self, clubid, status):
        ul = self.Request("/club/clubSrv/UpdateClub", {
            "ClubID": clubid,
            "Status": status,
        })
        return ul

    def RemoveClubMember(self, clubid, uid, status):
        ul = self.Request("/club/clubSrv/UpdateClubMember", {
            "ClubID": clubid,
            "UserID": uid,
        })
        return ul

    def ClubRecharge(self, amount, clubid):
        ul = self.Request("/club/clubSrv/ClubRecharge", {
            "ClubID": clubid,
            "Amount": amount,
            "AmountType": 2,
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
