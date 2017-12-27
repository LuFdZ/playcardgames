# -*-coding:utf-8 -*-

import time
import threading
import thread
from test import Test
from time import sleep, ctime

def PT(start,end):
    print 'Starting at:', ctime()
    threads = []
    for i in range(start,end):
        t = Test(i)
        #t.Ptest(i)
        try:
            thread.start_new_thread(t.CreateRoom, (i,))
        except Exception:
            print ("Error: unable to start thread")

        # t = Test(i)
        # th = threading.Thread(target=t.CreateRoom,args=(i,))
        # th.daemon = False
        # th.start()
        time.sleep(0.5)