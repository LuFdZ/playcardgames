# -*-coding:utf-8 -*-

import time
import thread
from test import Test

def PT(start,end):

    for i in range(start,end):
        t = Test(i)
        #t.Ptest(i)
        try:
            thread.start_new_thread(t.CreateRoom, (i,))
        except Exception:
            print ("Error: unable to start thread")
        time.sleep(1)