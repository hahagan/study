![整体](https://istio.io/latest/zh/docs/ops/deployment/architecture/arch.svg)

Istio 服务网格从逻辑上分为数据平面和控制平面。

- **数据平面** 由一组智能代理（[Envoy](https://www.envoyproxy.io/)）组成，被部署为 sidecar。这些代理负责协调和控制微服务之间的所有网络通信。他们还收集和报告所有网格流量的遥测数据。
- **控制平面** 管理并配置代理来进行流量路由



### Envoy

Envoy 是用 C++ 开发的高性能代理，用于协调服务网格中所有服务的入站和出站流量。Envoy 代理是唯一与数据平面流量交互的 Istio 组件。Envoy 代理被部署为服务的 sidecar.

问：envoy提供了什么能力，怎么提供的，通过什么方式配置，获取配置，修改配置。

因为Envoy为sidecar，那么Envoy能力仅限于自身网格，那么在整体需要一个整体的控制。



### Pilot

Pilot 为 Envoy sidecar 提供服务发现、用于智能路由的流量管理功能（例如，A/B 测试、金丝雀发布等）以及弹性功能（超时、重试、熔断器等）。

Pilot 将控制流量行为的高级路由规则转换为特定于环境的配置，并在运行时将它们传播到 sidecar。Pilot 将特定于平台的服务发现机制抽象出来，并将它们合成为任何符合 [Envoy API](https://www.envoyproxy.io/docs/envoy/latest/api/api) 的 sidecar 都可以使用的标准格式。

即服务发现规则抽象平台，并Envoy api分发到各个Envoy。



### Citadel

身份证书管理服务。

使用 Citadel，operator 可以执行基于服务身份的策略，而不是相对不稳定的 3 层或 4 层网络标识。



### Galley

底层平台与Istio的组件的中间层，用于对底层平台的封装抽象，向上提供统一的配置验证、提取、处理和分发，使得上层组件不需要了解底层平台，简化上层组件的开发，上层组件专注于自身的业务能力。

