# REF

[quic传输译文](https://autumnquiche.github.io/RFC9000_Chinese_Simplified/#RFC9000_QUIC)

[quic-TLS译文](https://autumnquiche.github.io/RFC9001_Chinese_Simplified/)

[tls1.3](https://www.rfc-editor.org/info/rfc8446)

# 阅读疑问与原理

## 为什么初次建立连接时TCP-TLS消耗较高而QUIC-TLS是1-RTT

TLS-tcp，首先为tcp连接建立握手，随后进行TLS握手流程，ClientHello/ServerHello(触发TLS和server证书获取)和clientTLSFinish/ServerTLSFinish(随机数3交互和客户端证书)。最后进行数据传输

quic-tls直接对接quic协议，quic协议允许将多种帧放在同一个包中发送，将TLS的底层连接建立与ClientHello/ServerHello合并，此时只使用了1-RTT。而TLS的后续握手阶段与数据传输阶段合并，即在下一次数据交互时完成TLS后续握手流程并可以直接传输数据。所以quic声称建立连接仅需要1-RTT，连接建立消耗的RTT比经典的TCP-TLS更低。

## quic这样合并连接建立RTT，包大小和分包怎么处理

[加密信息缓存](https://autumnquiche.github.io/RFC9000_Chinese_Simplified/#7.5_Cryptographic_Message_Buffering)

[分包](https://autumnquiche.github.io/RFC9000_Chinese_Simplified/#13_Packetization_and_Reliability)

缓存加密帧

