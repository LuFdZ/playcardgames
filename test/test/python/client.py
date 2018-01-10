# -*-coding:utf-8 -*-
import httplib2
import json
import websocket
import thread
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
            #print("req: {%s} -H Token:{%s}" % (path, self.token))
            print("UserID:{%s} req: {%s}" % (self.user_id ,path))
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
                    #time.sleep(1.0)
                    self.SetReady(self.pwd)
                    self.status = 1

                if info["Type"] == "RoomExist":
                    self.pwd = info["Data"]["Room"]["Password"]
                    #print("RoomExist:------------", self.pwd)
                    #time.sleep(1.0)
                    self.SetReady(self.pwd)
                if info["Type"] == "ThirteenGameStart":
                    self.status = 0
                    #time.sleep(1.0)
                    self.SubmitCardT()
                    #print("SubmitCardT:------------", self.pwd)
                if info["Type"] == "RoomResult":
                    RN1 = info["Data"]["RoundNow"]
                    RN2 = info["Data"]["RoundNumber"]
                    ST = info["Data"]["Status"]
                    print("Round:", RN1, "/", RN2)
                    if ST == 5:
                        if self.owner == 1:
                            time.sleep(1.0)
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


    def SetReady(self, Password):
        ul = self.Request("/room/roomSrv/SetReady", {
            "Password": Password
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

