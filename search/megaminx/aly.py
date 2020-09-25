import matplotlib.pyplot as plt
import random
import sys
import psutil
import time
import sqlite3
import argparse


class Analyst(object):
    def __init__(self):
        self.conn = sqlite3.connect("reports/status.db")

    def analysis_io(self):
        cmd = "select * from IO"
        result = self.conn.cursor().execute(cmd)

        x = list()
        read_count = list()
        read_bytes = list()
        write_count = list()
        write_bytes = list()
        read_chars = list()
        write_chars = list()
        for i in result:
            x.append(i[0])
            read_count.append(i[1])
            read_bytes.append(i[2])
            write_count.append(i[3])
            write_bytes.append(i[4])
            read_chars.append(i[5])
            write_chars.append(i[6])

        plt.figure(figsize=(18, 10))
        plt.plot(x, read_count, label="read_count")
        plt.plot(x, read_bytes, label="read_bytes")
        plt.plot(x, write_count, label="write_count")
        plt.plot(x, write_bytes, label="write_bytes")
        plt.plot(x, read_chars, label="read_chars")
        plt.plot(x, write_chars, label="write_chars")
        plt.legend()
        plt.savefig("reports/io.png")

    def analysis_cpu(self):
        cmd = "select * from CPU"
        result = self.conn.cursor().execute(cmd)
        x = list()
        user = list()
        system = list()
        iowait = list()
        children_user = list()
        children_system = list()
        percent = list()

        for i in result:
            x.append(i[0])
            user.append(i[1])
            system.append(i[2])
            iowait.append(i[3])
            children_user.append(i[4])
            children_system.append(i[5])
            percent.append(i[6])

        plt.figure(figsize=(18, 10))
        plt.plot(x, user, label="user")
        plt.plot(x, system, label="system")
        plt.plot(x, iowait, label="iowait")
        plt.plot(x, children_user, label="children_user")
        plt.plot(x, children_system, label="children_system")
        plt.legend()
        plt.savefig("reports/cpu.png")

        plt.figure(figsize=(18, 10))
        plt.plot(x, percent, label="percent")
        plt.legend()
        plt.savefig("reports/cpu_percent.png")

    def analysis_mm(self):
        cmd = "select * from MEMORY"
        result = self.conn.cursor().execute(cmd)
        x = list()
        rss = list()
        vms = list()
        shared = list()
        text = list()
        lib = list()
        data = list()
        dirty = list()

        for i in result:
            x.append(i[0])
            rss.append(i[1])
            vms.append(i[2])
            shared.append(i[3])
            text.append(i[4])
            lib.append(i[5])
            data.append(i[6])
            dirty.append(i[7])

        plt.figure(figsize=(18, 10))
        plt.plot(x, rss, label="rss")
        plt.plot(x, vms, label="vms")
        plt.plot(x, shared, label="shared")
        plt.plot(x, text, label="text")
        plt.plot(x, lib, label="lib")
        plt.plot(x, data, label="data")
        plt.plot(x, dirty, label="dirty")

        plt.legend()
        plt.savefig("reports/mm.png")

    def reports(self):
        cur = self.conn.cursor()
        cpu = "select avg(percent) from cpu where percent != 0;"
        result = cur.execute(cpu).fetchone()[0]
        print("----------CPU----------")
        print("Avg cpu percent: {0}%".format(result))

        r = cur.execute(
            "select * from cpu ORDER BY TIME DESC LIMIT 1;").fetchone()
        print("user: {0}\nsystem: {1}\niowait: {2}".format(
            r[1], r[2], r[3]))

        r = cur.execute(
            "select avg(rss)/1024/1024,avg(vms)/1024/1024,avg(data)/1024/1024 from MEMORY;").fetchone()
        print("----------MEMORY----------")
        print("Avg rss: {0}\nvms: {1}\ndata: {2}".format(r[0], r[1], r[2]))

        r = cur.execute(
            "select read_count, read_bytes/1024.0/1024.0, write_count, write_bytes/1024.0/1024.0,\
                 read_chars/1024.0/1024.0, write_chars/1024.0/1024.0 from IO ORDER BY TIME DESC LIMIT 1;").fetchone()
        print("----------IO----------")
        print("read_count: {0}\nread_bytes: {1}\nwrite_count: {2}\nwrite_bytes: {3}\n".format(
            r[0], r[1], r[2], r[3]))
        print("read_chars: {0}\nwrite_chars: {1}".format(r[4], r[5]))


class Sqlite3Writer(object):
    def __init__(self, path="reports/status.db"):
        super().__init__()
        self.conn = sqlite3.connect(path)
        self.create_io()
        self.create_cpu()
        self.create_mm()

    def __del__(self):
        self.conn.close()

    def create_io(self):
        cur = self.conn.cursor()
        cur.execute('''
        CREATE TABLE IF NOT EXISTS IO
        (
            time REAL PRIMARY KEY,
            read_count INTEGER,
            read_bytes INTEGER,
            write_count INTEGER,
            write_bytes INTEGER,
            read_chars INTEGER,
            write_chars INTEGER
        )
        ''')
        self.conn.commit()

    def create_cpu(self):
        cur = self.conn.cursor()
        cur.execute('''
        CREATE TABLE IF NOT EXISTS CPU
        (
            time REAL PRIMARY KEY,
            user INTEGER,
            system INTEGER,
            iowait INTEGER,
            children_user INTEGER,
            children_system INTEGER,
            percent REAL
        )
        ''')
        self.conn.commit()

    def create_mm(self):
        cur = self.conn.cursor()
        cur.execute('''
        CREATE TABLE IF NOT EXISTS MEMORY
        (
            time REAL PRIMARY KEY,
            rss INTEGER,
            vms INTEGER,
            shared INTEGER,
            text INTEGER,
            lib INTEGER,
            data INTEGER,
            dirty INTEGER
        )
        ''')
        self.conn.commit()

    def write_io(self, timestamp, io):
        cur = self.conn.cursor()
        timestamp = time.time()
        cmd = "INSERT INTO IO VALUES ({0},{1},{2},{3},{4},{5},{6})".format(
            timestamp, io.read_count, io.read_bytes, io.write_count,
            io.write_bytes, io.read_chars, io.write_chars
        )
        cur.execute(cmd)
        self.conn.commit()

    def write_cpu(self, timestamp, cpu_times, percent):
        cur = self.conn.cursor()
        timestamp = time.time()
        cmd = "INSERT INTO CPU VALUES ({0},{1},{2},{3},{4},{5},{6})".format(
            timestamp, cpu_times.user, cpu_times.system, cpu_times.iowait,
            cpu_times.children_user, cpu_times.children_system, percent
        )
        cur.execute(cmd)
        self.conn.commit()

    def write_mm(self, timestamp, mm):
        cur = self.conn.cursor()
        timestamp = time.time()
        cmd = "INSERT INTO MEMORY VALUES ({0},{1},{2},{3},{4},{5},{6},{7})".format(
            timestamp, mm.rss, mm.vms, mm.shared, mm.text, mm.lib,
            mm.data, mm.dirty
        )
        cur.execute(cmd)
        self.conn.commit()


def metric_process(pid):
    p = psutil.Process(pid)
    writer = Sqlite3Writer()

    while p.is_running:
        try:
            now = time.time()
            # https://stackoverflow.com/questions/3633286/what-do-the-counters-in-proc-pid-io-mean
            io = p.io_counters()

            # https://stackoverflow.com/questions/556405/what-do-real-user-and-sys-mean-in-the-output-of-time1
            cpu_times = p.cpu_times()
            cpu_peercent = p.cpu_percent(interval=None)

            # memory_info
            mm = p.memory_info()

            writer.write_io(now, io)
            writer.write_cpu(now, cpu_times, p.cpu_percent())
            writer.write_mm(now, mm)

            time.sleep(1)
        except psutil.NoSuchProcess:
            print("Processs stoped...")
            break


def graph():
    # grpah
    aly = Analyst()
    aly.analysis_io()
    aly.analysis_cpu()
    aly.analysis_mm()
    aly.reports()


if __name__ == "__main__":
    cmd = sys.argv[1].strip()
    # cmd = ""

    # parser = argparse.ArgumentParser(
    #     description="Metric process or analysis metric result")
    # subparsers = parser.add_subparsers()
    # parse_metric = subparsers.add_parser("metric")
    # parse_metric.add_argument("-o", "--out", type=str,
    #                           help="保存结果的目录绝对路径，已存在不会覆盖")
    # parse_metric.add_argument("-p", "--pid", nargs=1,
    #                           type=int, help="监控进程的PID")

    # def metric(args):
    #     pid = args.p
    #     print(pid)
    #     metric_process(int(pid))

    #     graph()

    # parse_metric.set_defaults(func=metric)
    if cmd == "metric":
        pid = sys.argv[2]
        print(pid)
        metric_process(int(pid))

        graph()

    elif cmd == "analysis":
        graph()
    else:
        # analysis_vector()
        print("cmd only support 'metric <pid>' and 'analysis'")
