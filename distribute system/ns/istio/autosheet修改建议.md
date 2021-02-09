[toc]

# 二、支持分布式链路追踪

## 2.1 链路追踪介绍

微服务架构中系统中各个微服务之间存在复杂的调用关系。一个来自客户端的请求在其业务处理过程中经过了多个微服务进程。我们如果想要对该请求的端到端调用过程进行完整的分析，则必须将该请求经过的所有进程的相关信息都收集起来并关联在一起，这就是“分布式追踪”。

### 2.1.2 术语

OpenTracing是一个分布式追踪标准规范，它定义了一套通用的数据上报接口，提供平台无关、厂商无关的 API，要求各个分布式追踪系统都来实现这套接口，使得开发人员能够方便的添加（或更换）追踪系统的实现。

#### span

Span 是 OpenTracing的逻辑工作单元，具有请求名称、请求开始时间、请求持续时间。Span 会被嵌套并排序以展示服务间的关系。

#### Trace

OpenTracing在微服务系统中记录的完整的请求执行过程，并显示为 `Trace`， `Trace` 是系统的数据/执行路径。一个端到端的 `Trace` 由一个或多个 `Span` 组成。trace和span关系如下

```
单个Trace中，span间的因果关系


        [Span A]  ←←←(the root span)
            |
     +------+------+
     |             |
 [Span B]      [Span C] ←←←(Span C 是 Span A 的孩子节点, ChildOf)
     |             |
 [Span D]      +---+-------+
               |           |
           [Span E]    [Span F] >>> [Span G] >>> [Span H]
                                       ↑
                                       ↑
                                       ↑
                         (Span G 在 Span F 后被调用, FollowsFrom)

```

```
单个Trace中，span间的时间关系

––|–––––––|–––––––|–––––––|–––––––|–––––––|–––––––|–––––––|–> time

 [Span A···················································]
   [Span B··············································]
      [Span D··········································]
    [Span C········································]
         [Span E·······]        [Span F··] [Span G··] [Span H··]
```

### 2.1.3 ref

* https://opentracing-contrib.github.io/opentracing-specification-zh/

* https://www.servicemesher.com/istio-handbook/practice/method-level-tracing.html



## 2.2 istio对链路追踪的要求

isttio控制的服务网格内，每个网格内的pod可以通过sidecar(一个envoy网格代理实例)拦截pod的输入输出流量，并基于一定的规则完成额外的处理而不需要应用实现复杂的逻辑。例如本节所说的分布式链路追踪，envoy会为http请求添加的头部信息，并完成对链路追踪数据的生成。为了获取一个trace内各个请求的span信息，对于同一请求链路的各个请求持有需要有相应的信息。

### 2.2.1 例子

​	假设服务A接收到HTTP请求a，envoy会在请求a中添加一些header，设为H。如果服务A为了完成请求a的任务产生的对服务B的http请求b，那么请求b中的header应附带信息H，这样链路追踪系统才能将请求a和请求b形成关联。

## 2.3 服务代码改造需求

### 2.3.1 服务端转发请求header

如2.2所言为了满足链路追踪的的数据生成。要求每个服务内在请求依赖服务时附带额外的请求header。因此在autosheet服务请求其他服务时需要转发请求上下文。例如以下代码中从请求上下文中获取对应的header，并在调用其他服务时使用该header。

```
var headers = []string{
	"x-request-id",
	"x-b3-traceid",
	"x-b3-spanid",
	"x-b3-parentspanid",
	"x-b3-sampled",
	"x-b3-flags",
	"x-ot-span-context",
	"x-ot-span-context",
	"x-datadog-trace-id",
	"x-datadog-parent-id",
	"x-datadog-sampled",
}
```

```
func (handler *PingHandler) Ping(c *gin.Context) {

	// 获取接收请求的header，并为后续调用设置header
	head := c.Request.Header
	res, err := http.NewRequest("GET", "http://"+dep+"/api/sheet/bar", nil)
	client := &http.Client{}
	for _, k := range headers {
		v := head.Get(k)
		if v != "" {
			res.Header.Set(k, head.Get(k))
		}
	}
	
	resp, err := client.Do(res)
	if err != nil {
		c.Status(500)
	}
	result, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	// end
	
	c.JSON(http.StatusOK, "Pong")
}
```



### 2.3.2 执行难点

#### 原代码改造

autosheet代码涉及对其他服务的请求分布较散，并且有些函数调用时参数传递未传递请求上下文。完成请求header转发的代码调整需要对整体代码比较了解且可能工作量较大。

#### 确定需要转发的请求header

不同分布式追踪系统在envoy上可能会要求转发不同的header信息。即使现在可以确定所使用的分布式追踪系统，但不保证未来是否会更改所使用的追踪系统或与已有子系统产生冲突。又或者是使用的数据采集对头部有映射行为导致真实需要转发的header有所改变

如果未来改变使用的追踪系统后再次调整服务代码工作冗余重复，因此可以**考虑通过转发头配置文件**的方式传递给服务，服务代码在启动时读取配置文件并初始化。例如平台管理者可以通过提供特殊的"configmap/配置接口"提供需要转发的header。

这么做的另一个好处是**有更好的扩展性**，如果未来需要扩展某些功能，并且同样需要转发某些头部信息，那么可以通过调整该配置文件实现扩展。

##### header说明表

表中所列仅供参考，实际情况根据实际提供的配置文件(理由见上文)。

简写说明：

* all: 不论依赖何种系统都应该转发，虽然未必是对应追踪系统所要求的，那么为什么呢？istio的统一要求的，可能是几个追踪系统的要求合集，也可能是某些特殊监控需要，又或者是某种功能的高级扩展
* [openzipkin](https://github.com/openzipkin/b3-propagation)：一种标准，为jaeger和zipkin所使用

| header                  | 追踪系统                                                     |
| ----------------------- | ------------------------------------------------------------ |
| `x-request-id`          | all                                                          |
| `x-b3-traceid`          | [openzipkin](https://github.com/openzipkin/b3-propagation)/all |
| `x-b3-parentspanid`     | [openzipkin](https://github.com/openzipkin/b3-propagation)/all |
| `x-b3-spanid`           | [openzipkin](https://github.com/openzipkin/b3-propagation)/all |
| `x-b3-sampled`          | [openzipkin](https://github.com/openzipkin/b3-propagation)/all |
| `x-b3-flags`            | all                                                          |
| `x-ot-span-context`     | all                                                          |
| `x-cloud-trace-context` | [OpenCensus](https://opencensus.io/)                         |
| `traceparent`           | [OpenCensus](https://opencensus.io/)                         |
| `grpc-trace-bin`        | [OpenCensus](https://opencensus.io/)                         |




## 2.4 实验代码

```
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var dep = os.Getenv("DEP")
var port = os.Getenv("SE_PORT")
var headers = []string{
	"x-request-id",
	"x-b3-traceid",
	"x-b3-spanid",
	"x-b3-parentspanid",
	"x-b3-sampled",
	"x-b3-flags",
	"x-ot-span-context",
	"x-ot-span-context",
	"x-datadog-trace-id",
	"x-datadog-parent-id",
	"x-datadog-sampled",
}

func foo(w http.ResponseWriter, r *http.Request) {
	head := r.Header
	io.WriteString(w, "foo header:\n")

	res, err := http.NewRequest("GET", "http://"+dep+"/api/sheet/bar", nil)
	fmt.Println("access: " + "http://" + dep + "/api/sheet/bar")
	if err != nil {
		io.WriteString(w, err.Error())
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("content-type", "application/json")

	client := &http.Client{}

	for k, i := range head {
		io.WriteString(w, k+": "+strings.Join(i, ", ")+"\n")
		// for _, v := range i {
		// 	res.Header.Add(k, v)
		// }
	}
	fmt.Fprintln(w)

	res.Header.Set("Content-Type", "application/json")
	res.Header.Set("content-type", "application/json")

	for _, k := range headers {
		v := head.Get(k)
		if v != "" {
			res.Header.Set(k, head.Get(k))
			// fmt.Println(k, v)
		}

	}
	resp, _ := client.Do(res)
	result, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		io.WriteString(w, err.Error())
		w.WriteHeader(500)

		return
	}

	_, err = w.Write(result)
	if err != nil {
		io.WriteString(w, err.Error())
		w.WriteHeader(500)
		return
	}

	io.WriteString(w, "\n")
}

func bar(w http.ResponseWriter, r *http.Request) {
	head := r.Header
	io.WriteString(w, "bar header:\n")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("content-type", "application/json")
	for k, i := range head {
		io.WriteString(w, k+": "+strings.Join(i, ", ")+"\n")
	}

}

func main() {
	fmt.Println("start .....")
	for i, k := range headers {
		head := strings.Title(k)
		headers[i] = head
	}
	http.HandleFunc("/api/sheet/foo", foo)
	http.HandleFunc("/api/sheet/bar", bar)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
```

服务中`/api/sheet/foo`接口会调用另一实例的`/api/sheet/bar`，并将两个接口的header信息打印，结果如下
```
root@transformation-849469666c-hk6td:/# curl collaboration-qwe:9800/api/sheet/foo -H "Content-Type: application/json"
foo header:
User-Agent: curl/7.52.1
X-Forwarded-Proto: http
X-Request-Id: 91a26c1e-43c6-955b-8c6e-a3968eceac90
X-Envoy-Attempt-Count: 1
X-Forwarded-Client-Cert: By=spiffe://cluster.local/ns/anyshare/sa/collaboration;Hash=15f0049650ee5e2df24a88817fc6e5f0e059ab951c3cb9d6877217838ab12df1;Subject="";URI=spiffe://cluster.local/ns/anyshare/sa/transformation
Accept: */*
Content-Type: application/json
Content-Length: 0
X-B3-Traceid: a4c23380fef5b7fc0480a780716a0858
X-B3-Spanid: c840f37d00677743
X-B3-Parentspanid: 0480a780716a0858
X-B3-Sampled: 1

bar header:
Content-Type: application/json
X-B3-Parentspanid: 0480a780716a0858
X-B3-Sampled: 1
X-B3-Spanid: c840f37d00677743
X-B3-Traceid: a4c23380fef5b7fc0480a780716a0858
Accept-Encoding: gzip
User-Agent: Go-http-client/1.1
X-Request-Id: 91a26c1e-43c6-955b-8c6e-a3968eceac9
```





# 三、InitContainer解决方案

## 3.1 问题描述

sheet服务在部署时，首先需要一个init-container完成数据库的初始化，随后证时启动sheet服务。由于istio的原因，会造成init-container无法连接外部数据库。

### 3.1.1 istio导致该问题的原因

简单描述就是两个init-container开始执行的顺序和执行完毕的顺序带来的问题。

简单介绍一下，一个pod的流量拦截配置同样是通过一个init-container完成iptables规则设置，将流量转发到对应的sidecar服务的。因此此时sheet的init-container在istio注入的init-container执行完毕前开始执行，那么将会导致init-container不受istio策略控制。

当istio的网络访问限制比较严格的情况下，如果服务需要访问网格外部服务，那么需要将网格外部服务注册为网格内的一个ServiceEntry。如果init-container对数据库的访问依赖于istio的资源对象，那么此时将无法建立与数据库的连接，从而导致pod启动失败。即使init-container不依赖于istio的配置访问数据库，那么同样会发生流量超出网格控制的问题。

备注：类似的在非init-container上同样也有可能发生，sidecar容器可能晚于应用容器启动。不过这个问题根据资料介绍仅在kubernetes 1.6和istio 1.4版本前出现。kubertest 1.6版本后istio根据lifecycle的一定配置策略解决了这个问题。这里作为一个记录。



### 3.2 解决方案

目前想到的解决方案有三种，第一种是在部署时完成数据库的初始化操作，第二种的将init-container移入lifecycle并改造应用容器的启动命令解决，第三种是直接改造应用容器的启动命令。

1. 部署时完成数据库的初始化操作，该方案并非让部署模块完成对应的数据库操作，而是将init-container的工作抽取出来，作为charts里的一个k8s job对象。通过job完成对应工作。在这种方案下，只有每次chart部署时会执行该job，而job在创建pod时会被istio注入init-container对象，从而保证了数据库初始化操作一定晚于网络环境的初始化。这种方案比较适合于对数据库进行升级或第一次初始化，因为仅会在第一次部署或升级时生成新的job，并执行对应操作。但是需要注意的是正式的应用服务需要能够识别到对应的初始化操作是否已经完成，因为job和deployment不存在先后关系，对应的deploy在job操作未完成前不应该继续并提供服务。
2. 将init-container移入lifecycle并改造应用容器的启动命令解决，
3. 修改服务启动脚本/命令，首先完成initcontainer需要完成的工作。

个人比较推荐第一种或第三种，第二种比较复杂。从升级角度来看第一种比较合适，对于每次启动都需要执行的角度来看第三种比较合适(并不需要init-container)。init-container比较适合的使用场景是同一个pod有多个contianer存在，并且这些container都需要相同的初始化工序，此时使用init-container满足多个container的初始化需要，例如多个容器都需要从相同个配置文件种读取配置，而这个配置文件是通过某种规律生成到特定位置的(即使在这种情况下仍然可以通过第三种方案解决，此时将init-container移动到普通container就好，其他container通过监控依赖项，例如某个empty volume内容 是否满足要求决定是否继续进行或休眠等待)。



# 四、chart调整

chart调整主要涉及的内容有pod是否需要被istio注入sidecar以控制网络流量，service名称在istio上需要满足istio的要求。



## 4.1 istio注入问题

在istio中对一个pod的流量处理主要是依赖于对pod网络命名空间下iptables规则的设置做到流量拦截，并启动一个sidecar的envoy服务对拦截的流量进行接收、处理和转发。

istio提供了pod注入的方式完成对普通pod到istio管理pod的升级。即在原pod的基础上添加init-container和istio-proxy(envoy)两个容器与其他配置。

而在实际大部分开发场景下中微服务开发者并不会在意微服务是否部署在服务网格内，往往是客户/部署者/集群管理者/部署架构实施决定(以后简称管理员)是否部署将一个微服务部署到网格内。因此不应该将istio是否注入的工作交给微服务开发者，但是管理者并不一定具备chart的编写或改写能力，即使有一个微服务的部署模板charts由多个团队管理不利于开发交流和效率。简单来说就是**开发者没权利和必要决定服务某些实际状态，而管理者没有能力和必要调整底层配置**

因此可以考虑由**微服务开发者提供可配置接口**，而**管理者提供配置**。

简单来说方法为在charts里的deployment中的pod中增加额外内容，内容如下(**对应values格式可以再讨论讨论**)

```
# 原deployment的pod配置
---
 metadata:
     annotations:
       checksum/configmap: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
     labels:
       app: {{ .Release.Name }}
   spec:
     {{- with .Values.nodeSelector }}
     nodeSelector:
       {{- toYaml . | nindent 8 }}
 
# 修改后deployment的pod配置
---
   metadata:
     annotations:
       checksum/configmap: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
       {{ with .Values.annotations }}   # add
       {{- toYaml . | nindent 8 }}      # add
       {{- end }}                       # add
     labels:
       app: {{ .Release.Name }}
       {{ with .Values.labels }}   # add
       {{- toYaml . | nindent 8 }}      # add
       {{- end }}                       # add
   spec:
     serviceAccountName: {{ .Release.Name }} # add serviceAccount
     {{- with .Values.nodeSelector }}
     nodeSelector:
       {{- toYaml . | nindent 8 }}
```

当微服务开发提供了这种行为的接口后，管理在完成管理决策并管理集群时即可以通过 `--set-values`的方式(或其他窄接口)对微服务的进行管理。在这种情况下，如果部署在istio上，那么部署模块只需要向values中添加对应的配置，不需要额外管理。

### 4.1.1 额外的好处

1. 更灵活，若未来更改服务网格所使用产品，可能不再需要微服务开发团队进行调整，仅需要部署模块或管理员对values进行调整
2. 配置分离，减少升级会带来的影响。已有的values由微服务团队，升级时可变，额外添加的values由部署模块管理，为升级时不可变。因此升级时不会导致配置回退
3. 有利于后续扩展，如果未来需要进行额外的扩展，可以通过添加额外的注解为pod增加属性，结合k8s的MutatingWebhookConfiguration对象和controller、webhook等机制完成微服务开发者无感知的扩展、注入等
4. 因此其实可以考虑对deployment同样进行相应的label和annotations可注入行为配置。但需要考虑好如何定义格式

### 4.1.2 局限性

* 以上是基于一个charts只有一个deployment的前提
* 如果为多个deployment，目前的chart的values布局也存在一定问题。可以详细讨论



## 4.2 网络配置调整

istio对服务的网络配置有一定的要求，不满足要求的网络配置将会有可能导致网络行为非预期。例如一个端口若是以tcp方式提供将会缺失istio提供的http相关能力，从而无法进行链路追踪。

对于网络调整有两种想法，第一种是根据istio的要求对现有的service配置进行调整。第二种想法是定义一种新的自定义资源对象(不是指k8s的crd)，由集群管理/部署模块为根据新的自定义资源对象完成istio和k8s资源对象的创建和管理。

### 4.3.1 对现有service进行调整

service需要以下规则与调整：

* service和serivice port名称需要以"<协议>-<名词>"或"<名称>"的形式。例如"public-port"变为"public"或"http-pubic"，service名称由"sheet-public"变为"sheet"

*  每个service中增加额外的标签，使得service可以通过标签进行筛选。内容调整类似istio注入。例如

  ```
  apiVersion: v1
  kind: Service
  metadata:
    annotations:
      {{ with .Values.Service.annotations }}   # add
      {{- toYaml . | nindent 8 }}      # add
      {{- end }}                       # add
    name: {{ .Release.Name }}
    namespace: {{ .Values.namespace }}
    labels:
      app: {{ .Release.Name }}
      {{ with .Values.Service.labels }}   # add
      {{- toYaml . | nindent 8 }}      	# add
      {{- end }}                       	# add
  spec:
    ports:
      - name: public
        port: {{ .Values.service.port }}
        protocol: TCP
        targetPort: {{ .Values.service.port }}
        {{- if eq .Values.service.type "NodePort" }}
        nodePort: {{ .Values.service.port }}
        {{- end}}
    selector:
      app: {{ .Release.Name }}
      {{ with .Values.pod.labels }}   	# add
      {{- toYaml . | nindent 8 }}      	# add
      {{- end }}                       	# add
    type: {{ .Values.service.type }}
  ```

### 缺点

1. 需要改动范围较大，涉及charts修改和ingress的修改。因为ingress并非由charts创建，而是由部署模块根据`application.conf`生成。
2. 需要改变已有的东西，易遗漏导致错误



### 4.3.2 不对现有service调整

另一种方案是不对已有配置进行调整，而是定义新的配置声明或使用已有的配置声明，通过配置生成所需要的网络通信所需要的配置。那么即使是已有的配置与istio要求不同，部署模块也可以通过对应的声明创建出能够满足要求的资源对象，并且不需要调整现有chart或改变现有部署代码。

**以下讨论前提是一个chart只负责一个服务(deploy/sts/daemonset)**

#### 配置定义

通过服务配置完成网络配置，需要知道一个服务都提供的访问入口，以目前状态来说就是服务端口、服务协议以及协议相关的详细配置。简单的配置声明可以参考

`service_access.yaml`和`application.conf`进行。

```
	  kind: crdService
	  version: v1	# 为升级兼容做预留
  	  name: sheet	# 服务名称(唯一标识)
	  port=80 	# 服务默认端口
	  protocol:
		type: https-443  # 自定义协议1/自定义协议2
		path=/oauth2/auth,/oauth2/revoke,/oauth2/token,/oauth2/sessions/logout #直接对外的路由
		csr="....."	# 协议相关配置1
		crt="....."	# 协议相关配置2

```

#### 探讨点

1. 为什么不直接使用已有的`application.conf`，这不是已经有了相关定义吗，为什么需要新的定义？

   如果以目前场景，仅通过该配置文件可以满足相关网络配置的创建。但这里的问题在于它是一种集中管理的方案，部署管理的范围是整个集群。而实际上模块化部署的方式不是一个集群/环境为整体的部署方式，扩展不够灵活(例如多版本部署，红绿部署....)，`application.conf`比较适合的是作为单个集群内实际运行状态的描述，提供运行在该集群上的服务获取配置信息。并且这个还会带来开发沟通上的效率问题，一个微服务如果需要调整或新的微服务模块增加，都需要与`application.conf`配置的管理者进行沟通与相关调整，两者间产生了依赖关系。

   所以这里的配置声明与`application.conf`不同，是分散式的，由微服务开发团队负责维护。那么在部署或扩展时，集群控制面在已知服务所能够提供的服务能力情况下可以创建不同的网络配置，为服务的pod注入不同的标签/注解，完成更细粒度的集群管理。而这个过程中微服务开发者并不需要关注。例如微服务多版本部署时，可以通过添加不同的版本标签生成不同的pod和service，同时调整负载策略完成对于的版本控制。例如每次部署不同版本时生成了不同的service和deployment

   ```
   # 以下为伪配置声明
   
   # 版本1：
   ....
   	Kind: deployment
   ....
   	name: sheet-v1
       label: 
       	version: v1
   	    app: sheet
   ---
   ...
   	Kind: Service
   	name: sheetV1
   	label:
   		version: v1
   		app: sheet
   	labelSelector:
   		version: v1
   		appL sheet
   ...
   ---
   
   # 版本2：
   ....
   	Kind: deployment
   ....
   	name: sheet-v2
       label: 
       	version: v2
   	    app: sheet
   ---
   ...
   	Kind: Service
   	name: sheetV1
   	label:
   		version: v2
   		app: sheet
   	labelSelector:
   		version: v2
   		appL sheet
   ...
   ---
   
   # 流量控制策略
   ....
   	kind: VritualService
   	name: sheet
   	match:
   		route: /api/sheet
   		destination:
   			- name: v1
   			  label: 
   			  	version: v1
   			  	app: sheet
   			  weight: 70
   			- name: v2
   			  label:
   			  	version: v2
   			  	app: sheet
   			  weight: 20
   ....
   			
   		
   
   ```

   而由于是配置声明是由微服务开发团队自身管理的，那么即使有所改变或者有新的微服务模块也不需要与部署团队进行沟通，协商开发资源。仅需要了解能够提供什么样的能力。该以什么样的规范完成该配置的声明，这些可以从相关文档中获知。

   其实思想和k8s的crd类似，也许未来可以将其逐渐向crd靠拢。

2.  协议中为什么是自定义协议，单纯的使用https、http、tcp或grpc等不行吗？

   这是因为考虑到每个服务可能对外提供服务的端口不一样，例如有些服务在nginx的443端口提供服务，而有些服务通过nginx的8000端口提供服务(使用不同端口是因为url冲突或者其他原因吗？)。以及某些服务的协议可能系统本身并不支持，需要部署时额外进行配置以完成协议升级或降级等

3. 该支持哪些协议，如何确定协议范围？

   不知道，看需求。例如若自身提供http，而服务间直接访问是默认为https且需要支持，此时按需求做处理。也许每种协议可以作为一个需求或需求下的子任务进行，逐渐扩充支持协议。

4. 具体格式

   讨论讨论



#### 处理流程

那么最终控制面则可以根据该配置文件根据管理员的选择生成所需要配置的配置项。一个微服务的部署流程将会变为

1. 管理员创建集群并将集群注册到集群控制面
2. 管理员通过集群控制面，选择集群部署微服务
   1. 集群控制面将微服务声明的网络配置生成在各个集群创建对应的网络配置，并使得集群间可以相互通信(以下检查为网络配置操作)。

以[预研验证](http://confluence.aishu.cn/pages/viewpage.action?pageId=93173549)的时进行的操作为例子，该例子中的集群控制面可以人为是实验者。

参照以上流程：

1. 第一步是实验者创建了两个集群，集群1部署了anyshare主体服务，集群2部署了k8s服务(注册集群=>实验者知道有两个集群要管理，以及两个集群的特性)。
   1. 实验者根据集群1内的微服务网络声明配置在集群2内创建serviceEntry注册集群1内的已有服务，并配置相关destinationRoute完成集群2到集群1的通信(网络配置操作)，以及集群1对外开放的网络配置。
2. 实验者在集群2部署了系列微服务autosheet，此时服务未完成网络配置，需要根据提供的配置打通两个集群的配置
   1. 实验者在集群2根据autosheet的配置声明，中生成autosheet相关网络配置，例如istio中的VirtualService(url)，destinationRoute(协议、安全配置)等。autosheet能够通过整个serviceEntry(或者说是虚拟IP或虚拟域名)访问其依赖
   2. 实验者调整集群1的网络，将autosheet相关配置注册到集群1中的虚拟对象中，完成集群1到集群2的微服务访问控制(例如将 sheet服务注册到集群1的某个虚拟域名的负载中)。

如果此时集群1需要部署新的微服务，那么流程与上类似，实验者根据微服务声明和集群特性，在集群1内创建网络配置(例如ingress、service等)，并调整集群2的serviceEntry或其他配置使其注册到集群2的虚拟对象负载中。

#### 优点

1. 不需要调整已有, 只需要添加自身的服务声明。未来逐渐取缔`application.conf`的配置管理，改为由集群控制面生成每个集群的`application.conf`或相关网络配置。
2. 未来扩展灵活，更灵活的版本管理，服务管理(可能)
3. 新的配置声明可以作为一层抽象与具体平台分离，即使未来调整所使用系统或增加运行平台，微服务开发团队不需要关注(可能)

#### 缺点

1. 这里所说的一切是以**一个chart管理一个服务**为前提，当一个chart管理多个服务时情况更复杂。因为此时分散管理的配置可以与对应的服务一一对应，可以知道为哪个服务增加标签或注解。





