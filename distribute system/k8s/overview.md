[toc]

![](https://d33wubrfki0l68.cloudfront.net/2475489eaf20163ec0f54ddc1d92aa8d4c87c96b/e7c81/images/docs/components-of-kubernetes.svg)

## 控制平面组件（Control Plane Components）

控制平面的组件对集群做出全局决策(比如调度)，以及检测和响应集群事件（例如，当不满足部署的 `replicas` 字段时，启动新的 [pod](https://kubernetes.io/docs/concepts/workloads/pods/pod-overview/)）。

控制平面组件可以在集群中的任何节点上运行。 然而，为了简单起见，设置脚本通常会在同一个计算机上启动所有控制平面组件，并且不会在此计算机上运行用户容器。 请参阅[构建高可用性集群](https://kubernetes.io/zh/docs/setup/production-environment/tools/kubeadm/high-availability/) 中对于多主机 VM 的设置示例。

### kube-apiserver

API 服务器是 Kubernetes [控制面](https://kubernetes.io/zh/docs/reference/glossary/?all=true#term-control-plane)的组件， 该组件公开了 Kubernetes API。 API 服务器是 Kubernetes 控制面的前端。

Kubernetes API 服务器的主要实现是 [kube-apiserver](https://kubernetes.io/zh/docs/reference/command-line-tools-reference/kube-apiserver/)。 kube-apiserver 设计上考虑了水平伸缩，也就是说，它可通过部署多个实例进行伸缩。 你可以运行 kube-apiserver 的多个实例，并在这些实例之间平衡流量。

### etcd

etcd 是兼具一致性和高可用性的键值数据库，可以作为保存 Kubernetes 所有集群数据的后台数据库。

您的 Kubernetes 集群的 etcd 数据库通常需要有个备份计划。

要了解 etcd 更深层次的信息，请参考 [etcd 文档](https://etcd.io/docs/)

### kube-scheduler

主节点上的组件，该组件监视那些新创建的未指定运行节点的 Pod，并选择节点让 Pod 在上面运行。

调度决策考虑的因素包括单个 Pod 和 Pod 集合的资源需求、硬件/软件/策略约束、亲和性和反亲和性规范、数据位置、工作负载间的干扰和最后时限。

### kube-controller-manager

在主节点上运行 [控制器](https://kubernetes.io/zh/docs/concepts/architecture/controller/) 的组件。

从逻辑上讲，每个[控制器](https://kubernetes.io/zh/docs/concepts/architecture/controller/)都是一个单独的进程， 但是为了降低复杂性，它们都被编译到同一个可执行文件，并在一个进程中运行。

这些控制器包括:

- 节点控制器（Node Controller）: 负责在节点出现故障时进行通知和响应。
- 副本控制器（Replication Controller）: 负责为系统中的每个副本控制器对象维护正确数量的 Pod。
- 端点控制器（Endpoints Controller）: 填充端点(Endpoints)对象(即加入 Service 与 Pod)。
- 服务帐户和令牌控制器（Service Account & Token Controllers）: 为新的命名空间创建默认帐户和 API 访问令牌

## Node 组件

节点组件在每个节点上运行，维护运行的 Pod 并提供 Kubernetes 运行环境。

### kubelet

一个在集群中每个节点上运行的代理。 它保证容器都运行在 Pod 中。

kubelet 接收一组通过各类机制提供给它的 PodSpecs，确保这些 PodSpecs 中描述的容器处于运行状态且健康。 kubelet 不会管理不是由 Kubernetes 创建的容器。

### kube-proxy

[kube-proxy](https://kubernetes.io/zh/docs/reference/command-line-tools-reference/kube-proxy/) 是集群中每个节点上运行的网络代理， 实现 Kubernetes [服务（Service）](https://kubernetes.io/zh/docs/concepts/services-networking/service/) 概念的一部分。

kube-proxy 维护节点上的网络规则。这些网络规则允许从集群内部或外部的网络会话与 Pod 进行网络通信。

如果操作系统提供了数据包过滤层并可用的话，kube-proxy 会通过它来实现网络规则。否则， kube-proxy 仅转发流量本身。

### 容器运行时（Container Runtime）

容器运行环境是负责运行容器的软件。

Kubernetes 支持多个容器运行环境: [Docker](https://kubernetes.io/zh/docs/reference/kubectl/docker-cli-to-kubectl/)、 [containerd](https://containerd.io/docs/)、[CRI-O](https://cri-o.io/#what-is-cri-o) 以及任何实现 [Kubernetes CRI (容器运行环境接口)](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-node/container-runtime-interface.md)。

## 插件（Addons）

插件使用 Kubernetes 资源（[DaemonSet](https://kubernetes.io/zh/docs/concepts/workloads/controllers/daemonset/)、 [Deployment](https://kubernetes.io/zh/docs/concepts/workloads/controllers/deployment/)等）实现集群功能。 因为这些插件提供集群级别的功能，插件中命名空间域的资源属于 `kube-system` 命名空间。

下面描述众多插件中的几种。有关可用插件的完整列表，请参见 [插件（Addons）](https://kubernetes.io/zh/docs/concepts/cluster-administration/addons/)。

### DNS

尽管其他插件都并非严格意义上的必需组件，但几乎所有 Kubernetes 集群都应该 有[集群 DNS](https://kubernetes.io/zh/docs/concepts/services-networking/dns-pod-service/)， 因为很多示例都需要 DNS 服务。

集群 DNS 是一个 DNS 服务器，和环境中的其他 DNS 服务器一起工作，它为 Kubernetes 服务提供 DNS 记录。

Kubernetes 启动的容器自动将此 DNS 服务器包含在其 DNS 搜索列表中。

### Web 界面（仪表盘）

[Dashboard](https://kubernetes.io/zh/docs/tasks/access-application-cluster/web-ui-dashboard/) 是Kubernetes 集群的通用的、基于 Web 的用户界面。 它使用户可以管理集群中运行的应用程序以及集群本身并进行故障排除。

### 容器资源监控

[容器资源监控](https://kubernetes.io/zh/docs/tasks/debug-application-cluster/resource-usage-monitoring/) 将关于容器的一些常见的时间序列度量值保存到一个集中的数据库中，并提供用于浏览这些数据的界面。

### 集群层面日志

[集群层面日志](https://kubernetes.io/zh/docs/concepts/cluster-administration/logging/) 机制负责将容器的日志数据 保存到一个集中的日志存储中，该存储能够提供搜索和浏览接口。



# 控制器

在机器人技术和自动化领域，控制回路（Control Loop）是一个非终止回路，用于调节系统状态。

这是一个控制环的例子：房间里的温度自动调节器。

当你设置了温度，告诉了温度自动调节器你的*期望状态（Desired State）*。 房间的实际温度是*当前状态（Current State）*。 通过对设备的开关控制，温度自动调节器让其当前状态接近期望状态。

在 Kubernetes 中，控制器通过监控[集群](https://kubernetes.io/zh/docs/reference/glossary/?all=true#term-cluster) 的公共状态，并致力于将当前状态转变为期望的状态。

## 控制器模式

一个控制器至少追踪一种类型的 Kubernetes 资源。这些 [对象](https://kubernetes.io/zh/docs/concepts/overview/working-with-objects/kubernetes-objects/) 有一个代表期望状态的 `spec` 字段。 该资源的控制器负责确保其当前状态接近期望状态。

控制器可能会自行执行操作；在 Kubernetes 中更常见的是一个控制器会发送信息给 [API 服务器](https://kubernetes.io/zh/docs/reference/command-line-tools-reference/kube-apiserver/)，这会有副作用。 具体可参看后文的例子。

### 通过 API 服务器来控制

[Job](https://kubernetes.io/zh/docs/concepts/workloads/controllers/job/) 控制器是一个 Kubernetes 内置控制器的例子。 内置控制器通过和集群 API 服务器交互来管理状态。

Job 是一种 Kubernetes 资源，它运行一个或者多个 [Pod](https://kubernetes.io/docs/concepts/workloads/pods/pod-overview/)， 来执行一个任务然后停止。 （一旦[被调度了](https://kubernetes.io/zh/docs/concepts/scheduling-eviction/)，对 `kubelet` 来说 Pod 对象就会变成了期望状态的一部分）。

在集群中，当 Job 控制器拿到新任务时，它会保证一组 Node 节点上的 `kubelet` 可以运行正确数量的 Pod 来完成工作。 Job 控制器不会自己运行任何的 Pod 或者容器。Job 控制器是通知 API 服务器来创建或者移除 Pod。 [控制面](https://kubernetes.io/zh/docs/reference/glossary/?all=true#term-control-plane)中的其它组件 根据新的消息作出反应（调度并运行新 Pod）并且最终完成工作。

创建新 Job 后，所期望的状态就是完成这个 Job。Job 控制器会让 Job 的当前状态不断接近期望状态：创建为 Job 要完成工作所需要的 Pod，使 Job 的状态接近完成。

控制器也会更新配置对象。例如：一旦 Job 的工作完成了，Job 控制器会更新 Job 对象的状态为 `Finished`。

（这有点像温度自动调节器关闭了一个灯，以此来告诉你房间的温度现在到你设定的值了）。

### 直接控制

相比 Job 控制器，有些控制器需要对集群外的一些东西进行修改。

例如，如果你使用一个控制回路来保证集群中有足够的 [节点](https://kubernetes.io/zh/docs/concepts/architecture/nodes/)，那么控制器就需要当前集群外的 一些服务在需要时创建新节点。

和外部状态交互的控制器从 API 服务器获取到它想要的状态，然后直接和外部系统进行通信 并使当前状态更接近期望状态。

（实际上有一个[控制器](https://github.com/kubernetes/autoscaler/) 可以水平地扩展集群中的节点。请参阅

这里，很重要的一点是，控制器做出了一些变更以使得事物更接近你的期望状态， 之后将当前状态报告给集群的 API 服务器。 其他控制回路可以观测到所汇报的数据的这种变化并采取其各自的行动。

在温度计的例子中，如果房间很冷，那么某个控制器可能还会启动一个防冻加热器。 就 Kubernetes 集群而言，控制面间接地与 IP 地址管理工具、存储服务、云驱动 APIs 以及其他服务协作，通过[扩展 Kubernetes](https://kubernetes.io/zh/docs/concepts/extend-kubernetes/) 来实现这点。

## 期望状态与当前状态

Kubernetes 采用了系统的云原生视图，并且可以处理持续的变化。

在任务执行时，集群随时都可能被修改，并且控制回路会自动修复故障。 这意味着很可能集群永远不会达到稳定状态。

只要集群中控制器的在运行并且进行有效的修改，整体状态的稳定与否是无关紧要的。

## 设计

作为设计原则之一，Kubernetes 使用了很多控制器，每个控制器管理集群状态的一个特定方面。 最常见的一个特定的控制器使用一种类型的资源作为它的期望状态， 控制器管理控制另外一种类型的资源向它的期望状态演化。

使用简单的控制器而不是一组相互连接的单体控制回路是很有用的。 控制器会失败，所以 Kubernetes 的设计正是考虑到了这一点。

> **说明：**
>
> 可以有多个控制器来创建或者更新相同类型的对象。 在后台，Kubernetes 控制器确保它们只关心与其控制资源相关联的资源。
>
> 例如，你可以创建 Deployment 和 Job；它们都可以创建 Pod。 Job 控制器不会删除 Deployment 所创建的 Pod，因为有信息 （[标签](https://kubernetes.io/zh/docs/concepts/overview/working-with-objects/labels/)）让控制器可以区分这些 Pod。

## 运行控制器的方式

Kubernetes 内置一组控制器，运行在 [kube-controller-manager](https://kubernetes.io/docs/reference/generated/kube-controller-manager/) 内。 这些内置的控制器提供了重要的核心功能。

Deployment 控制器和 Job 控制器是 Kubernetes 内置控制器的典型例子。 Kubernetes 允许你运行一个稳定的控制平面，这样即使某些内置控制器失败了， 控制平面的其他部分会接替它们的工作。

你会遇到某些控制器运行在控制面之外，用以扩展 Kubernetes。 或者，如果你愿意，你也可以自己编写新控制器。 你可以以一组 Pod 来运行你的控制器，或者运行在 Kubernetes 之外。 最合适的方案取决于控制器所要执行的功能是什么。