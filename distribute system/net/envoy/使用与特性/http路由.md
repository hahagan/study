# HTTP路由

Envoy 包括一个 HTTP [路由器过滤器](https://www.servicemesher.com/envoy/configuration/http_filters/router_filter.html#config-http-filters-router)，可以安装它来执行高级路由任务。这对于处理边缘流量（传统的反向代理请求处理）以及构建服务间的 Envoy 网格（通常是通过主机/授权 HTTP 头的路由到达特定的上游服务集群）非常有用。Envoy 也可以被配置为转发代理。在转发代理配置中，网格客户端可以通过适当地配置其 http 代理来成为 Envoy。路由在一个高层级接受一个传入的 HTTP 请求，将其与上游集群相匹配，在上游集群中获得一个[连接池](https://www.servicemesher.com/envoy/intro/arch_overview/connection_pooling.html#arch-overview-conn-pool)，并转发请求。路由过滤器支持以下特性：

- 将域/授权映射到一组路由规则的虚拟主机。
- 前缀和精确路径匹配规则（包括[大小写敏感](https://www.envoyproxy.io/docs/envoy/latest/api-v1/route_config/route#config-http-conn-man-route-table-route-case-sensitive)和大小写不敏感）。目前还不支持正则/slug 匹配，这主要是因为难以用编程方式确定路由规则是否相互冲突。因此我们并不建议在反向代理级别使用正则/slug 路由，但是我们将来可能会根据用户需求量增加对这个特性的支持。
- 在虚拟主机级别上的 [TLS 重定向](https://www.envoyproxy.io/docs/envoy/latest/api-v1/route_config/vhost#config-http-conn-man-route-table-vhost-require-ssl)。
- 在路由级别的[路径/主机](https://www.envoyproxy.io/docs/envoy/latest/api-v1/route_config/vhost#config-http-conn-man-route-table-vhost-require-ssl)重定向。
- 在路由级别的直接（非代理）HTTP 响应。
- [显式重写](https://www.envoyproxy.io/docs/envoy/latest/api-v1/route_config/route#config-http-conn-man-route-table-route-host-rewrite)。
- [自动主机重写](https://www.envoyproxy.io/docs/envoy/latest/api-v1/route_config/route#config-http-conn-man-route-table-route-auto-host-rewrite)，基于所选的上游主机的 DNS 名称。
- [前缀重写](https://www.envoyproxy.io/docs/envoy/latest/api-v1/route_config/route#config-http-conn-man-route-table-route-prefix-rewrite)。
- 在路由级别上的 [Websocket升级](https://www.envoyproxy.io/docs/envoy/latest/api-v1/route_config/route#config-http-conn-man-route-table-route-use-websocket)。
- 请求通过 HTTP 头或路由配置指定[请求重试](https://www.servicemesher.com/envoy/intro/arch_overview/http_routing.html#arch-overview-http-routing-retry)。
- 请求通过 [HTTP头](https://www.servicemesher.com/envoy/configuration/http_filters/router_filter.html#config-http-filters-router-headers) 或[路径配置](https://www.envoyproxy.io/docs/envoy/latest/api-v1/route_config/route#config-http-conn-man-route-table-route-timeout)指定超时。
- 通过[运行时](https://www.envoyproxy.io/docs/envoy/latest/api-v1/route_config/route#config-http-conn-man-route-table-route-runtime)从一个上游集群转移到另一个集群（参见[流量转移/分割](https://www.servicemesher.com/envoy/configuration/http_conn_man/traffic_splitting.html#config-http-conn-man-route-table-traffic-splitting)）。
- 使用基于[权重/百分比的路由](https://www.envoyproxy.io/docs/envoy/latest/api-v1/route_config/route#config-http-conn-man-route-table-route-weighted-clusters)的流量跨越多个上游集群（参见[流量转移/分割](https://www.servicemesher.com/envoy/configuration/http_conn_man/traffic_splitting.html#config-http-conn-man-route-table-traffic-splitting-split)）。
- 任意标题匹配[路由规则](https://www.envoyproxy.io/docs/envoy/latest/api-v1/route_config/route#config-http-conn-man-route-table-route-headers)。
- 虚拟集群规范。虚拟集群是在虚拟主机级别上指定的，由 Envoy 使用，在标准集群级别上生成额外的统计信息。虚拟集群可以使用正则表达式匹配。
- 基于[优先级](https://www.servicemesher.com/envoy/intro/arch_overview/http_routing.html#arch-overview-http-routing-priority)的路由。
- 基于[哈希](https://www.envoyproxy.io/docs/envoy/latest/api-v1/route_config/route#config-http-conn-man-route-table-hash-policy)策略的路由。
- [绝对 url ](https://www.envoyproxy.io/docs/envoy/latest/api-v1/network_filters/http_conn_man#config-http-conn-man-http1-settings)支持非 tls 转发代理。



