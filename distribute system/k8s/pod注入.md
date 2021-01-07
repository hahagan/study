[toc]

# Pod Preset

**FEATURE STATE:** `Kubernetes v1.6 [alpha]`

本文提供了 PodPreset 的概述。 在 Pod 创建时，用户可以使用 PodPreset 对象将特定信息注入 Pod 中，这些信息可以包括 Secret、卷、卷挂载和环境变量。

## 理解 Pod Preset

`Pod Preset` 是一种 API 资源，在 Pod 创建时，用户可以用它将额外的运行时需求信息注入 Pod。 使用[标签选择算符](https://kubernetes.io/zh/docs/concepts/overview/working-with-objects/labels/#label-selectors) 来指定 Pod Preset 所适用的 Pod。

使用 Pod Preset 使得 Pod 模板编写者不必显式地为每个 Pod 设置信息。 这样，使用特定服务的 Pod 模板编写者不需要了解该服务的所有细节。

## 在你的集群中启用 Pod Preset

为了在集群中使用 Pod Preset，必须确保以下几点：

1. 已启用 API 类型 `settings.k8s.io/v1alpha1/podpreset`。 例如，这可以通过在 API 服务器的 `--runtime-config` 配置项中包含 `settings.k8s.io/v1alpha1=true` 来实现。 在 minikube 部署的集群中，启动集群时添加此参数 `--extra-config=apiserver.runtime-config=settings.k8s.io/v1alpha1=true`。

2. 已启用准入控制器 `PodPreset`。 启用的一种方式是在 API 服务器的 `--enable-admission-plugins` 配置项中包含 `PodPreset` 。在 minikube 部署的集群中，启动集群时添加以下参数：

   ```shell
   --extra-config=apiserver.enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota,PodPreset
   ```

## PodPreset 如何工作

Kubernetes 提供了准入控制器 (`PodPreset`)，该控制器被启用时，会将 Pod Preset 应用于接收到的 Pod 创建请求中。 当出现 Pod 创建请求时，系统会执行以下操作：

1. 检索所有可用 `PodPresets` 。
2. 检查 `PodPreset` 的标签选择器与要创建的 Pod 的标签是否匹配。
3. 尝试合并 `PodPreset` 中定义的各种资源，并注入要创建的 Pod。
4. 发生错误时抛出事件，该事件记录了 pod 信息合并错误，同时在 *不注入* `PodPreset` 信息的情况下创建 Pod。
5. 为改动的 Pod spec 添加注解，来表明它被 `PodPreset` 所修改。 注解形如： `podpreset.admission.kubernetes.io/podpreset-<pod-preset 名称>": "<资源版本>"`。

一个 Pod 可能不与任何 Pod Preset 匹配，也可能匹配多个 Pod Preset。 同时，一个 `PodPreset` 可能不应用于任何 Pod，也可能应用于多个 Pod。 当 `PodPreset` 应用于一个或多个 Pod 时，Kubernetes 修改 pod spec。 对于 `Env`、 `EnvFrom` 和 `VolumeMounts` 的改动， Kubernetes 修改 pod 中所有容器 spec，对于卷的改动，Kubernetes 会修改 Pod spec。

> **说明：**
>
> 适当时候，Pod Preset 可以修改 Pod 规范中的以下字段：
>
> - `.spec.containers` 字段
> - `initContainers` 字段

### 为特定 Pod 禁用 Pod Preset

在一些情况下，用户不希望 Pod 被 Pod Preset 所改动，这时，用户可以在 Pod 的 `.spec` 中添加形如 `podpreset.admission.kubernetes.io/exclude: "true"` 的注解



## 使用

https://kubernetes.io/zh/docs/tasks/inject-data-application/podpreset/