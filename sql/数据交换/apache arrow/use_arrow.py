# -*- coding: utf-8 -*-
"""
    File Name: use_arrow
    Description: ""
    Author: haha.gan
    Date: 2020/6/23 18:04
"""
import pyarrow as pa

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


def write(data, name):
    batch = pa.record_batch(data, names=name)
    sink = pa.BufferedOutputStream()
    writer = pa.ipc.new_stream(sink, batch.schema)
    writer.write(batch)
    writer.close()
    return sink


def read(sink):
    buf = sink.getvalue()
    reader = pa.ipc.open_stream(buf)
    batches = [b for b in reader]
    return batches


if __name__ == '__main__':
    s = write([pa.array([1, 2, 3, 4]),pa.array(['foo', 'bar', 'baz', None])], ['s1', 's0'])
    b = read(s)
    print(b)
