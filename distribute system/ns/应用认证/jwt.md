## 概念
[jwt](https://tools.ietf.org/html/rfc7519)用于url信息传输安全的包含认证信息的json结构体对象。jwt在请求中包含可以用于认证的信息。jwt可以使用对称加密或非对称加密与签名生成。
jwt主要由三个部分组成：header，payload和signature
signature可用于验证header和payload的完整性。signature往往使用密钥进行加密。

* header： header部分持有jwt的签名算法。
* payload：jwt的数据负载，一般认证需要的信息，例如用户或其他[信息字段](https://tools.ietf.org/html/rfc7519#section-4.1)

## 优点
在分布式系统中认证信息保存在jwt中，跨域或请求到非签发节点时，也能够通过jwt获取认证信息。将服务器变为无状态服务(仅针对认证系统的认证信息)。
