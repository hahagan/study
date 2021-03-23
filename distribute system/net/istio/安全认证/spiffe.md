[toc]



# 简介

**SPIFFE**(the Secure Production Identity Framework for Everyone),一种为动态或异构的软件安全认证开源标准。采用SPIFFE的系统可以简单可靠的双向认证。

在网络策略中常规的流量管理策略使用对外IP进行管理，而在异构或动态的环境下难以继续采用，例如k8s的PodIP是动态的，网络访问可以被NAT改写源端等行为。而SPIFFE是一种能够跨异构环境和组织边界引导和向服务发布标识的框架。



# 基本概念

## Workload

一个`Workload`代表一种工作负载，可以由多个执行相同任务的软件实例组成。在不同的软件系统中有不同的定义，例如：

* 在集群上通过负载均衡负载的一组web应用实例
* 一个mysql实例
* 在队列里中处理对象的处理程序
* 一系列独立部署但共同协作的程序集合。例如使用数据库的web服务。当然也可以将其各个独立组件作为`workload`

使用`workload`可以比使用虚拟主机或物理主机更细粒度的身份划分,但需要注意的不同`workload`间的隔离性,以确保不会被恶意`workload`偷取其他`workload`的身份信息,而SPIFFE不会保证这种隔离性。



## SPIFFE ID

`SPIFFE ID`是一个唯一标识一个`workload`身份的字符串。一个`SPIFFE ID`为[uri](https://tools.ietf.org/html/rfc3986)格式遵循`spiffe://trust-domain/workload-indetifer`格式，并且最大长度不超过2048字节

`trust-domain`非空，且不存在全局统一控制，如果出现`trust-domain`碰撞，不同`trust-domain`间会通信失败(同一`trust-domain`内仍能运行)，因为碰撞的`trust-domain`相同，但证书不同，认证失败。因此可以考虑使用前缀标识或使用随机ID以减少碰撞

`workload-indetifer`通常用于指向`workload`的唯一标识，结构可如同文件目录般分级。但不可以'/'结尾。该路径可以由使用者根据自行策略进行设置。



## Trust Domain

`trust domain`对应系统的root。一个`trust domain`可以代表个人、组织、区域、或独立运行的SPIFFE基础设施的部门。在同一`trust domain`中所有`workload`已颁发的身份文档(svid)，可以根据`trust domain`的root密钥进行验证。

一般建议为不同的位置、安全策略环境配置不同的`trust doamin`。例如不同的云环境或数据中心使用不同的`trust domain`



## SPIFFE Vetifinable Indentity  Document(SVID)

SVID是`workload`用来向资源或调用方证明其身份的文档。目前SVID包含一个SPIFFE ID，并支持通过X.509证书或JWT  TOKEN的方式进行编码。由于token容易被用于执行重放攻击，因此建议使用证书类型的SVID。

当一个svid被SPIFFE ID的trust domain签名后，可以认为该svid是有效的。

### [X.509-svid](https://github.com/spiffe/spiffe/blob/master/standards/X509-SVID.md)

#### 格式

* SPIFFE ID
  * 以uri类型保存在Subject Alternative Name extension (SAN extension, see [RFC 5280 section 4.2.16](https://tools.ietf.org/html/rfc5280#section-4.2.1.6))。svid有且只有一个san的uri字段，但可以san可以包含其他字段。
* 私钥，用于对workload行为进行数据签名。以及一个对应的短期x.509证书会被创建。这个证书可以用于建立tls连接或其他验证其他workload
* 一组有trust bundle指定的证书，可以用于认证其他workload的X509-svid的证书



#### 证书关系

证书分为根证书、叶子证书以及中间证书。

叶子证书用于认证调用或资源，使用于程序的认证。叶子证书的SPIFFE ID必须是非根路径格式。

其他证书又被称为签名证书，签名证书会在`key usage extension`上设置`keyCertSign`，在`basic constraits extension`中将`ca`标志设为真值，用于签发其他证书和验证签发证书，不用于认证。一个签名证书自身作为一个SVID，其SPIFFE ID不存在资源路径，并可以驻留在其所签发的叶子SVID中。



#### 验证

svid验证在遵循X509证书路径验证的基础上额外不加了一些SPIFFE指定的验证行为。而对svid验证过程中使用的证书被称为CA bundle。CA bundle的查找由SPIFFE workload api决定。

##### 叶子节点验证

首先需要进行X509标准验证，确保证书为叶子证书，签名机构有权限进行签发。

其次，证书验证时需要确保`basic constraints extensio`的 `CA`字段必须为false，并且`key usage extension`的`keyCertSign` and `cRLSign`字段没有设置。`SAN`的uri为SPIFFE ID格式。



#### SPIFFE Bundle的表示

##### 生成

SVID的`trust domain`d的ca证书以JWK的形式保存在SPIFFE bundle中，每个jwk代表一个CA证书。每个JWK证书的`use`参数必须设为`x509-svid`，`kid`参数不能设置，` x5c`参数必须等于其代表整数的CA证书的base64 DER(Distinguished Encoding Rules)编码。

##### 使用

首先识别一个SPIFFE Bundle内所含的jwk类型，对`use`参数为`x509-svid`的jwk进行处理，否则忽略(表示改信任域不支持X509-SVID)，对`x5c`参数进行解码处理，如果值为空则被忽略。x509-sivd的ca bundle为从x509-svid中提取的ca证书的合集。当`x5c`为多个值时除第一个值外全都需要进行提取



#### x509字段附录

| Extension              | Field                     | Description                                                  |
| ---------------------- | ------------------------- | ------------------------------------------------------------ |
| Subject Alternate Name | uniformResourceIdentifier | This field is set equal to the SPIFFE ID. Only one instance of this field is permitted. |
| Basic Constraints      | CA                        | This field must be set to `true` if and only if the SVID is a signing certificate. |
| Basic Constraints      | pathLenConstraint         | This field may be be set if the implementor wishes to enforce a finite CA hierarchy depth, for example, if inherited from an existing private key infrastructure (PKI). |
| Name Constraints       | permittedSubtrees         | This field may be set if the implementor wishes to use URI name constraints. It will be required in a future version of this document. |
| Key Usage              | keyCertSign               | This field must be set if and only if the SVID is a signing certificate. |
| Key Usage              | cRLSign                   | This field may be set if and only if the SVID is a signing certificate. |
| Key Usage              | keyAgreement              | This field may be set if and only if the SVID is a leaf certificate. |
| Key Usage              | keyEncipherment           | This field may be set if and only if the SVID is a leaf certificate. |
| Key Usage              | digitalSignature          | This field must be set if and only if the SVID is a leaf certificate. |
| Extended Key Usage     | id-kp-serverAuth          | This field may be set for either leaf or signing certificates. |
| Extended Key Usage     | id-kp-clientAuth          | This field may be set for either leaf or signing certificates. |



### [jwt-svid](https://github.com/spiffe/spiffe/blob/master/standards/X509-SVID.md)

#### **格式**

* SPIFFE ID
* JWT TOKEN
* 一组有trust bundle指定的证书，可以用于认证其他workload的jwt-svid的证书

jwt-svid时一种标准jwt token附加了额外的限制。这是由于jose自身存在的安全问题，jwt-svid在此基础上额外限制以避免这些安全问题。

#### 额外限制

* header：除了[官网](https://github.com/spiffe/spiffe/blob/master/standards/JWT-SVID.md#2-jose-header)说明的header和其值外，jwt-svid jose head中不可以添加其他的head或值
* jwt claim：除下方说明外，没有特殊说明，但对其进行额外设置时需要考虑是否会影响双方的互操作性
  * subject：必须设置为SPIFFE ID
  * audience：必须存在一个或多个值，验证者必须拒绝不存在该值或该值不在验证者持有范围内的请求。
  * exp：必须设置，为空时拒绝

jwt-svid是通过JWS Compact Serialization的jws( JSON Web Signature )结构体，因此不能使用JWS JSON Serialization。

#### token签名与验证

jwt-svid签名和验证于常规的JWT/JWS相同，但验证者必须确保`alg`头部必须与官方限制相同。jwt的生成过程需要遵循[RFC 7519 section 7](https://tools.ietf.org/html/rfc7519#section-7)。

#### JWT的传输

* 必须使用[RFC 7515 Section 3.1](https://tools.ietf.org/html/rfc7515#section-3.1)说明的`Compact Serialization`序列化方式，不能使用jws json序列化。
* 通过http传输，需要将值附加于`Authorization`头部(http2使用`authorization`头部)，并使用"Bearer"作为前缀
* 使用grpc协议需要遵循http2的header设置

#### SPIFFE Bundle的表示

##### 生成

与x509-svid类似，一个jwk作为一个签名密钥。jwk的`use`参数必须设置为`jwt-svid`。每个`kid`参数必须设置

##### 使用

首先获取`use`参数识别jwk类型，值为`jwt-svid`的进行处理，否则忽略。随后根据[JWT RFC 7519](https://github.com/spiffe/spiffe/blob/master/standards/JWT-SVID.md)描述进行验证，但需要注意上文描述的限制。

#### 安全风险

* 重放攻击，使用`aud`和`exp`可以减少但不能完全解决
* audience 数量，如果使用多个audience claim，则存在token被劫持与滥用的可能
* 传输安全，token拦截使得拦截者可能获得jwt所赋予权限的风险



#### ref

* [JWT RFC 7519](https://github.com/spiffe/spiffe/blob/master/standards/JWT-SVID.md)
* [jwt-svid](https://github.com/spiffe/spiffe/blob/master/standards/JWT-SVID.md)



## workload api

SPIFFE Workload API是一种方法，通过该方法，工作负载或计算流程可以获取其SVID。它提供使工作负载或计算过程能够利用SPIFFE身份和基于SPIFFE的身份验证系统的信息和服务。

在[AWS EC2 Instance Metadata API](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html)和[Google GCE Instance Metadata API](https://cloud.google.com/compute/docs/storing-retrieving-metadata)上不需要workload了解任何认证相关的信息，应用不需要做任何的认证相关的处理。与这些API不同，Workload API与平台无关，并且可以在进程级别和内核级别识别正在运行的服务。这意味着它适用于容器调度服务。

为了最小程度的密钥暴露风险，所有的密钥和对应证书都是短期的，并会被频繁且自动的轮转。workload可以在密钥过期前通过workload api请求新的证书和trust bundles。



###  X509-SVID profile

SPIFFE Workload api服务端实现grpc流模式进行响应，以便更快速的传播更新信息(例如ca证书)。需要注意服务端的响应信息需要持有所有信息而不仅仅是改变信息，并且由对应改变响应的事件决定其时间。客户端收到响应请求时则认为该响应产生了某个事件。并且客户端应注意推送响应的时机和频率。

`FetchX509SVID` rpc返回SVID的`trust domain`的`trust bundle`，也可以提供外部`trust domain`的`trust bundle`

 `FetchX509Bundles` RPC返回 SPIFFE ID的`trust domain`的`trust bundle`。由于该请求不会返回svid，因此补零时在`trust domain`内外，响应所有中的bundle会以相同的方式编码。

处理客户端连接利用其`trust domain`声明，选择对应的`trust domain`和bundles，达到对内外`trust domain`服务的认证的能力。

X509-SVID profile支持配置多个身份，但同时只能使用其中的一个，如果应用不需要知道自身身份，则使用定义的默认身份。这是因为大部分情况下，应用并不需要了解自身身份，而是由管理者进行权限授予，但仍有需要从多身份中选择特定身份完成特定行为的场景。

如果服务端对一个workload 的所有svid或客户端的所有`trust bundles`进行了修改，则应该项客户端返回终止grpc流响应，客户端接收到响应后重新创建对应grpc流

#### 服务定义

Workload api可以通过proto buf配置文件定义服务。因此可以通过阅读Workload api的配置文件以获得与每种方法相关的消息定义

```protobuf
syntax = "proto3";

message X509SVIDRequest {  }

service SpiffeWorkloadAPI {
    // X.509-SVID Profile
    // Fetch all SPIFFE identities the workload is entitled to, as
    // well as related information like trust bundles and CRLs. As
    // this information changes, subsequent messages will be sent.
    rpc FetchX509SVID(X509SVIDRequest) returns (stream X509SVIDResponse);

    // Fetch trust bundles and CRLs.  Useful for clients that only
    // need to validate SVIDs without obtaining an SVID for themself.
    // As this information changes, subsequent messages will be sent.
    rpc FetchX509Bundles(X509BundlesRequest) returns (stream X509BundlesResponse);
}
```

#### rpc消息定义

`X509SVID`消息中的所有字段都是必须且无默认值的。`X509SVIDResponse`消息这个i不过`svid`为必须的。

```protobuf
// The X509SVIDResponse message carries a set of X.509 SVIDs and their
// associated information. It also carries a set of global CRLs and a
// map of federated bundles the workload should trust.
message X509SVIDResponse {
    // A list of X509SVID messages, each of which includes a single
    // SPIFFE Verifiable Identity Document, along with its private key
    // and bundle.
    repeated X509SVID svids = 1;

    // ASN.1 DER encoded certificate revocation list.
    repeated bytes crl = 2;

    // CA certificate bundles belonging to foreign Trust Domains that the
    // workload should trust, keyed by the SPIFFE ID of the foreign
    // domain. Bundles are ASN.1 DER encoded.
    map<string, bytes> federated_bundles = 3;
}

// The X509SVID message carries a single SVID and all associated
// information, including CA bundles.
message X509SVID {
    // The SPIFFE ID of the SVID in this entry. MUST match the SPIFFE ID
    // encoded in the `x509_svid` certificate.
    string spiffe_id = 1;

    // ASN.1 DER encoded certificate chain. MAY include intermediates,
    // the leaf certificate (or SVID itself) MUST come first.
    bytes x509_svid = 2;

    // ASN.1 DER encoded PKCS#8 private key. MUST be unencrypted.
    bytes x509_svid_key = 3;

    // CA certificates belonging to the Trust Domain
    // ASN.1 DER encoded
    bytes bundle = 4;

}

// The X509BundlesResponse message carries a set of global CRLs and a
// map of trust bundles the workload should trust.
message X509BundlesResponse {
    // ASN.1 DER encoded certificate revocation list.
    repeated bytes crl = 1;

    // CA certificate bundles belonging to Trust Domains that the
    // workload should trust, keyed by the SPIFFE ID of the trust
    // domain. Bundles are ASN.1 DER encoded.
    map<string, bytes> bundles = 2;
}
```

### 例子

#### 服务端

![](E:\study\distribute system\ns\istio\images\workload_api_server_diagram.png)

1. `SPIFFE Workload Endpoint`开始监听
   * `SPIFFE Workload Endpoint`服务于对初始身份的初始化，对根信任的分发和管理
   * 但从整个过程中发现，SPIFFE Workload endpoint的缺少对应的保护机制，因此需要额外的机制进行保护，例如该服务仅对特定端点开放，unix开放，基础网络设施基于IP的身份强验证等
2. 提供 grpc服务
3. 接收到`FetchX509SVIDRequest `，检查连接，确保其有可用身份
4. 服务端向客户端响应`FetchX509SVIDResponse`
5. 进入连接等待状态，在中断连接或取消时进入关闭状态
6. 在等待状态获知需要更新时，对客户端进行验证确保其仍连接且有对应的身份。
7. 关闭流，返回相应错误码
8. 在监听器无法创建或grpc服务端遇到严重错误时退出

#### 客户端

![](E:\study\distribute system\ns\istio\images\workload_api_client_diagram.png)

1. 连接服务端
2. 调用`FetchX509SVID `请求服务端
3. 客户端等待`X509SVIDResponse`响应
4. 更新自身SVID和CRLs，bundles
5. 遇到异常，退出
6. 根据返回错误码类型，选择重试或退出



## Trust Bundle

当使用X.509-SVID时，`trust-bunndle`被用于目标`workload`识别源端`workload`。`trust bundle`是一个或多个可信赖ca的root证书的集合。

`trust bundles`包含X.509和JWT类型SVID的公钥。公钥在认证X.509 SVID时是一系列证书。公钥在认证JWT时是一个原始公钥。并且`trust bundle`会频繁轮转，workload会在调用workload api时获取到新的`trust bundle`



### 格式

bundles以[RFC 7517](https://tools.ietf.org/html/rfc7517) 的jwk集合形式表示。使用jwk的原因有两个，一个是该格式在表达多种密码学加密方式较灵活，便于后续扩展或调整。第二个是，jwk有较广泛的支持和应用，有利于多域级联。

#### jwk set

SPIFFEJWK的基础上做了额外的限制。

`spiffe_sequence`必须设置，单调自增，bundle改变时必须更新，可用于进行传播检测，更新的顺序性和证书代谢能力。并且未定义其上限

`spiffe_refresh_hint`必须设置，用于指示使用者多久之后应该回检以更行配置，单位为秒。该值为建议值，实际情况由使用者决定，可以进一步提高或减少频率。

`key`必须存在，值为一个`JWKs`数组。该数组可以为空数组，为空时表示`trust domain`已吊销任何先前发布的密钥。

#### jwk

同一个jwk元素代表一个密钥，用于对同一类型svid的认证。SPIFFE做了一些额外的限制，实现者不能额外定义除了以下说明以外的参数和值

`kty`必须设置且遵循[RFC 7517 Section 4.1](https://tools.ietf.org/html/rfc7517#section-4.1) ，代表密钥的加密算法。

`use`必须设置，用于表示SVID的类型，目前仅支持两个值`x509-svid`和`jwt-svid`



### 跨域

在跨`trust doamin`通信时需要获取外域的bundle，因此需要一种bundle的传输机制。SPIFFE的主要bundle传输机制类似于[OpenID Connect](https://openid.net/connect/)中的`jwks_uri`,又被称为`bundle endpoint`。

`bundle endpoint`的服务端和客户端必须支持TLS保护的HTTP传输用于保证SPIFFE各种实现的互操作性以及和OpenId连接的兼容性。

#### url稳定性

通过利用`spiffe_sequence`和`spiffe_refresh_hint`可以确定对应`trust bundle`的utl，客户端通过sequence可以确定新的bundle，通过hint可以决定多久之后poll进行更新。即使bundle更新也需要保持url不变，以确保客户端可以获得对应的版本。

#### bundle endpoint的安全性

bundle endpoint的传输非常重要，如果传输是不安全的，那么攻击者将可以伪装为任意身份。SPIFFE对此提出了两种建议的机制。

第一种是web pki，利用具有公信力的外部基础设施完成传输安全。例如从绑定dns或IP地址的公共CA种获取证书，通过操作者配置SPIFFE控制面，导入外部bundle。

第二种是利用SPIFFE 认证，通过X509-SVID进行保护(跨域怎么办？)。在`trust domain`内通过自身提供的身份完成传输。导入外部bundle，则又管理员对SPIFFE控制面完成。在导入外部包后，可以通过连接通道自行的完成证书的轮转。

### ref

https://github.com/spiffe/spiffe/blob/master/standards/SPIFFE_Trust_Domain_and_Bundle.md



#### 



