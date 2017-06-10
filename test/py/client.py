# -*-coding:utf-8 -*-

import httplib2
import json
import time
import websocket
import thread

Host = "http://127.0.0.1:8080"


class Client(object):
    def __init__(self, name, host=Host):
        self.token = ""
        self.user_id = 0
        self.user = None
        self.account = self.GenAccount(name)
        self.host = host

    def GenAccount(self, name):
        return {
            "Username": name,
            "Password": name,
            "Nickname": name,
            "Email": name + "@x.com",
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

    def GetProperty(self):
        """
        获得用户属性信息
        :return: 用户属性
        """
        return self.Request("/user/UserSrv/GetProperty", self.account)

    def AddFriend(self, friendid):
        """
        添加
        :return: 朋友列表
        """
        ul = self.Request("/friend/FriendSrv/AddFriend", {
            "FriendID": friendid
        })
        return ul

    def DelFriend(self,friendid):
        """
        删除
        :return:朋友列表
        """
        ul = self.Request("/friend/FriendSrv/DelFriend",{
            "FriendID": friendid
        })
        return ul

    def FriendList(self):
        """
        查询
        :return:朋友列表
        """
        ul = self.Request("/friend/FriendSrv/FriendList",{
        })
        return ul

    def ItemList(self):
        """
        列出玩家所有物品
        """
        return self.Request("/store/StoreSrv/ItemList", {})

    def AddRegion(self):
        """
        添加区域
        :return: 添加结果
        """
        return self.Request("/region/RegionSrv/AddRegion", {
            "RegionName" : "111111",
            "Description" : "new brige 1",
            "Level" : 10,
            "ProfitTime" : 5,
            "Status" : 1
        })

    def AddRegionEvent(self):
        """
        添加区域事件
        """
        return self.Request("/region/RegionSrv/AddRegionEvent", {
            "EventName" : "one event",
            "Description" : "test event description",
            "Rate" : 85,
            "ValueStart" : 500,
            "ValueEnd" : 50000,
            "ValueType" : 1,
            "RegionID" : 1
        })

    def RegionList(self):
        """
        列出所有区域
        """
        return self.Request("/region/RegionSrv/RegionList", {})

    def OpenRegion(self, rid, uid):
        """
        为指定用户开通区域
        """
        return self.Request("/region/RegionSrv/OpenRegion", {
            "RegionID" : rid,
            "UserID" : uid
        })

    def EnterRegion(self, rid):
        """
        进入指定区域
        """
        return self.Request("/region/RegionSrv/EnterRegion", {
            "RegionID" : rid
        })

    def LeaveRegion(self):
        """
        退出指定区域
        """
        return self.Request("/region/RegionSrv/LeaveRegion", {})

    def RegionChat(self, msg):
        """
        地域聊天
        """
        return self.Request("/chat/ChatSrv/Chat", {
            "Message" : msg,
            "Type" : 2
        })

    def Stream(self):

        def on_open(ws):
            """
            websocket 连接上后，首先发送用户 Token 认证
            """
            print "websocket connected"
            # send Token to login
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
            on_message = on_message,
            on_error = on_error,
            on_close = on_close,
            on_open = on_open,
        )

        thread.start_new_thread(self.ws.run_forever, ())

    def StreamSend(self, method, args):
        data = {
            "Method": method,
            "Args": args
        }

        b = json.dumps(data)
        self.ws.send(b)

    def SubscribeRegionMessage(self, region_id):
        self.StreamSend("SubscribeRegionMessage", region_id)

    def SubscribeChat(self):
        self.StreamSend("SubscribeChatMessage", {})

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

    def ListConfig(self):
        ul = self.Request("/activity/ActivitySrv/ListConfig", {})

    def UpdateConfig(self):
        ul = self.Request("/activity/ActivitySrv/UpdateConfig",{
            "COnfigID":10001,
            "Description": "十二星座json活动配置",
            "Parameter": json.dumps({
                "StartDurationInday": "0s",
                "EndEDurationInDay": "1439m",
                "LifeDuration": "1m",
                "InitialGoldPool": 60000,
                "Rate": 0.1
            })
        });

    def GetConfigByID(self, configid):
        ul = self.Request("/activity/ActivitySrv/GetConfigByID", {
          "ID":configid
        })

    def ZodiacRoundList(self, code="", status=0):
        """
        星座查询
        """
        ul = self.Request("/zodiac/ZodiacSrv/RoundList",{
            "Code": code,
            "Staatus": status,
        })
        return ul

    def AddZodiacBet(self, code, zodiac={}):
        """
        星座下注
        """
        ul = self.Request("/zodiac/ZodiacSrv/Bet",{
            "Code" : code,
            "Zodiac": zodiac
        })
        return ul

    def ZodiacRoundAllList(self,code):
        """
        星座查询
        """
        ul = self.Request("/zodiac/ZodiacSrv/ZodiacRoundAllList",{
            "Code": code
        })
        return ul


    def ZodiacRoundFix(self,code,result1,result2,result3):
        """
        星座结果更改
        """
        ul = self.Request("/zodiac/ZodiacSrv/ZodiacRoundFix",{
            "Code" : code,
            "Result1" : result1,
            "Reuslt2" : result2,
            "Result3" : result3,
        })
        return ul

    def SubscribeZodiacUpdate(self):
        self.StreamSend("SubscribeZodiacMessage",None)


    def PageZodiacList(self):
        """
        星座查询
        """
        ul = self.Request("/zodiac/ZodiacSrv/PageZodiacList",{
            "Code" :"",
            "Type" : 1,
            "Status" : 1,
            "UserID":0,
            "Page" :{
                "Page":0,
                "PageSize":15,
                "Time":{
                    "End":0,
                    "Start":0,
                }
            },
        })
        return ul

    def ZRList(self,code,status):
        ul = self.Request("/zodiac/ZodiacSrv/ZodiacRoundList",{
            "Code":code,
            "Status":status
        })
        return ul
