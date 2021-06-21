# REF

* https://tools.ietf.org/html/rfc5246



# 1. 介绍

主要目的是保证应用间隐私和数据安全。协议由两层构成，`tls Record protocol`和`tls Handshake`组成。

`tls Record protocol`分层基于更底层的可靠网络传输层(tcp)。`tls Record protocol`基于两个特性保证连接安全：

1. 连接私有。使用对称加密对数据进行加密，每个连接的密钥独一无二，并由其他协议谈判形成(`tls Handshake protocol`)
2. 连接可靠。通过使用密钥MAC校验数据。使用安全hash(sha-1)计算mac。tls也可以在无mac的前提下运行，但一般仅在谈判中的安全参数进行传输时使用

`tls Record protocol`会被更上层协议封装，`tls handshake protocol`会封装`tls record protocol`用于在服务端和客户端间协商加密算法和密钥。`tls handshake`通过三个特性保证连接安全：

1. 对端身份通过非对称方式、公钥、密码学等手段加密。该认证是可选的，但一般来说最少对一端进行认证
2. 协商得出的密钥必须是安全的。窃听者获取无效，中间伪装者无法获取
3. 协商过程可靠。网络攻击者不可更改协商通信



## 2. 目标

1. 加密算法安全: tls应用于在两方间建立安全的连接
2. 互操作性: 可在未知对方代码情况下，利用tls成功的交换加密算法参数，并进行开发
3. 扩展性: 提供一套框架在有必要时可以增加公钥和系列加密算法，可分解为两个子目标，确保需要时可以提供新协议，避免实现一个全新的安全库
   1. 性能：加密操作往往为CPU密集型，尤其是公钥操作。因此tls协议包含了一个可选的会话缓存方案，以减少需要从头开始建立的连接数量。



## 4. 数据介绍

### 4.1 基本块大小

所有的数据项都是显式表示的。每个基本的数据块为一个字节。多字节数据自左向右将字节进行串联

### 4.3 Vectors

一个数组为了一组**同构**的数据元素。数组的大小可能会在文档进行定义时决定，也可能在运行时都尚未决定。为了兼顾两种情况，vector的长度指其字节数，而不是元素个数。假设类型`T`的数组`T T'[n]`，这里表示`T'`在数据流中占用了n个字节，在编码的字节流中不包含vector的元素数量。

这里一个例子，`Datum`为三个连续的字节，而`Data`为三个连续的`Datum`，总共9字节

```	c
      opaque Datum[3];      /* three uninterpreted bytes */
      Datum Data[9];        /* 3 consecutive 3 byte vectors */
```

可变长数组通过`T T'<floor..ceiling>;`表示。以下例子中`mandatory `是一个包含了300到400个字节的数组并且不可为空。实际为长度占用为2个字节的uint16类型。在`longer`数组中可以表示400个unint16。在实际编码中会包含两个字节的数组长度。

```c
opaque mandatory<300..400>;            /* length field is 2 bytes, cannot be empty */      
uint16 longer<0..800>;            /* zero to 400 16-bit unsigned integers */
```

## 4.4 数字

最基本的数字类型为一字节的uint8。其他更长的数字通过定长的基本块数据组成多字节数据，且同为非符号类型数字。多字节数字使用**网络字节序(大端存储)**例如

```c
      uint8 uint16[2];
      uint8 uint24[3];
      uint8 uint32[4];
      uint8 uint64[8];
```

### 4.5 枚举

枚举仅能假设该类型中的值为目标值，一个枚举的大小为枚举内最大类型的大小

.....



## 5. HMAC与伪随机函数





## 6. tls record protocol

每条信息都会包含长度、描述、内容等字段。`tls record`会将message碎片化为可管理的block，并对其传输。`tls record protocol`还可以进行额外的压缩、加密、MAC等。对端接收到数据后则对其进行解压、解密重组等行为，最后再提交给上层应用。

### 6.1 connection states

连接状态表示的是`tls record protocol`的操作环境。主要包含压缩算法、加密算法、MAC算法。逻辑上连接包含4钟状态，当前读/写状态，悬挂读/写状态。所有`tls record protocol`都在当前读/写状态下进行，悬挂读/写状态由`tls handshake protocol`设置参数，并通过`ChangeCipherSpec`可选择的将两者进行转换或重新初始化。

读写状态的安全参数有：

* connection end：表示客户端或服务端
* PRF algorithm：从secrect中生成密钥的算法
* bulk encryption algorithm：用于块加密的算法
* MAC algorithm：message校验算法
* compress algorithm：数据压缩算法
* master secrete：两端共享的48字节密钥
* client random：客户端提供的32字节
* server random：服务端提供的32字节

```
 struct {
          ConnectionEnd          entity;
          PRFAlgorithm           prf_algorithm;
          BulkCipherAlgorithm    bulk_cipher_algorithm;
          CipherType             cipher_type;
          uint8                  enc_key_length;
          uint8                  block_length;
          uint8                  fixed_iv_length;
          uint8                  record_iv_length;
          MACAlgorithm           mac_algorithm;
          uint8                  mac_length;
          uint8                  mac_key_length;
          CompressionMethod      compression_algorithm;
          opaque                 master_secret[48];
          opaque                 client_random[32];
          opaque                 server_random[32];
      } SecurityParameters;
```

tls record 将会使用安全参数创建如下对象

```
      client write MAC key
      server write MAC key
      client write encryption key
      server write encryption key
      client write IV
      server write IV
```

每个当前状态包含以下状态，这些状态应该在安全参数被设置后，对应状态被实例化时，为每个已处理的record进行更新:

```
compression state
      The current state of the compression algorithm.
      
cipher state
      The current state of the encryption algorithm.  This will consist
      of the scheduled key for that connection.  For stream ciphers,
      this will also contain whatever state information is necessary to
      allow the stream to continue to encrypt or decrypt data.
      
MAC key
      The MAC key for this connection, as generated above.
    
sequence number
      Each connection state contains a sequence number, which is
      maintained separately for read and write states.  The sequence
      number MUST be set to zero whenever a connection state is made the
      active state.  Sequence numbers are of type uint64 and may not
      exceed 2^64-1.  Sequence numbers do not wrap.  If a TLS
      implementation would need to wrap a sequence number, it must
      renegotiate instead.  A sequence number is incremented after each
      record: specifically, the first record transmitted under a
      particular connection state MUST use sequence number 0.
```



### 6.2 record layer

tls record layer 从上层接收任意大小非空的未解释数据块

#### 6.2.1 Fragmentation

record layer将数据块拆分为TLSPlaintext,大小为16k或更小的组。客户端的信息边界不会在record layer保存。多个客户端的相同ContentType 的信息可能会被合并为一个TLSPlaintext record，或一条信息被切分为多个recrod

```
      struct {
          uint8 major;
          uint8 minor;
      } ProtocolVersion;

      enum {
          change_cipher_spec(20), alert(21), handshake(22),
          application_data(23), (255)
      } ContentType;

      struct {
          ContentType type;
          ProtocolVersion version;
          uint16 length;
          opaque fragment[TLSPlaintext.length];
      } TLSPlaintext;
```

#### 6.2.2 Record Compression and Decompression

根据当前会话的状态，record将会被tls压缩为TlsCompressed。所有的状态中都必须含有压缩算法，初始默认的压缩算法为CompressionMethod.null。并且整个压缩应该是无损的，增加的长度不应超过1024个字节，对端解压后大小不应超过16K字节，否则应返回fatal decompression failure error

```
      struct {
          ContentType type;       /* same as TLSPlaintext.type */
          ProtocolVersion version;/* same as TLSPlaintext.version */
          uint16 length;
          opaque fragment[TLSCompressed.length];
      } TLSCompressed;
```

#### 6.2.3 Record Payload Protection

加密和MAC函数将TLSCompressed转换为TLSCiphertext。同时MAC的record还会记录序列号，因此可以实现丢包、额外、重复的信息(接收一般的情况怎么办)

```
      struct {
          ContentType type;
          ProtocolVersion version;
          uint16 length;
          select (SecurityParameters.cipher_type) {
              case stream: GenericStreamCipher;
              case block:  GenericBlockCipher;
              case aead:   GenericAEADCipher;
          } fragment;
      } TLSCiphertext;
```

#### 6.2.3.1 Null or Standard Stream Cipher

stream cipher 将`TLSCompressed.fragment`转换为`TLSCiphertext.fragment`

```c
      stream-ciphered struct {
          opaque content[TLSCompressed.length];
          opaque MAC[SecurityParameters.mac_length];
      } GenericStreamCipher;
```

```
    MAC = MAC(MAC_write_key, seq_num +
                            TLSCompressed.type +
                            TLSCompressed.version +
                            TLSCompressed.length +
                            TLSCompressed.fragment);
                            
    seq_num
      The sequence number for this record.

   	MAC
      The MAC algorithm specified by SecurityParameters.mac_algorithm.
```

#### 6.2.3.2  CBC Block Cipher

