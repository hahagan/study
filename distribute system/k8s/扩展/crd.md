# 定制资源

*定制资源（Custom Resource）* 是对 Kubernetes API 的扩展。 本页讨论何时向 Kubernetes 集群添加定制资源，何时使用独立的服务。 本页描述添加定制资源的两种方法以及怎样在二者之间做出抉择。

## 定制资源

*资源（Resource）* 是 [Kubernetes API](https://kubernetes.io/zh/docs/concepts/overview/kubernetes-api/) 中的一个端点， 其中存储的是某个类别的 [API 对象](https://kubernetes.io/zh/docs/concepts/overview/working-with-objects/kubernetes-objects/) 的一个集合。 例如内置的 *pods* 资源包含一组 Pod 对象。

*定制资源（Custom Resource）* 是对 Kubernetes API 的扩展，不一定在默认的 Kubernetes 安装中就可用。定制资源所代表的是对特定 Kubernetes 安装的一种定制。 不过，很多 Kubernetes 核心功能现在都用定制资源来实现，这使得 Kubernetes 更加模块化。

定制资源可以通过动态注册的方式在运行中的集群内或出现或消失，集群管理员可以独立于集群 更新定制资源。一旦某定制资源被安装，用户可以使用 [kubectl](https://kubernetes.io/zh/docs/reference/kubectl/overview/) 来创建和访问其中的对象，就像他们为 *pods* 这种内置资源所做的一样。

## 定制控制器

就定制资源本身而言，它只能用来存取结构化的数据。 当你将定制资源与 *定制控制器（Custom Controller）* 相结合时，定制资源就能够 提供真正的 *声明式 API（Declarative API）*。

使用[声明式 API](https://kubernetes.io/zh/docs/concepts/overview/kubernetes-api/)， 你可以 *声明* 或者设定你的资源的期望状态，并尝试让 Kubernetes 对象的当前状态 同步到其期望状态。控制器负责将结构化的数据解释为用户所期望状态的记录，并 持续地维护该状态。

你可以在一个运行中的集群上部署和更新定制控制器，这类操作与集群的生命周期无关。 定制控制器可以用于任何类别的资源，不过它们与定制资源结合起来时最为有效。 [Operator 模式](https://coreos.com/blog/introducing-operators.html)就是将定制资源 与定制控制器相结合的。你可以使用定制控制器来将特定于某应用的领域知识组织 起来，以编码的形式构造对 Kubernetes API 的扩展。

![自定义资源](https://kubernetes.io/zh/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)



# Operator 模式

Operator 是 Kubernetes 的扩展软件，它利用 [定制资源](https://kubernetes.io/zh/docs/concepts/extend-kubernetes/api-extension/custom-resources/) 管理应用及其组件。 Operator 遵循 Kubernetes 的理念，特别是在[控制器](https://kubernetes.io/zh/docs/concepts/architecture/controller/) 方面。

## 初衷

Operator 模式旨在捕获（正在管理一个或一组服务的）运维人员的关键目标。 负责特定应用和 service 的运维人员，在系统应该如何运行、如何部署以及出现问题时如何处理等方面有深入的了解。

在 Kubernetes 上运行工作负载的人们都喜欢通过自动化来处理重复的任务。 Operator 模式会封装你编写的（Kubernetes 本身提供功能以外的）任务自动化代码。

## Kubernetes 上的 Operator

Kubernetes 为自动化而生。无需任何修改，你即可以从 Kubernetes 核心中获得许多内置的自动化功能。 你可以使用 Kubernetes 自动化部署和运行工作负载， *甚至* 可以自动化 Kubernetes 自身。

Kubernetes [控制器](https://kubernetes.io/zh/docs/concepts/architecture/controller/) 使你无需修改 Kubernetes 自身的代码，即可以扩展集群的行为。 Operator 是 Kubernetes API 的客户端，充当 [定制资源](https://kubernetes.io/zh/docs/concepts/extend-kubernetes/api-extension/custom-resources/) 的控制器。

## Operator 示例

使用 Operator 可以自动化的事情包括：

- 按需部署应用
- 获取/还原应用状态的备份
- 处理应用代码的升级以及相关改动。例如，数据库 schema 或额外的配置设置
- 发布一个 service，要求不支持 Kubernetes API 的应用也能发现它
- 模拟整个或部分集群中的故障以测试其稳定性
- 在没有内部成员选举程序的情况下，为分布式应用选择首领角色

想要更详细的了解 Operator？这儿有一个详细的示例：

1. 有一个名为 SampleDB 的自定义资源，你可以将其配置到集群中。
2. 一个包含 Operator 控制器部分的 Deployment，用来确保 Pod 处于运行状态。
3. Operator 代码的容器镜像。
4. 控制器代码，负责查询控制平面以找出已配置的 SampleDB 资源。
5. Operator 的核心是告诉 API 服务器，如何使现实与代码里配置的资源匹配。
   - 如果添加新的 SampleDB，Operator 将设置 PersistentVolumeClaims 以提供 持久化的数据库存储，设置 StatefulSet 以运行 SampleDB，并设置 Job 来处理初始配置。
   - 如果你删除它，Operator 将建立快照，然后确保 StatefulSet 和 Volume 已被删除。
6. Operator 也可以管理常规数据库的备份。对于每个 SampleDB 资源，Operator 会确定何时创建（可以连接到数据库并进行备份的）Pod。这些 Pod 将依赖于 ConfigMap 和/或具有数据库连接详细信息和凭据的 Secret。
7. 由于 Operator 旨在为其管理的资源提供强大的自动化功能，因此它还需要一些 额外的支持性代码。在这个示例中，代码将检查数据库是否正运行在旧版本上， 如果是，则创建 Job 对象为你升级数据库。

## 部署 Operator

部署 Operator 最常见的方法是将自定义资源及其关联的控制器添加到你的集群中。 跟运行容器化应用一样，控制器通常会运行在 [控制平面](https://kubernetes.io/zh/docs/reference/glossary/?all=true#term-control-plane) 之外。 例如，你可以在集群中将控制器作为 Deployment 运行。

## 使用 Operator

部署 Operator 后，你可以对 Operator 所使用的资源执行添加、修改或删除操作。 按照上面的示例，你将为 Operator 本身建立一个 Deployment，然后：

```shell
kubectl get SampleDB                   # 查找所配置的数据库

kubectl edit SampleDB/example-database # 手动修改某些配置
```

可以了！Operator 会负责应用所作的更改并保持现有服务处于良好的状态。

## 编写你自己的 Operator

如果生态系统中没可以实现你目标的 Operator，你可以自己编写代码。在 [接下来](https://kubernetes.io/zh/docs/concepts/extend-kubernetes/operator/#what-s-next)一节中，你会找到编写自己的云原生 Operator 需要的库和工具的链接。

你还可以使用任何支持 [Kubernetes API 客户端](https://kubernetes.io/zh/docs/reference/using-api/client-libraries/) 的语言或运行时来实现 Operator（即控制器）。