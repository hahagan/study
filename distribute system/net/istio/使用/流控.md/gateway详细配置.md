### 介绍
在网格边界的负载均衡器，负责接收入口 HTTP/TCP 连接。 其中配置了对外暴露的端口、协议等(配置逻辑上，实际不是)。

通过gateway配置，可以决定对外服务端口、端口服务协议、tls安全相关配置。



### Gateway资源对象

描述gateway通过selector选择IngressGateway实例进行挂载，IngressGateway根据gateway对象的server字段描述的服务端接收数据。

| field    | type               | description                | required |
| :------- | :----------------- | :------------------------- | -------- |
| servers  | Server list        | 服务端描述列表             | 是       |
| selector | Map<string,string> | 标签选择IngressGateway实例 | 是       |



#### Server字段

描述对外提供服务的服务入口

| field           | type                                                         | description                                                  | required |
| --------------- | ------------------------------------------------------------ | ------------------------------------------------------------ | -------- |
| port            | [port](####Port字段)                                         | 描述对外开发的端口行为                                       | 是       |
| host            | string[]                                                     | 对外开放的host列表。vs挂载到gateway时必须具有至少一个匹配的host | 是       |
| tls             | [TlsOptions](####Tlhttps://istio.io/latest/zh/docs/reference/config/networking/gateway/#Server-TLSOptions) | tls相关配置                                                  | 否       |
| defaultEndpoint | string                                                       | The loopback IP endpoint or Unix domain socket to which traffic should be forwarded to by default. | 否       |



#### Port字段

描述暴露端口

| field    | type   | description                                             | required |
| -------- | ------ | ------------------------------------------------------- | -------- |
| number   | uint32 | 开放端口                                                | yes      |
| protocol | string | 端口协议，支持HTTP\|HTTPS\|GRPC\|HTTP2\|MONGO\|TCP\|TLS | yes      |
| name     | string | 端口标签                                                | no       |





