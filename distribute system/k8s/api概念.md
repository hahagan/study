## 标准 API 术语

大多数 Kubernetes API 资源类型都是 [对象](https://kubernetes.io/zh/docs/concepts/overview/working-with-objects/kubernetes-objects/#kubernetes-objects)： 它们代表的是集群中某一概念的具体实例，例如一个 Pod 或名字空间。 为数不多的几个 API 资源类型是“虚拟的” - 它们通常代表的是操作而非对象本身， 例如访问权限检查（使用 POST 请求发送一个 JSON 编码的 `SubjectAccessReview` 负载到 `subjectaccessreviews` 资源）。 所有对象都有一个唯一的名字，以便支持幂等的创建和检视操作，不过如果虚拟资源类型 不可检视或者不要求幂等，可以不具有唯一的名字。

Kubernetes 一般会利用标准的 RESTful 术语来描述 API 概念：

- **资源类型（Resource Type）** 是在 URL 中使用的名称（`pods`、`namespaces`、`services`）
- 所有资源类型都有具有一个 JSON 形式（其对象的模式定义）的具体表示，称作**类别（Kind）**
- 某资源类型的实例的列表称作 **集合（Collection）**
- 资源类型的单个实例被称作 **资源（Resource）**

所有资源类型要么是集群作用域的（`/apis/GROUP/VERSION/*`），要么是名字空间 作用域的（`/apis/GROUP/VERSION/namespaces/NAMESPACE/*`）。 名字空间作用域的资源类型会在其名字空间被删除时也被删除，并且对该资源类型的 访问是由定义在名字空间域中的授权检查来控制的。

某些资源类型有一个或多个子资源（Sub-resource），表现为对应资源下面的子路径：

- 集群作用域的子资源：`GET /apis/GROUP/VERSION/RESOURCETYPE/NAME/SUBRESOURCE`
- 名字空间作用域的子资源：`GET /apis/GROUP/VERSION/namespaces/NAMESPACE/RESOURCETYPE/NAME/SUBRESOURCE`

取决于对象是什么，每个子资源所支持的动词有所不同 - 参见 API 文档以了解更多信息。 跨多个资源来访问其子资源是不可能的 - 如果需要这一能力，则通常意味着需要一种 新的虚拟资源类型了。

## 高效检测变更

为了使客户端能够构造一个模型来表达集群的当前状态，所有 Kubernetes 对象资源类型 都需要支持一致的列表和一个称作 **watch** 的增量变更通知信源（feed）。 每个 Kubernetes 对象都有一个 `resourceVersion` 字段，代表该资源在下层数据库中 存储的版本。检视资源集合（名字空间作用域或集群作用域）时，服务器返回的响应 中会包含 `resourceVersion` 值，可用来向服务器发起 watch 请求。 服务器会返回所提供的 `resourceVersion` 之后发生的所有变更（创建、删除和更新）。 这使得客户端能够取回当前的状态并监视其变更，且不会错过任何变更事件。 客户端的监视连接被断开时，可以从最后返回的 `resourceVersion` 重启新的监视连接， 或者执行一个新的集合请求之后从头开始监视操作。

给定的 Kubernetes 服务器只会保留一定的时间内发生的历史变更列表。 使用 etcd3 的集群默认保存过去 5 分钟内发生的变更。 当所请求的 watch 操作因为资源的历史版本不存在而失败，客户端必须能够处理 因此而返回的状态代码 `410 Gone`，清空其本地的缓存，重新执行 list 操作， 并基于新的 list 操作所返回的 `resourceVersion` 来开始新的 watch 操作。

## 资源版本

资源版本采用字符串来表达，用来标示对象的服务器端内部版本。 客户端可以使用资源版本来判定对象是否被更改，或者在读取、列举或监视资源时 用来表达数据一致性需求。 客户端必需将资源版本视为不透明的对象，将其原封不动地传递回服务器端。 例如，客户端一定不能假定资源版本是某种数值标识，也不可以对两个资源版本值 进行比较看其是否相同（也就是不可以比较两个版本值以判断其中一个比另一个 大或小）。

### `metadata` 中的 `resourceVersion`

客户端可以在资源中看到资源版本信息，这里的资源包括从服务器返回的 Watch 事件 以及 list 操作响应：

[v1.meta/ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta) - 资源 的 `metadata.resourceVersion` 值标明该实例上次被更改时的资源版本。

[v1.meta/ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#listmeta-v1-meta) - 资源集合 （即 list 操作的响应）的 `metadata.resourceVersion` 所标明的是 list 响应被构造 时的资源版本。

### `resourceVersion` 参数

GET、LIST 和 WATCH 操作都支持 `resourceVersion` 参数。

参数的具体含义取决于所执行的操作和所给的 `resourceVersion` 值：

对于 GET 和 LIST 而言，资源版本的语义为：

**GET：**

| resourceVersion 未设置 | resourceVersion="0" | resourceVersion="<非零值>" |
| ---------------------- | ------------------- | -------------------------- |
| 最新版本               | 任何版本            | 不老于给定版本             |

**LIST：**

v1.19 及以上版本的 API 服务器支持 `resourceVersionMatch` 参数，用以确定如何对 LIST 调用应用 resourceVersion 值。 强烈建议在为 LIST 调用设置了 `resourceVersion` 时也设置 `resourceVersionMatch`。 如果 `resourceVersion` 未设置，则 `resourceVersionMatch` 是不允许设置的。 为了向后兼容，客户端必须能够容忍服务器在某些场景下忽略 `resourceVersionMatch` 的行为：

- 当设置 `resourceVersionMatch=NotOlderThan` 且指定了 `limit` 时，客户端必须能够 处理 HTTP 410 "Gone" 响应。例如，客户端可以使用更新一点的 `resourceVersion` 来重试，或者回退到 `resourceVersion=""` （即允许返回任何版本）。
- 当设置了 `resourceVersionMatch=Exact` 且未指定 `limit` 时，客户端必须验证 响应数据中 `ListMeta` 的 `resourceVersion` 与所请求的 `resourceVersion` 匹配， 并处理二者可能不匹配的情况。例如，客户端可以重试设置了 `limit` 的请求。

除非你对一致性有着非常强烈的需求，使用 `resourceVersionMatch=NotOlderThan` 同时为 `resourceVersion` 设定一个已知值是优选的交互方式，因为与不设置 `resourceVersion` 和 `resourceVersionMatch` 相比，这种配置可以取得更好的 集群性能和可扩缩性。后者需要提供带票选能力的读操作。

| resourceVersionMatch 参数             | 分页参数                    | resourceVersion 未设置  | resourceVersion="0"                   | resourceVersion="<非零值>"            |
| ------------------------------------- | --------------------------- | ----------------------- | ------------------------------------- | ------------------------------------- |
| resourceVersionMatch 未设置           | limit 未设置                | 最新版本                | 任意版本                              | 不老于指定版本                        |
| resourceVersionMatch 未设置           | limit=<n>, continue 未设置  | 最新版本                | 任意版本                              | 精确匹配                              |
| resourceVersionMatch 未设置           | limit=<n>, continue=<token> | 从 token 开始、精确匹配 | 非法请求，视为从 token 开始、精确匹配 | 非法请求，返回 HTTP `400 Bad Request` |
| resourceVersionMatch=Exact [1]        | limit 未设置                | 非法请求                | 非法请求                              | 精确匹配                              |
| resourceVersionMatch=Exact [1]        | limit=<n>, continue 未设置  | 非法请求                | 非法请求                              | 精确匹配                              |
| resourceVersionMatch=NotOlderThan [1] | limit 未设置                | 非法请求                | 任意版本                              | 不老于指定版本                        |
| resourceVersionMatch=NotOlderThan [1] | limit=<n>, continue 未设置  | 非法请求                | 任意版本                              | 不老于指定版本                        |

**脚注：**

[1] 如果服务器无法正确处理 `resourceVersionMatch` 参数，其行为与未设置该参数相同。

GET 和 LIST 操作的语义含义如下：

- **最新版本：** 返回资源版本为最新的数据。所返回的数据必须一致 （通过票选读操作从 etcd 中取出）。
- **任意版本：** 返回任意资源版本的数据。优选最新可用的资源版本，不过不能保证 强一致性；返回的数据可能是任何资源版本的。请求返回的数据有可能是客户端以前 看到过的很老的资源版本。尤其在某些高可用配置环境中，网络分区或者高速缓存 未被更新等状态都可能导致这种状况。不能容忍这种不一致性的客户端不应采用此 语义。

- **不老于指定版本：** 返回至少比所提供的 `resourceVersion` 还要新的数据。 优选最新的可用数据，不过最终提供的可能是不老于所给 `resourceVersion` 的任何版本。 对于发给能够正确处理 `resourceVersionMatch` 参数的服务器的 LIST 请求，此语义 保证 `ListMeta` 中的 `resourceVersion` 不老于请求的 `resourceVersion`，不过 不对列表条目之 `ObjectMeta` 的 `resourceVersion` 提供任何保证。 这是因为 `ObjectMeta.resourceVersion` 所跟踪的是列表条目对象上次更新的时间， 而不是对象被返回时是否是最新。
- **确定版本：** 返回精确匹配所给资源版本的数据。如果所指定的 resourceVersion 的数据不可用，服务器会响应 HTTP 410 "Gone"。 对于发送给能够正确处理 `resourceVersionMatch` 参数的服务器的 LIST 请求而言， 此语义会保证 ListMeta 中的 `resourceVersion` 与所请求的 `resourceVersion` 匹配， 不过不对列表条目之 `ObjectMeta` 的 `resourceVersion` 提供任何保证。 这是因为 `ObjectMeta.resourceVersion` 所跟踪的是列表条目对象上次更新的时间， 而不是对象被返回时是否是最新。
- **Continue 令牌、精确匹配：** 返回原先带分页参数的 LIST 调用中指定的资源版本的数据。 在最初的带分页参数的 LIST 调用之后，所有分页式的 LIST 调用都使用所返回的 Continue 令牌来跟踪最初提供的资源版本，

对于 WATCH 操作而言，资源版本的语义如下：

**WATCH：**

| resourceVersion 未设置   | resourceVersion="0"      | resourceVersion="<非零值>" |
| ------------------------ | ------------------------ | -------------------------- |
| 读取状态并从最新版本开始 | 读取状态并从任意版本开始 | 从指定版本开始             |

WATCH 操作语义的含义如下：

- **读取状态并从最新版本开始：** 从最新的资源版本开始 WATCH 操作。这里的 最新版本必须是一致的（即通过票选读操作从 etcd 中取出）。为了建立初始状态， WATCH 首先会处理一组合成的 "Added" 事件，这些事件涵盖在初始资源版本中存在 的所有资源实例。 所有后续的 WATCH 事件都是关于 WATCH 开始时所处资源版本之后发生的变更。

- **读取状态并从任意版本开始：** 警告：通过这种方式初始化的 WATCH 操作可能会 返回任何状态的停滞数据。请在使用此语义之前执行复核，并在可能的情况下采用其他 语义。此语义会从任意资源版本开始执行 WATCH 操作，优选最新的可用的资源版本， 不过不是必须的；采用任何资源版本作为起始版本都是被允许的。 WATCH 操作有可能起始于客户端已经观测到的很老的版本。在高可用配置环境中，因为 网络分裂或者高速缓存未及时更新的原因都会造成此现象。 如果客户端不能容忍这种不一致性，就不要使用此语义来启动 WATCH 操作。 为了建立初始状态，WATCH 首先会处理一组合成的 "Added" 事件，这些事件涵盖在 初始资源版本中存在的所有资源实例。 所有后续的 WATCH 事件都是关于 WATCH 开始时所处资源版本之后发生的变更。

- **从指定版本开始：** 从某确切资源版本开始执行 WATCH 操作。WATCH 事件都是 关于 WATCH 开始时所处资源版本之后发生的变更。与前面两种语义不同，WATCH 操作 开始的时候不会生成或处理为所提供资源版本合成的 "Added" 事件。 我们假定客户端既然能够提供确切资源版本，就应该已经拥有了起始资源版本对应的初始状态。

### "410 Gone" 响应

服务器不需要提供所有老的资源版本，在客户端请求的是早于服务器端所保留版本的 `resourceVersion` 时，可以返回 HTTP `410 (Gone)` 状态码。 客户端必须能够容忍 `410 (Gone)` 响应。 参阅[高效检测变更](https://kubernetes.io/zh/docs/reference/using-api/api-concepts/#efficient-detection-of-changes)以了解如何在监测资源时 处理 `410 (Gone)` 响应。

如果所请求的 `resourceVersion` 超出了可应用的 `limit`，那么取决于请求是否 是通过高速缓存来满足的，API 服务器可能会返回一个 `410 Gone` HTTP 响应。

### 不可用的资源版本

服务器不必未无法识别的资源版本提供服务。针对无法识别的资源版本的 LIST 和 GET 请求 可能会短暂等待，以期资源版本可用。如果所给的资源版本在一定的时间段内仍未变得 可用，服务器应该超时并返回 `504 (Gateway Timeout)`，且可在响应中添加 `Retry-After` 响应头部字段，标明客户端在再次尝试之前应该等待多少秒钟。 目前，`kube-apiserver` 也能使用 `Too large resource version（资源版本过高）` 消息来标识这类响应。针对某无法识别的资源版本的 WATCH 操作可能会无限期 （直到请求超时）地等待下去，直到资源版本可用。