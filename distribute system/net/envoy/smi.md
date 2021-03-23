SMI 项目本身不涉及具体的服务网格实现，只是试图定义通用标准规范。SMI 不会定义服务网格的具体功能范围，它只是一个通用子集。SMI 具体应该提供什么？官方网站上描述如下：

- 为运行在 Kubernetes 上的服务网格提供了标准接口定义
- 为最通用的服务网格使用场景提供了基本功能合集
- 为服务网格提供了随时间推移不断支持新实现的灵活性
- 依赖于服务网格技术来为生态系统提供了创新的空间



规范目前涵盖的网格能力范围以及接口实现：

- Traffic Policy（流量策略）：在不同的服务之间应用身份和传输加密等策略
  - Traffic Access Control（流量访问控制）：根据客户端的身份标识来配置对特定 [pod](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#pod) 的访问以及路由，以将应用的访问锁定在可被允许的用户和服务。
  - Traffic Specs（流量规范）：定义了基于不同协议的流量表示方式，这些资源通过与访问控制和其它策略协同工作以在协议级别进行流量管理。
- Traffic Telemetry（流量遥测）：捕获例如错误率、服务间调用延迟等关键指标。
  - Traffic Metrics（流量指标）：为控制台和自动扩缩容工具暴露一些通用的流量指标。
- Traffic Management（流量管理）：在不同服务之间进行流量切换。
  - Traffic Split（流量分割）：通过在不同服务之间逐步调整流量百分比，以帮服务进行金丝雀发布。



SMI 之所以在提出的时候就表明只为基于 Kubernetes 之上的服务网格提供规范，是因为 SMI 是通过 Kubernetes Custom Resource Definitions（[CRD](https://www.servicemesher.com/istio-handbook/GLOSSARY.html#crd)）和 Extension API Servers 来实现的。