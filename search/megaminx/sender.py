#!/usr/bin/python
# -*- coding: UTF-8 -*-
import argparse
import socket
import _thread
import sys
import threading
import multiprocessing
import os


class Tcp(threading.Thread):
    def __init__(self, path, ip, port):
        threading.Thread.__init__(self)
        self._path = path
        self.ip = ip
        self.port = int(port)

    def run(self):

        client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        client.connect((self.ip, self.port))

        for log in open(self._path, "r"):
            client.send(log.encode('utf-8'))

        client.close()


def tcp(ip, port, path):
    try:
        client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        client.connect((ip, port))
        print("send file: ", path)

        for log in open(path, "r"):
            client.send(log.encode('utf-8'))

        client.close()
    except Exception as e:
        print(e)
        print("send file: ", path)
        return

    print("send file eof: ", path)


def test_io(path):
    old_path = os.path.split(path)
    new_path = os.path.join("tmp", old_path[-1])
    print(new_path)

    try:
        with open(new_path, "w") as fout:
            for log in open(path, "r"):
                fout.write(log+"\n")
        print("end: ", new_path)
    except Exception as e:
        print(e)
        return


if __name__ == "__main__":
    muti = list()
    pool = multiprocessing.Pool(8)
    for i in sys.argv[1:]:
        muti.append(pool.apply_async(
            tcp, args=("127.0.0.1", 5140, i,)))

        # muti.append(pool.apply_async(
        #     test_io, args=(i,)))

    pool.close()
    pool.join()
