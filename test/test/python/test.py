#Register
#Login
#Stream
#CreateTRoom/SetReady
#EnterRoom\SetReady
#SubmitCardT
# -*-coding:utf-8 -*-

import time
import thread
import threading
import random
from client import Client
	
# if __name__ == "__main__":
    # a = Test()
    # print("************************** Test begin **************************")
    # for i in range(2):
    #     thread.start_new_thread(Test.CreateRoom()(i), ())
    # print("************************** Test over ***************************")

#Host = "http://192.168.1.76:8080"
#WS_URL = "ws://192.168.1.76:8999/stream"
listHost = ["http://111.230.51.146:8888", "http://111.230.51.228:8888", "http://118.89.110.172:8888",
            "http://47.100.19.126:8888","http://101.132.141.6:8888","http://101.132.34.210:8888","http://139.196.86.65:8888"];
listWS = ["ws://111.230.51.146:8888/stream", "ws://111.230.51.228:8888/stream", "ws://118.89.110.172:8888/stream",
          "ws://47.100.19.126:8888/stream","ws://101.132.141.6:8888/stream","ws://101.132.34.210:8888/stream","ws://139.196.86.65:8888/stream"];

listHostLocal = ["http://192.168.1.76:8888", "http://192.168.1.76:8080", "http://192.168.1.76:8080",
                 "http://192.168.1.76:8080","http://192.168.1.76:8080","http://192.168.1.76:8080","http://192.168.1.76:8080"];
listWSLocal = ["ws://192.168.1.76:8888/stream", "ws://192.168.1.76:8080/stream", "ws://192.168.1.76:8080/stream",
               "ws://192.168.1.76:8080/stream","ws://192.168.1.76:8080/stream","ws://192.168.1.76:8080/stream","ws://192.168.1.76:8080/stream"];

def Init(account):
    account.Login()
    account.Stream()
    #time.sleep(0.5)

    # account.SubscribeRegionMessage(1)
    # account.SubscribeRegionChat(1)
    # account.SubscribeThirteen()
    # account.SubscribeNiu()
    # account.SubscribeBill()


class Test(object):
    def __init__(self,index):
        self.index= index
        self.lock = threading.RLock()
        name = '************* New Test %d *************' % self.index
        print(name)


    def Ptest(self,index):
        #self.CreateRoom(times)
        # print("************************** Test begin **************************")
        # for i in range(times):
        try:
            thread.start_new_thread(self.CreateRoom, (index,))
        except Exception:
            print ("Error: unable to start thread")
        # print("************************** Test over ***************************")

    def CreateRoom(self,index):
        if self.lock.acquire():
            rd = random.randint(0,6)
            key = index *4
            name = 'IPRANDOM%d' % rd
            print(name)
            idA = 'Tuser%d' % (key+1)
            userA = Client(idA,idA,idA,listHost[rd],listWS[rd])
            idB = 'Tuser%d' % (key+2)
            userB = Client(idB,idB,idB,listHost[rd],listWS[rd])
            idC = 'Tuser%d' % (key+3)
            userC = Client(idC,idC,idC,listHost[rd],listWS[rd])
            idD = 'Tuser%d' % (key+4)
            userD = Client(idD,idD,idD,listHost[rd],listWS[rd])


            Init(userA)
            Init(userB)
            Init(userC)
            Init(userD)
            time.sleep(2.0)

            userA.CreateTRoom(1,4,20)
            time.sleep(2.0)
            if userA.Pwd()== "------":
                print ("Error: Create Room Err")
                return
            time.sleep(4.0)
            #userA.CheckRoomExist()
            time.sleep(1.0)
            userB.EnterRoom(userA.Pwd())
            time.sleep(1.0)
            userC.EnterRoom(userA.Pwd())
            time.sleep(1.0)
            userD.EnterRoom(userA.Pwd())
            self.lock.release()


        while True:
            time.sleep(6.0)
            if userA.Status() == 1:
                userB.EnterRoom(userA.Pwd())
                time.sleep(5)
                userC.EnterRoom(userA.Pwd())
                time.sleep(6)
                userD.EnterRoom(userA.Pwd())
                time.sleep(7)