# -*- coding: utf-8 -*-
"""
    File Name: use_arrow
    Description: ""
    Author: haha.gan
    Date: 2020/6/23 18:04
"""
import pyarrow as pa
from pyarrow import json as pa_json
import json
import time
import logging
import pickle

logging.basicConfig(level=logging.DEBUG)

LARGE_TIMES = 1000
SMALL_TIMES = 1000


def test():
    t1 = pa.int32()
    t2 = pa.string()
    t3 = pa.binary()
    t4 = pa.binary(10)
    t5 = pa.timestamp('ms')
    t6 = pa.list_(t1)

    f0 = pa.field('int32_field', t1)

    fields = [
        pa.field('s0', t1),
        pa.field('s1', t2),
        pa.field('s2', t4),
        pa.field('s3', t6)
    ]

    s0 = pa.struct(fields)
    s1 = pa.struct([('s0', t1), ('s1', t2), ('s2', t4), ('s3', t6)])

    pa.schema([('s0', t1), ('s1', t2), ('s2', t4), ('s3', t6)])


def cost_time(times=1):
    def decorate(func):
        def wrapper(*args, **kwargs):
            result = None
            cost = 0.0
            for i in range(times):
                start = time.time()
                result = func(*args, **kwargs)
                cost += time.time() - start
            # cost = round(cost/times, 4)
            cost /= times
            logging.debug("{0} Cost time: {1} ms".format(func, cost * 1000))
            return result

        return wrapper

    return decorate


@cost_time(SMALL_TIMES)
def test_json_file_to_arrow(f):
    return pa_json.read_json(f)


@cost_time(SMALL_TIMES)
def test_json_file_to_dict(f):
    with open(f, "r") as fin:
        return json.load(fin)


@cost_time(LARGE_TIMES)
def test_output_ipc(batch):
    sink = pa.BufferOutputStream()
    writer = pa.ipc.new_stream(sink, batch.schema)
    writer.write(batch)
    writer.close()

    return sink


@cost_time(LARGE_TIMES)
def test_input_ipc(buf):
    reader = pa.ipc.open_stream(buf)
    batches = [b for b in reader]
    # reader = pa.input_stream(buf)
    # batches = reader.read()
    return pa.Table.from_batches(batches)


@cost_time(LARGE_TIMES)
def test_serialize(data):
    context = pa.serialize(data)
    buf = context.to_buffer()
    return buf


@cost_time(LARGE_TIMES)
def test_deserialize(buf):
    return pa.deserialize(buf)


@cost_time(SMALL_TIMES)
def test_pickle_serialize(data):
    context = pickle.dumps(data)
    return context


@cost_time(SMALL_TIMES)
def test_pickle_deserialize(ctx):
    data = pickle.loads(ctx)
    return data


@cost_time(SMALL_TIMES)
def test_json_loads(ctx):
    return json.loads(ctx)


@cost_time(SMALL_TIMES)
def test_json_dumps(data):
    return json.dumps(data)


@cost_time(SMALL_TIMES)
def test_arrow_to_dict(batch):
    return batch.to_pydict()


@cost_time(SMALL_TIMES)
def test_dict_to_arrow(dict_object):
    return pa.Table.from_pydict(dict_object)


@cost_time(SMALL_TIMES)
def test_compress(buf, codec=None):
    return pa.compress(buf)


@cost_time(SMALL_TIMES)
def test_decompress(buf, *args, **kwargs):
    return pa.decompress(buf, *args, **kwargs)


@cost_time(SMALL_TIMES)
def test_compress_stream(batch):
    raw = pa.BufferOutputStream()
    with pa.CompressedOutputStream(raw, "lz4") as compressed:
        pa.serialize(batch).write_to(compressed)
    cdata = raw.getvalue()
    raw = pa.BufferReader(cdata)
    with pa.CompressedInputStream(raw, "lz4") as compressed:
        tmp = pa.deserialize(compressed.read())


def client(ctx):
    import socket

    serverName = '127.0.0.1'
    serverPort = 11000
    ADDR = (serverName, serverPort)

    clientSocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    clientSocket.connect(ADDR)

    clientSocket.send(ctx)
    clientSocket.send(ctx)
    clientSocket.close()


if __name__ == '__main__':
    import os

    test_file = "test1.json"
    test_file = os.path.join("practice", "json_data", test_file)

    # logging.info("包含文件打开的，json数据转换:")
    # test_json_file_to_arrow(test_file)
    # test_json_file_to_dict(test_file)

    # batch = pa_json.read_json(test_file)
    with open(test_file) as fin:
        raw_dict = json.load(fin)
        raw_dict = {'hits': raw_dict['hits']['hits'] * 1}
        batch = pa.Table.from_pydict(raw_dict)

    # logging.info("arrow 序列化反序列化:")
    # buf = test_serialize(batch)
    # data = test_deserialize(buf)

    raw_bytes = str.encode(json.dumps(raw_dict))
    print("raw txt: ", len(raw_bytes))
    print("Table: ", batch.nbytes)
    buf = pa.serialize(batch).to_buffer()
    print("serialize buf: ", len(buf.to_pybytes()))

    com_buf = pa.compress(buf, codec='gzip')
    com_txt = pa.compress(raw_bytes, codec='gzip')

    print("compressed raw txt", len(com_txt.to_pybytes()))
    print("compress buf: ", len(com_buf.to_pybytes()))
    print(buf.to_pybytes())
    print(raw_bytes)

    # array = batch.to_batches()[0][0]
    # field = ['_id', '_index', '_score', '_source', '_type']
    # sum = 0
    # array_size = 0
    # for i in field:
    #     tmp = array.field(i)
    #     print('-'*50)
    #     print(i, tmp.nbytes)
    #     array_size += tmp.nbytes
    #     tmp = pa.serialize(tmp).to_buffer().to_pybytes()
    #     print(i,tmp.__sizeof__())
    #     tmp = pa.compress(tmp)
    #     print(i, tmp.size)
    #     sum += tmp.size
    # print(sum, array_size)

    # logging.info("arrow 压缩buf")
    # compress = test_compress(buf)
    # decompress = test_decompress(compress, decompressed_size=buf.size)
    #
    # logging.info("arrow 压缩流 处理buf")
    # test_compress_stream(buf)
    #
    # with open(test_file) as fin:
    #     batch1 = json.load(fin)
    #
    # logging.info("arrow 序列化反序列化dict:")
    # buf = test_serialize(batch1)
    # data = test_deserialize(buf)
    #
    # logging.info("python pickle 序列化反序列化:")
    # ctx = test_pickle_serialize(batch1)
    # data1 = test_pickle_deserialize(ctx)
    #
    # logging.info("python json模块 序列化反序列化:")
    # ctx = test_json_dumps(batch1)
    # data2 = test_json_loads(ctx)
    # #
    # logging.info("arrow ipc模块 写入与读取:")
    # ctx = pa.serialize(batch)
    # sink = test_output_ipc(batch)
    # t = sink.getvalue()
    # batch2 = test_input_ipc(t)
    #
    # logging.info("arrow table与dict对象转换:")
    # d = test_arrow_to_dict(batch2)
    # t = test_dict_to_arrow(d)

    # socket 传输
    # ctx = pa.serialize(batch).to_buffer()
    # client(ctx)

    pause = True
