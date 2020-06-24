序列化和进程间通信(IPC).md

# 序列化和进程通信(IPC)
序列化数据的基本单元在列式存储中称为"record batch"。从语义上讲，一个"record batch"是一组有序的数组集合，称为数组字段，每个数组的长度彼此相同，但数据类型可能不同。
record batch 的字段名和类型组成了 batch的schema

arrow的协议支持将record batch序列化为以二进制再提的流式数据，并且在不需要内存复制的情况下可以从流数据中进行反序列化出record batch
IPC协议利用了以下属性：
* Schema
* RcordBatch
* DictionaryBatch

## 1. 数据封装
为了进程间通信定义"封装"消息格式。在仅检查消息元数据，而不需要复制或移动任何实际数据，可以将消息反序列化位内存中的arrow数组对象
封装格式如下：
* 32位的连续指示符。 值0xFFFFFFFF表示一个有效信息
* 32位小端长度的prefix表明元数据大小
* 使用[Message.fbs](https://github.com/apache/arrow/blob/master/format/Message.fbs)中定义的类型的消息元数据
* 字节填充
* 消息正文，长度位8个字节的倍数
```
<metadata_size: int32>
<metadata_flatbuffer: bytes>
<padding>
<message body>
```
metadata_size的大小包括Message额外的填充。
metadata_flatbuffer包含一个序列化的Message Flatbuffer值，包含以下：
* 版本号
* 特殊的Message类型值(one of Schema, RecordBatch, or DictionaryBatch)
* 消息体大小
* custom_metadata 字段，用于支持任意应用支持的元数据

当读取输入流式，首先对Message元数据进行解析和验证，从而获得body大小，然后读取body

## 2. Schema message
[Schema.fbs](https://github.com/apache/arrow/blob/master/format/Schema.fbs)文件中包含了内置逻辑类型和scheme元数据类型的定义。
schema数据结构如下，其不包含任何数据，值包含类型元数据
```
table Schema {

  /// endianness of the buffer
  /// it is Little Endian by default
  /// if endianness doesn't match the underlying system then the vectors need to be converted
  endianness: Endianness=Little;

  fields: [Field];
  // User-defined metadata
  custom_metadata: [ KeyValue ];
}


table Field {
  /// Name is not required, in i.e. a List
  name: string;

  /// Whether or not this field can contain nulls. Should be true in general.
  nullable: bool;

  /// This is the type of the decoded value if the field is dictionary encoded.
  type: Type;

  /// Present only if the field is dictionary encoded.
  dictionary: DictionaryEncoding;

  /// children apply only to nested data types like Struct, List and Union. For
  /// primitive types children will have length 0.
  children: [ Field ];

  /// User-defined metadata
  custom_metadata: [ KeyValue ];
}
```
其中的Field类型包含一个单独数组的元数据。主要包含：
* 字段名
* 字段的逻辑类型
* 该字段在语义上是否可为空
* Field嵌套类型的子值集合
* dictionary指明字段属性是否是dictionary-encoded。如果是，分配字典'id'以允许后续字典IPC消息与适当的字段进行匹配

## 3. RecordBatch Message
一条RecordBatch消息包含由shcema决定的、与物理内存布局对应的真实数据缓存。消息的元数据提供了每一条缓存的地址和大小，从而允许使用指针算法重建数组数据结构，无需进行内存复制。

Record batch的序列化格式如下：
* MessageHeader,被定义为RecordBatch
* body，首尾相连的、适当填充确保8字节对齐的内存缓存

MessageHeader包含：
* 每个字段的长度和空值计数
* 每个 recode batch 的body中组成buffer的长度和内存偏移量
```
table RecordBatch {
  /// number of records / rows. The arrays in the batch should all have this
  /// length
  length: long;

  /// Nodes correspond to the pre-ordered flattened logical schema
  nodes: [FieldNode];

  /// Buffers correspond to the pre-ordered flattened buffer tree
  ///
  /// The number of buffers appended to this list depends on the schema. For
  /// example, most primitive arrays will have 2 buffers, 1 for the validity
  /// bitmap and 1 for the values. For struct arrays, there will only be a
  /// single buffer for the validity (nulls) bitmap
  buffers: [Buffer];

  /// Optional compression of the message body
  compression: BodyCompression;
}


struct FieldNode {
  /// The number of value slots in the Arrow array at this level of a nested
  /// tree
  length: long;

  /// The number of observed nulls. Fields with null_count == 0 may choose not
  /// to write their physical validity bitmap out as a materialized buffer,
  /// instead setting the length of the bitmap buffer to 0.
  null_count: long;
}

struct Buffer {
  /// The relative offset into the shared memory page where the bytes for this
  /// buffer starts
  offset: long;

  /// The absolute length (in bytes) of the memory buffer. The memory is found
  /// from offset (inclusive) to offset + length (non-inclusive). When building
  /// messages using the encapsulated IPC message, padding bytes may be written
  /// after a buffer, but such padding bytes do not need to be accounted for in
  /// the size here.
  length: long;
}
```

### 数据序列化和反序列化
1. 首先会对schema进行深度优先便利扁平化扩展出field和buffuers
```
col1: Struct<a: Int32, b: List<item: Int64>, c: Float64>
col2: Utf8
```
2. 扁平化处理后如下
```
FieldNode 0: Struct name='col1'
    FieldNode 1: Int32 name='a'
    FieldNode 2: List name='b'
        FieldNode 3: Int64 name='item'
    FieldNode 4: Float64 name='c'
FieldNode 5: Utf8 name='col2'
```

```
FieldNode 0: Struct name='col1'
FieldNode 1: Int32 name='a'
FieldNode 2: List name='b'
FieldNode 3: Int64 name='item'
FieldNode 4: Float64 name='c'
FieldNode 5: Utf8 name='col2'
```

3. 那么当buffer生成时就会根据和schema相应的处理方式
```
buffer 0: field 0 validity          // col1
    buffer 1: field 1 validity      // col1.a
    buffer 2: field 1 values        // col1.a
    buffer 3: field 2 validity      // col1.b
    buffer 4: field 2 offsets       // col1.b
        buffer 5: field 3 validity  // col1.b.item
        buffer 6: field 3 values    // col1.b.item
    buffer 7: field 4 validity      // col1.c
    buffer 8: field 4 values        // col1.c
buffer 9: field 5 validity          // col2
buffer 10: field 5 offsets          // col2
buffer 11: field 5 data             // col2
```

```
buffer 0: field 0 validity
buffer 1: field 1 validity
buffer 2: field 1 values
buffer 3: field 2 validity
buffer 4: field 2 offsets
buffer 5: field 3 validity
buffer 6: field 3 values
buffer 7: field 4 validity
buffer 8: field 4 values
buffer 9: field 5 validity
buffer 10: field 5 offsets
buffer 11: field 5 data
```

## 4. 字节顺序
Arrow默认小端存储。Scheme 元数据中由一个endianness字段来表明RecordBatch的字节顺序

## 5. IPC流格式
一条流数据报文，代表一组封装的信息。每条信息格式如下
```
<SCHEMA>
<DICTIONARY 0>
...
<DICTIONARY k - 1>
<RECORD BATCH 0>
...
<DICTIONARY x DELTA>
...
<DICTIONARY y DELTA>
...
<RECORD BATCH n - 1>
<EOS [optional]: 0xFFFFFFFF 0x00000000>
```

上述每一行对应节 "1. 数据封装"中的一条消息，不过其'metadata_flatbuffer'分别为'Schema','RecordBatch'或'DictionaryBatch'

## 6. IPC文件格式

```
<magic number "ARROW1">
<empty padding bytes [to 8 byte boundary]>
<STREAMING FORMAT with EOS>
<FOOTER>
<FOOTER SIZE: int32>
<magic number "ARROW1">
```
