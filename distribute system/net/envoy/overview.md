### 评价

在功能方面，[Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 旨在实现服务与边缘代理功能，通过管理微服务之间的交互以确保应用程序性能。该项目提供的超时、速率限制、断路、负载均衡、重试、统计、日志记录以及分布式追踪等高级功能，可以帮助用户以高容错性和高可靠性的方式处理各类网络故障问题。[Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 支持的协议与功能包括 HTTP/2、gRPC、MongoDB、Redis、Thrift、外部授权、全局速率限制以及一个富配置 API 等等。

与 [Istio](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#istio) 体系下的 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 相比，Linkerd 基于 Scala 为主的技术栈，作为数据平面代理并无优势。

## 名词介绍

![](D:/study/distribute%20system/net/envoy/overview/images/concepts-envoy-arch-simple.png)

1. 下游(Upstream): 数据请求的入口方向被称之为下游，在istio中又对应pod之外的客户端连接
2. 上游(Downstream): 数据请求的出口方向则称之为上游，在istio中对应pod之内被流量拦截的服务
3. 监听器(Listener): Envoy 使用监听器（Listener）来监听数据端口，接受下游连接和请求；
   1. istio中的sidecar proxy默认使用了端口为15090的端口作为prometheus遥测相关接口
   2. istio中的sidecar proxy默认使用了端口为15021端口作为istio-agent健康检查端口
   3. istio中的sidecar proxy默认使用了端口为15006，作为虚拟输出cluster，15001作为虚拟输入cluster
4. Cluster: 在上游方向，[Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 使用集群（[Cluster](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#cluster)）来抽象上游服务，管理连接池以及与之相关的健康检查等配置。
5. filter：filter是 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 中可拔插的多种功能组件的统称。
   1. 最为核心的 HTTP 代理功能就是构筑在一个名为“HTTP 连接管理器（Http Connection Manager）”的 L4 筛选器之上的。而 L7 筛选器（绝大部分情况下 L7 筛选器都可以和 HTTP 筛选器划等号）则是作为 L4 筛选器的子筛选器存在，用于支撑实现更加丰富的流量治理能力。



## 线程模型

[Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 采用多线程以及基于 [Libevent ](https://libevent.org/)(简单理解为基于文件描述符事件或超时或信号量)的事件触发机制来保证其超高的性能。一共存在三种不同的线程，分别是：Main 线程、Worker 线程以及文件刷新线程。

1. Main 线程负责配置更新（对接 xDS 服务）、监控指标刷新和输出、对外提供 Admin 端口、负责整个进程的管理等工作。
2. Worker 线程是一个非阻塞的事件循环，每个 Worker 线程都会监听所有的 Listener，并处理相关连接和请求事件。
3. 文件刷新线程负责将 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 需要持久化的数据写入磁盘。在 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 中，所有打开的文件（主要是日志文件）都分别对应一个独立的文件刷新线程用于周期性的把内存缓冲的数据写入到磁盘文件之中。而 Worker 线程在写文件时，实际只是将数据写入到内存缓冲区，最终由文件刷新线程落盘。如此可以避免 Worker 线程被磁盘 IO 所阻塞。

为了尽可能的减少线程间由于数据共享而引入的争用以及锁操作，[Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 设计了一套非常巧妙的 Thread Local Store 机制（简称 TLS）

![](https://www.servicemesher.com/istio-handbook/images/concept-envoy-thread.png)



## 扩展

筛选器本质上就是插件，因此通过扩展开发筛选器，可以在不侵入 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 主干源码的前提下，实现对 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 功能的扩展增强。而且 L3/L4 筛选器架构大大拓宽了 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 中“扩展”二字的可能性。

接收：

1. 连接建立：当操作系统接收到来自下游的连接时，会随机选择一个 Worker 来处理该事件。然后每一个监听器筛选器（Listener Filter）都会被用于处理该连接。直到所有的监听器筛选器执行完成，一个可操作的 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 连接对象才会被建立，[Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 开始接受来自下游的请求或数据。
2. 数据处理：当该连接具体的请求或者数据到来之时，各个 L4（Network）筛选器开始工作。L4 筛选器分为 Read 和 Write 两种不同类型，分别用于读取外部数据和向外部发送数据，它可以直接操作连接上的二进制字节流。在大部分的实现当中，L4 筛选器负责将连接中的二进制字节流解析为具有的协议语义的数据（如 HTTP Headers，Body 等）并交由 L7 筛选器进一步处理。
   1. 目前社区已经提供了与 HTTP、Dubbo、Mongo、Kafka、Thrift 等协议对应的多种 L4 筛选器。
   2. **实际上，在 L4 筛选器和 L7 筛选器之间，应该有一层专门的编解码器。解码器常用于对协议进行解析**
3. 路由处理：在所有的 L7 筛选器都执行完成之后，路由组件（Router）将会被调用，将请求通过连接池发送给后端服务，并异步等待后端响应。

响应执行流程为连接的倒序执行。

![](https://www.servicemesher.com/istio-handbook/images/concept-envoy-filter.png)



### xDS协议

与 HAProxy 以及 Nginx 等传统网络代理依赖静态配置文件来定义各种资源以及数据转发规则不同，[Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 几乎所有配置都可以通过订阅来动态获取，如监控指定路径下的文件、启动 **gRPC 流**或轮询 REST 接口，对应的发现服务以及各种各样的 API 统称为 xDS。

xDS 协议是由 [Envoy](https://envoyproxy.io/) 提出的，在 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) v2 版本 API 中最原始的 xDS 协议指的是 CDS（[Cluster](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#cluster) Discovery [Service](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#service)）、EDS（Endpoint Discovery [service](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#service)）、LDS（Listener Discovery [Service](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#service)） 和 RDS（Route Discovery [Service](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#service)），后来在 v3 版本中又发展出了 Scoped Route Discovery [Service](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#service)（SRDS）、Virtual Host Discovery [Service](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#service) （VHDS）、Secret Discovery [Service](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#service)（SDS）、Runtime Discovery [Service](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#service)（RTDS）详见 [xDS REST and gRPC protocol](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol)。

以 [Istio](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#istio) 中 [Pilot](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#pilot) 为例，当 [Pilot](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#pilot) 发现新的服务或路由规则被创建（通过监控 Kubernetes 集群中特定 [CRD](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#crd) 资源变化、或者发现 Consul 服务注册和配置变化），[Pilot](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#pilot) 会通过已经和 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 之间建立好的 GRPC 流将相关的配置推送到 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy)。针对不同类型的资源，[Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 提供了不同的 xDS API，包括 LDS、CDS、RDS等等。

1. LDS:  用于向 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 下发监听器的相关配置用于动态创建新的监听器或者更新已有监听器。
2. CDS: CDS 用于向 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 下发集群的相关配置用于创建新的集群或者更新已有的集群。其中包括健康检查配置、连接池配置等等。
3. EDS: CDS 服务负责集群类型的推送。而当该集群类型为 EDS 时，说明该集群的所有可访问的端点（Endpoints）也需要由通过 xDS 协议动态下发，而不使用 DNS 等手段解析。负责下发端点的服务就称之为 EDS。
4. RDS: RDS 用于下发动态的路由规则。路由中最关键的配置包含匹配规则和目标集群，此外，也可能包含重试、分流、限流等等。

筛选器作为核心的一种资源，但是并没有与之对应的专门的 xDS API 用于发现和动态下发筛选器的配置。筛选器的所有配置都是嵌入在 LDS、RDS、以及 CDS 当中，比如 LDS 下发的监听器和 CDS 下发的集群中会包含筛选器链的配置，而 RDS 推送的路由配置当中，也可能包含与具体路由相关的一些筛选器配置。

![](https://www.servicemesher.com/istio-handbook/images/concept-envoy-xds.png)



## 可观测性

日志（Access log），指标（Metrics），追踪（[Tracing](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#tracing)）三个模块从三个不同的维度来实现对所有流经 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 的请求的统计、观察和监测。

1. 日志是对 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 中事件（主要是指下游请求）的详细记录，用于定位一些疑难问题。
2. 指标是对 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 中事件的数值化统计，往往需要搭配 Prometheus 等事件数据库配合使用。允许筛选器自由的扩展属于自己的独特指标计数，如 HTTP 限流、鉴权等筛选器都扩展了对应的指标，使得 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 也可以从某个具体的流量治理功能的角度观察流量情况。
3. 追踪是对 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 以及上下游服务中多个事件因果关系的记录，必须要上下游服务同时支持，并对接外部追踪系统。[Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 原生支持了 Lightstep、Zipkin 等多种追踪系统，无需额外的修改或者开发，只需要简单的配置即可。

### 分布式追踪

### Envoy-Jaeger 架构

`Envoy` 原生支持 `Jaeger`，追踪所需 `x-b3` 开头的 Header 和 `x-request-id` 在不同的服务之间由业务逻辑进行传递，并由 `Envoy` 上报给 `Jaeger`，最终 `Jaeger` 生成完整的追踪信息。

在 `Istio` 中，`Envoy` 和 `Jaeger` 的关系如下：

![](https://www.servicemesher.com/istio-handbook/images/envoy-jaeger.png)

![](https://www.servicemesher.com/istio-handbook/images/jaeger-architecture.png)

图中 Front [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 指的是第一个接收到请求的 [Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) [Sidecar](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#sidecar)，它会负责创建 Root Span 并追加到请求 Header 内，请求到达不同的服务时，[Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) [Sidecar](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#sidecar) 会将追踪信息进行上报。在 `Istio` 提供“开箱即用”的追踪环境中，`Jaeger` 的部署方式是 `all-in-one` 的方式。该模式下部署的 [Pod](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#pod) 为 `istio-tracing`，使用 `jaegertracing/all-in-one` 镜像，包含：`Jaeger-agent`、`Jaeger-collector`、`Jaeger-query(UI)` 几个组件。

不同的是，`Bookinfo` 的业务代码并没有集成 `Jaeger-client` ，而是由 `Envoy` 将追踪信息直接上报到 `Jaeger-collector`，另外，存储方式默认为内存，随着 [Pod](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#pod) 销毁，追踪数据将会被删除。

## Envoy-Zipkin 架构

在 `Istio` 中，`Envoy` 原生支持分布式追踪系统 Zipkin。当第一个请求被 `Envoy` [Sidecar](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#sidecar) 拦截时，`Envoy` 会自动为 HTTP Headers 添加 `x-b3` 开头的 Headers 和 `x-request-id`，业务系统在调用下游服务时需要将这些 Headers 信息加入到请求头中，下游的 `Envoy` [Sidecar](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#sidecar) 收到请求后，会将 `Span` 上报给 `Zipkin` ，最终由 Zipkin 解析出完整的调用链。

详细的过程大致为：

- 如果请求来源没有 Trace 相关的 Headers，则会在流量进入 [POD](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#pod) 之前创建一个 Root Span。
- 如果请求来源包含 Trace 相关的 Headers，[Envoy](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#envoy) 将会解析 Span 上下文信息，然后在流量进入 [POD](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#pod) 之前创建一个新的 Span 继承自旧的 Span 上下文。

![](https://www.servicemesher.com/istio-handbook/images/zipkin-principle.png)



在 [Istio](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#istio) 服务网格内，调用链的 `Span` 信息由 `Envoy` 通过 proxy 直接上报给 `ZipkinServer`。服务和 `Zipkin` 的调用流程以及 Zipkin 内部结构大致如下：

![](https://www.servicemesher.com/istio-handbook/images/zipkin-architecture.png)

- 传输层：默认为 HTTP 的方式，当然还可以使用 Kafka、Scribe 的方式。
- 收集器（Collector）：收集发送过来的 Span 数据，并进行数据验证、存储以及创建必要的索引。
- 存储（Storage）：默认是 in-memory 的方式，仅用于测试。请注意此方式的追踪数据并不会被持久化。其他可选方式有 JDBC（Mysql）、Cassandra、Elasticsearch 。
- API：提供 API 外部调用查询，主要是 Web UI 调用。
- User Interfaces（Web UI）：提供追踪数据查询的 Web UI 界面展示。

## 待深入

1. Envoy Thread Local Store
2. filter开发与实现细节
3. xDS协议各种API [xDS REST and gRPC protocol](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol)。
4. https://www.servicemesher.com/istio-handbook/concepts/mosn.html
5. 几种分布式追踪的详细实现与优劣
6. evnoy对象详细属性与使用



## ref

1. [opentracing](https://opentracing-contrib.github.io/opentracing-specification-zh/)
2. [opentracing增强](https://www.servicemesher.com/istio-handbook/practice/enhance-tracing.html)
3. [service mesh](https://www.servicemesher.com/istio-handbook/concepts/overview.html)

