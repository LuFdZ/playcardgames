# -*-coding:utf-8 -*-

import time
from client import Client

a = Client("aaaaaa","1001","1001")
b = Client("bbbbbb","1002","1002")
c = Client("cccccc","1003","1003")
d = Client("dddddd","1004","1004")
e = Client("eeeeee","1005","1005")

def Init(account):
    account.Login()
    account.Stream()

    time.sleep(0.5)

    account.SubscribeRegionMessage(1)
    account.SubscribeRegionChat(1)
    account.SubscribeThirteen()
    account.SubscribeNiu()
    account.SubscribeBill()
