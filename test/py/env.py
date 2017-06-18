# -*-coding:utf-8 -*-

import time
from client import Client

a = Client("aaaaaa")
b = Client("bbbbbb")
c = Client("cccccc")

def Init(account):
    account.Login()
    account.Stream()

    time.sleep(0.5)

    account.SubscribeRegionMessage(1)
    account.SubscribeRegionChat(1)
