# Pod 拓扑分布约束

**FEATURE STATE:** `Kubernetes v1.19 [stable]`

你可以使用 *拓扑分布约束（Topology Spread Constraints）* 来控制 [Pods](https://kubernetes.io/docs/concepts/workloads/pods/pod-overview/) 在集群内故障域 之间的分布，例如区域（Region）、可用区（Zone）、节点和其他用户自定义拓扑域。 这样做有助于实现高可用并提升资源利用率。

> **说明：** 在 v1.19 之前的 Kubernetes 版本中，如果要使用 Pod 拓扑扩展约束，你必须在 [API 服务器](https://kubernetes.io/zh/docs/concepts/overview/components/#kube-apiserver) 和[调度器](https://kubernetes.io/zh/docs/reference/command-line-tools-reference/kube-scheduler/) 中启用 `EvenPodsSpread` [特性门控](https://kubernetes.io/zh/docs/reference/command-line-tools-reference/feature-gates/)。

## 先决条件

### 节点标签

拓扑分布约束依赖于节点标签来标识每个节点所在的拓扑域。 例如，某节点可能具有标签：`node=node1,zone=us-east-1a,region=us-east-1`

假设你拥有具有以下标签的一个 4 节点集群：

```
NAME    STATUS   ROLES    AGE     VERSION   LABELS
node1   Ready    <none>   4m26s   v1.16.0   node=node1,zone=zoneA
node2   Ready    <none>   3m58s   v1.16.0   node=node2,zone=zoneA
node3   Ready    <none>   3m17s   v1.16.0   node=node3,zone=zoneB
node4   Ready    <none>   2m43s   v1.16.0   node=node4,zone=zoneB
```

然后从逻辑上看集群如下：

zoneAzoneBNode1Node2Node3Node4

你可以复用在大多数集群上自动创建和填充的 [常用标签](https://kubernetes.io/zh/docs/reference/kubernetes-api/labels-annotations-taints/)， 而不是手动添加标签。

## Pod 的分布约束

### API

`pod.spec.topologySpreadConstraints` 字段定义如下所示：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mypod
spec:
  topologySpreadConstraints:
    - maxSkew: <integer>
      topologyKey: <string>
      whenUnsatisfiable: <string>
      labelSelector: <object>
```

你可以定义一个或多个 `topologySpreadConstraint` 来指示 kube-scheduler 如何根据与现有的 Pod 的关联关系将每个传入的 Pod 部署到集群中。字段包括：

- maxSkew

   

  描述 Pod 分布不均的程度。这是给定拓扑类型中任意两个拓扑域中 匹配的 pod 之间的最大允许差值。它必须大于零。取决于

   

  ```
  whenUnsatisfiable
  ```

   

  的 取值，其语义会有不同。

  - 当 `whenUnsatisfiable` 等于 "DoNotSchedule" 时，`maxSkew` 是目标拓扑域 中匹配的 Pod 数与全局最小值之间可存在的差异。
  - 当 `whenUnsatisfiable` 等于 "ScheduleAnyway" 时，调度器会更为偏向能够降低 偏差值的拓扑域。

- **topologyKey** 是节点标签的键。如果两个节点使用此键标记并且具有相同的标签值， 则调度器会将这两个节点视为处于同一拓扑域中。调度器试图在每个拓扑域中放置数量 均衡的 Pod。

- whenUnsatisfiable

   

  指示如果 Pod 不满足分布约束时如何处理：

  - `DoNotSchedule`（默认）告诉调度器不要调度。
  - `ScheduleAnyway` 告诉调度器仍然继续调度，只是根据如何能将偏差最小化来对 节点进行排序。

- **labelSelector** 用于查找匹配的 pod。匹配此标签的 Pod 将被统计，以确定相应 拓扑域中 Pod 的数量。 有关详细信息，请参考[标签选择算符](https://kubernetes.io/zh/docs/concepts/overview/working-with-objects/labels/#label-selectors)。

你可以执行 `kubectl explain Pod.spec.topologySpreadConstraints` 命令以 了解关于 topologySpreadConstraints 的更多信息