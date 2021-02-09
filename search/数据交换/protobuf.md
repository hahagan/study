### 概念

google定义的一种独立于语言和平台的结构化数据序列化的拓展。它比XML更小更快更简单。在定义对应的结构体后，可以使用生成的代码简单的从数据流中读写结构化数据。

目前支持生成java、python、objective-c和c++。如果使用proto3版本，还支持Dart、gp、ruby和C#。



### 编码

#### 信息定义结构体

一个message定义如下

```
syntax = "proto3";

message SearchRequest {
  string query = 1;
  int32 page_number = 2;
  int32 result_per_page = 3;
}
```

每个message由多个字段定义组成。一个字段定义由字段类型、字段名、字段ID以及字段规则四个部分组成。

#### 字段规则

Message fields can be one of the following:

- singular: a well-formed message can have zero or one of this field (but not more than one). And this is the default field rule for proto3 syntax.
- `repeated`: this field can be repeated any number of times (including zero) in a well-formed message. The order of the repeated values will be preserved.

#### 字段类型表

| Type | Meaning          | Used For                                                 |
| :--- | :--------------- | :------------------------------------------------------- |
| 0    | Varint           | int32, int64, uint32, uint64, sint32, sint64, bool, enum |
| 1    | 64-bit           | fixed64, sfixed64, double                                |
| 2    | Length-delimited | string, bytes, embedded messages, packed repeated fields |
| 3    | Start group      | groups (deprecated)                                      |
| 4    | End group        | groups (deprecated)                                      |
| 5    | 32-bit           | fixed32, sfixed32, float                                 |

#### 序列化编码

每一个信息流中的字段都是`(字段ID << 3) | 字段类型`的varint表示。即最后三个位表示字段类型，前n个有效位表示字段在结构体中的顺序。

```protobuf
message Test1 {
  optional int32 a = 1;	// 类型为 int32, 字段ID为1，字段规则为optional
}
```

例如Test1结构体实例序列化后的数据表示为`08 96 01`。第一个字节中的`08`为`0000 1000`，根据字段定义，可以确定，第一个字段在类型为0，即`varint`类型。即表示`96 01`代表了一个varint类型的数据，通过解码得到其值为150。

序列化后的数据与字段名无关，因此对数据进行反序列化时只要类型和结构相同，即使字段名不同也可以正常反序列化。