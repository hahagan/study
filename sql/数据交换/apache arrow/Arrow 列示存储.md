Arrow 列示存储.md

## 特性

* 语言无关
* 内存
* 元数据序列化
* 序列化与通用数据传输协议

列式存储特性
* 顺序存取
* O(1)的随机访问
* SMID支持
* 可重定位，从而允许零拷贝

## 单词统一

* Array/Vector： 拥有已知长度且类型相同的序列
* Slot： 某些特定数据类型的数组中的单个逻辑值
* buf/Contiguous region： 给定长度的顺序虚拟地址空间。可以通过单个指针偏移达到任何字节
* Pysical Layout： 数组的基础内存布局，例如32位整型数组与32位浮点型数组具有相同布局
* primitive type： 没有子类型的数据类型
* nested type： 数据类型完整结构取决与一个或多个其他子类型
* parent/child arrays： 用于解释nested类型中物理数组值间关系的名字
	* a List<T>-type parent array has a T-type array as its child
* logical type： 使用某些物理布局实现面向应用程序的语义值类型

## 物理内存结构
数组定义：
* logical type
* 连续的buffer
* 长64位的有符号整数的长度，length
* 长64位有符号整数的null值计数器，count
* 可选的dictionary，用于字典编码的数组
* 嵌套类型中额外的chilld arrays项

### 物理结构类型

* primitive(fixes-size): 拥有相同类型或位宽的序列
* Variable-size： 拥有可变字节长度的序列。
* Fixed list： 拥有相同数量子数组元素的嵌套类型
* Variable-size list: 拥有不同数量子数组元素的嵌套类型
* struct： 由不同拥有相同长度但类型可能不同的子字段组成的嵌套类型
* sparse / dense union： 表示值序列的嵌套布局，每个值都可以具有从子数组类型集合中选择的类型
* NUll： 空值

以上内存布局仅适用于数据，不适用于元数据

### 内存对齐与填充

内存对齐最优实践：
* 数字数组通过对齐访问进行检索
* 对齐有助于cache line
* 64位对齐方式有利于SIMD

###  数组长度
在Arrow元数据中数组长度以64位符号整型表示，同时实现支持32位最大符号整数。多语言环境下建议限制位32位空间以下个元素。使用多个数组块表示大数据集

### null计数器
元数据中以64位符号整型，表示物理序列中的空值数量，数据结构的一部分

### 位图(有效位图)
任何类型都可以由空值。具有空值的数组必须具有连续的内存缓冲区，称为空位图。空位图长度位64字节的倍数，确保每个数组中的元素在空位图中至少有一位对应

空位图中，对应数组位空则设为0，否则设为1
```
values = [0, 1, null, 2, null ,3] => 使用LSB编码，一组8个位中 LSBbitmap(0，0，1，0，1，0，11)
```
不具备null值的数组可以不分配空位图，为实现方便，始终选择分配一个空位图，但内存共享时应注意

嵌套类型具有自己的空位图和空计数，不管子数组的空值和空位

## primitive型结构
primitive值数组代表了一个拥有相同物理槽宽度数据组成的数组。
槽宽度一般使用字节衡量，特殊的使用bit-packed类型提供。

通常数组包含一个连续的内存缓冲区，其大小至少于其槽宽的整数倍相同.
相关位图连续分配，但不必于数据缓存相邻
```
[1, null, 2, 4, 8]


* Length: 5, Null count: 1
* Validity bitmap buffer:

  |Byte 0 (validity bitmap) | Bytes 1-63            |
  |-------------------------|-----------------------|
  | 00011101                | 0 (padding)           |

* Value Buffer:

  |Bytes 0-3   | Bytes 4-7   | Bytes 8-11  | Bytes 12-15 | Bytes 16-19 | Bytes 20-63 |
  |------------|-------------|-------------|-------------|-------------|-------------|
  | 1          | unspecified | 2           | 4           | 8           | unspecified |
```

### Variable-size Binary Layou
由offsset和data缓存区组成。
offset缓冲区含有length+1个符号整型组成，代表了每个在data缓冲区中的数据的起始偏移量。offset两个临值的差用于计算对应data的长度
offset缓冲区中最后一个槽位数组长度

### Variable-size List Layou
嵌套类型。由两个buffer，一个位图，一个人offset buffer和一个子数组组成。
布局类似，不过offset偏移量代表子数组的数据项索引
```
[[12, -7, 25], null, [0, -127, 127, 50], []]

* Length: 4, Null count: 1
* Validity bitmap buffer:

  | Byte 0 (validity bitmap) | Bytes 1-63            |
  |--------------------------|-----------------------|
  | 00001101                 | 0 (padding)           |

* Offsets buffer (int32)

  | Bytes 0-3  | Bytes 4-7   | Bytes 8-11  | Bytes 12-15 | Bytes 16-19 | Bytes 20-63 |
  |------------|-------------|-------------|-------------|-------------|-------------|
  | 0          | 3           | 3           | 7           | 7           | unspecified |

offset buffer指向了values array对应的数据

* Values array (Int8array):
  * Length: 7,  Null count: 0
  * Validity bitmap buffer: Not required
  * Values buffer (int8)

    | Bytes 0-6                    | Bytes 7-63  |
    |------------------------------|-------------|
    | 12, -7, 25, 0, -127, 127, 50 | unspecified |
```

```
[[12, -7, 25], null, [0, -127, 127, 50], []]

* Length: 4, Null count: 1
* Validity bitmap buffer:

  | Byte 0 (validity bitmap) | Bytes 1-63            |
  |--------------------------|-----------------------|
  | 00001101                 | 0 (padding)           |

* Offsets buffer (int32)

  | Bytes 0-3  | Bytes 4-7   | Bytes 8-11  | Bytes 12-15 | Bytes 16-19 | Bytes 20-63 |
  |------------|-------------|-------------|-------------|-------------|-------------|
  | 0          | 3           | 3           | 7           | 7           | unspecified |

* Values array (Int8array):
  * Length: 7,  Null count: 0
  * Validity bitmap buffer: Not required
  * Values buffer (int8)

    | Bytes 0-6                    | Bytes 7-63  |
    |------------------------------|-------------|
    | 12, -7, 25, 0, -127, 127, 50 | unspecified |
```