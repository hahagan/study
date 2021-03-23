[toc]

### 一、移入istio网格内的charts调整：

1. service和serivice port名称需要以"<协议>-<名词>"或"<名称>"的形式，当前为“public-port"

2. 多service标签使用同一标签，由于某些资源的对象的绑定基于标签(例如 istio的virtualservice)或未来的多版本多集群管理，可能需要能够从标签区分service

3. 如果一个 Pod 同时属于多个 service 那么这些 Service 不能同时在一个端口号上使用不同的协议（比如：HTTP 和 TCP）

4. **带有 app 和 version 标签（label）的 Deployment**: 建议显式地给 Deployment 加上 `app` 和 `version` 标签。给使用 Kubernetes `Deployment` 部署的 Pod 部署配置中增加这些标签，可以给 Istio 收集的指标和遥测信息中增加上下文信息。

5. 根据使用需要charts中可能需要增加标签注入，注解注入。提供可注入便于使用者提供全局参数。
   1. 注解注入可以使部署时注入特殊用途的注解，例如istio是否需要注入
   2. 标签注入同上，例如模块管理
   
   ```
   # 原deployment配置
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
   
   # 测试时添加的配置
   ----
   
       metadata:
         annotations:
           checksum/configmap: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
           {{ with .Values.annotations }}    # add
           {{- toYaml . | nindent 8 }}       # 
           {{- end }}                        # 
         labels:
           app: {{ .Release.Name }}
       spec:
         serviceAccountName: {{ .Release.Name }}  # add
         {{- with .Values.nodeSelector }}
         nodeSelector:
           {{- toYaml . | nindent 8 }}
   
   ```
   
   
   
6. 增加charts使用的的serviceaccoount。

7. 内部服务间的访问，目前基于https访问nginx负载到对应服务，但对应不服可能只提供http，因此如果需要直连，需要能够提供配置项，而不是硬编码。

```
depServices:
  efastdaemon:
    privateHttpHost: 192.168.79.79
    privateHttpPort: "9080"
  hydra:
    administrativeHost: 192.168.79.79
    administrativePort: "9080"
  ## 如果与sheet间的通信直接改为服务和服务间通信，那么就会造成无法通过配置修改。此时sheet提供http，而客户端仍然通过https访问
  sheet:
    publicHost: 192.168.79.79
    publicPort: "443"

```



8. charts中各个服务的values的depservice，将对应IP改为slb-nginx入口。手工调整很麻烦
9. 基于第7，第8点与slb的调整，考虑对网络访问进行规划，从配置上能生成slb-nginx，charts需要使用的网络配置。例如加入网络集群规划，再最后配置service_access.conf中体现，或者各个服务自行管理配置，更上层的服务解析各个服务配置并集成体现在service_access.conf。
10. 访问IP可参考，k8s域名规划并按需要进行扩展。svc.nampsapce.cluster.local或version.svc.namespace.cluster.location或svc.namespace.cluster.appcluster.platformcluster.location



### 二、slb-nginx调整：

nginx对应服务路由地址由nginx-ingresscontroller调整为istio入口istio-ingressgateway



### 三、istio相关

1. 自动注入默认行为，默认注入或默认不注入。影响服务部署配置，默认不注入，则对有需要的pod需要添加特定注解

2. ingressgateway网关配置和virtualservice配置
   
   1. 考虑基于网络规则自动读取配置生成或由各个服务开发按一定规则声明
   
3. 网关外服务访问，需要设置se将网格外服务注册到网格内，通过网络规划，识别配置生成。**如何规划**

4. grafana、prometheus监控集成，如何将istio的监控集成到已有的监控平台上？

5. 高可用问题，分布式链路追踪、监控数据高可用。istio自身数据目前未发现其自身自带的数据库而是基于k8s平台进行crd管理，可能不暂时不需要考虑(又似乎有某种资源需要)。**待进行**

6. 性能、稳定性、资源。**待评估**
   1. 由于流量网格内流量被sidecar劫持，因此sidecar的不稳定因素也增加了服务的不稳定性
   2. 性能，官方声称每个请求会增加延时6ms，待验证但以此为前提。但是需要注意一个服务如果是延时敏感型(或实时要求较高)型服务则需要慎重考虑是否进行sidecar流量拦截，或者数据间有逻辑顺序无法并行处理的服务不适合
      1. 什么是延时敏感型，我也不知道。那什么服务可以放入istio服务网格，有sidecar进行流量劫持进行额外处理
         1. 例子一，elasticsearch写入不算。原因假设某服务单线程写入es性能为450 op/ms，那么即使引入6ms延时，该性能下降至64 op/ms，为了使得该服务恢复原写入性能，此时只需要增加6个io线程就好，虽然增加了线程，但大部分时间下这些线程处于休眠状态，并不会太过于影响该服务的性，即使是python写的(前提是io行为和cpu行为分开)。
         2. 例子二，实时性要求不高的服务。例如用户通过前端查看某个表单，多个6ms和没多又有什么区别。但是有一种极端的假设情况需要考虑延时的影响，那就是超时时限的设置(例子不是很恰当，只是这个意思)，如果一个接口原本的超时时间设置为10s，而在服务繁忙时这个接口响应时间接近9.9s，并且该接口需要机械能深度分为10其他网格内服务调用，此时相应时间将有可能超过超时时限。因此需要评估服务进入网格后对超时时间的影响。
      2. 请求间有逻辑顺序要求。例如大量有序数据写入行为不适合。考虑某个服务A需要请求服务B共n次，且这n次请求业务上不能并行处理，需要顺序处理。那么对比进行流量劫持前，整个业务需要额外花费的时间为6n ms, 例如某种机器学习任务需要从其他服务分批请求数据，数据和数据间有顺序逻辑，需要获取当前数据后才能根据当前数据特征获取下一批数据。
      3. 
   3. 资源占用。
      1. 第一种资源为流量劫持后进行处理需要对流量进行额外处理、请求指标收集，因此需要额外的计算资源，在服务器负载(业务)较高时可能会有所影响，待深入。
      2. 第二种资源为流量转发需要耗费的额外网络流量(服务->虚拟网卡->sidecar->虚拟网卡->服务)，待研究。
      3. istio自身配置同步的资源占用。目前环境来看不高，但当网格内的负载较高后并不确定。
   
7. 网络影响，pod外无法直接通过svc或pod IP连接网格内服务，需要通过对外网格ingressgateway或网格内服务。原因待探索。

8. **安全认证**，目前自带的认证token与istio不同，导致istio开启认证后连接在sidecar处被拦截，需要取消其认证行为。如何兼容待探索。需要结合接下来如何使用认证考虑与尝试。

9. 安装部署与集成，目前为最简模式部署，**复杂模型**下的集成待尝试。

10. 网格外服务注册ServiceEntry，目前尝试发现，设置的se在**无法使用虚假的域名**进行连接，在**缺少dns服务器的情况下需要指定外部服务的ip地址并且需要为其设置网格内的虚拟IP**。例如以下配置中无法通过域名或IP进行访问，而是通过虚拟IP。这意味我们需要管理虚拟IP

    ```
    kind: ServiceEntry
    metadata:
      name: slb-nginx
      namespace : anyshare
    spec:
      hosts:
        - qwe.asd
      location: MESH_EXTERNAL
      addresses:
      #- 192.168.78.78/32
      - 10.5.5.5/32
      ports:
        - name: http
          number: 80
          protocol: http
      resolution: STATIC
      endpoints:
      - address: 192.168.79.79  # 自定义一个内网的ip
        ports:
          http: 80
    ---
    apiVersion: networking.istio.io/v1alpha3
    kind: ServiceEntry
    metadata:
      name: rds-out
      namespace : anyshare
    spec:
      hosts:
        - qwe.zxc
      location: MESH_EXTERNAL
      addresses:
      - 10.5.5.4/32
      ports:
        - name: tcp
          number: 3320
          protocol: tcp
      resolution: STATIC
      endpoints:
      - address: 192.168.79.79  # 自定义一个内网的ip
        ports:
          tcp: 3320
    ```

    

11. istio使用问题，出现问题，**如何排查**



### 四、服务代码相关

1. 为了满足链路追踪，服务处理请求时，如果多个请求处于同一业务或事务时，如果需要进对其进行链路追踪。则需要在请求后后续服务器时携带接收请求的请求上下文，需要转发的上下文有以下(或更多)
   1. `x-request-id`
   2. `x-b3-traceid`
   3. `x-b3-spanid`
   4. `x-b3-parentspanid`
   5. `x-b3-sampled`
   6. `x-b3-flags`
   7. `x-ot-span-context`







### 四、怪异：

#### 4.1 以下两个大致相同的请求第一个404，第二个可以正常执行

不清楚内部逻辑。但网络表面上是联通的。需要了解其内部依赖。

```
// 返回404
curl 'https://192.168.79.79/api/sheet/v1/sheet/8c791d8b-53c1-11eb-a945-0242e58f7073/form' \
  -H 'Connection: keep-alive' \
  -H 'sec-ch-ua: "Google Chrome";v="87", " Not;A Brand";v="99", "Chromium";v="87"' \
  -H 'Authorization: Bearer xSXz9wgYaGh0E9owDYUa1em3fwMbZzi-J7jx_Lb5vjI.zwe9WQMjEyd0UaAohKFBcpIFbpZJUmvnegyx3PyjsjE' \
  -H 'Accept-Language: zh-CN' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36' \
  -H 'Accept: */*' \
  -H 'Sec-Fetch-Site: same-origin' \
  -H 'Sec-Fetch-Mode: cors' \
  -H 'Sec-Fetch-Dest: empty' \
  -H 'Referer: https://192.168.79.79/autosheets/excel?item_id=8c791d61-53c1-11eb-a945-0242e58f7073' \
  -H 'Cookie: lastVisitedOrigin=https%3A%2F%2F192.168.79.79%3A8080; _csrf=4rHXHIJUqIhgh6rcrqYMYXpG; lang=zh-cn; csrftoken=6I4gLPh8bLObEAZ1HePXOurWyuKxcl4G; clustersid=s%3AtOf8HBNnEXft9EGjYCh-AknNXr4jCqTX.gKOrNOD%2BfeENt7i7ZEtB%2BQb6xcMI52yZ%2F5oDzOPJMqg; clustertoken=uHIR0ZE-1h-wZQdUdQUFn0T6f7ThDO_EiFUKu8ULtHc.M3NN8hAtxfHbYCqyFKLXLXhmo9gfxMHkHuWiITT0KBc; sessionid=t56yu6e14maa7zucn29vmtaa2lcsm64j; consoletoken=MUvhZYEKPG28HJEjUb4gLuAqOKAtdKluwi1ITCDAXQ0.051v7v_KiStT0UhrLI0_GAToLoWSCd0vaZOFyNlB0V8; oauth2_authentication_csrf=MTYxMDM1MTIwMXxEdi1CQkFFQ180SUFBUkFCRUFBQVB2LUNBQUVHYzNSeWFXNW5EQVlBQkdOemNtWUdjM1J5YVc1bkRDSUFJR1F3Wm1WbU5HRXhaR0UxTURRMk1EWmhOakF6WWpWbE9XSXdaakk1T0RReHzU13kdszJdqgf9gJSv1qNnzwP5V0gCmE58yRsnSKVNjA==; oauth2_consent_csrf=MTYxMDM1MTIwNXxEdi1CQkFFQ180SUFBUkFCRUFBQVB2LUNBQUVHYzNSeWFXNW5EQVlBQkdOemNtWUdjM1J5YVc1bkRDSUFJRFk1T0RjNE5HTTRPV05oTVRReFpUZGhOV0ZqTUdGbU16bGxNR0ZtWkdRNHzn-cwm12Ssy40m6NRSXIIHyLijBdHXs6tWP9OEMIDe6Q==; token=xSXz9wgYaGh0E9owDYUa1em3fwMbZzi-J7jx_Lb5vjI.zwe9WQMjEyd0UaAohKFBcpIFbpZJUmvnegyx3PyjsjE; id_token=eyJhbGciOiJSUzI1NiIsImtpZCI6InB1YmxpYzowOWMwOTkwYS1jMGM5LTQyZWEtOTBjNy04OTAzM2NmNmFhMWQiLCJ0eXAiOiJKV1QifQ.eyJhdF9oYXNoIjoiT1JQRlRoRXV3enlmdXpFejhqN296QSIsImF1ZCI6WyIxZmVkZWJlMi01ODQ3LTQwZmEtOWMyMi1mNmM1NzA1MWI5YWMiXSwiYXV0aF90aW1lIjoxNjEwMzUxMjA1LCJleHAiOjE2MTAzNTQ4MDUsImlhdCI6MTYxMDM1MTIwNSwiaXNzIjoiaHR0cHM6Ly8xOTIuMTY4Ljc5Ljc5OjQ0My8iLCJqdGkiOiI0NjFkM2I5Zi04OWZlLTQyMjYtOTRmMC1kZGM5ZDE0YWQ1ZjIiLCJub25jZSI6IiIsInJhdCI6MTYxMDM1MTIwMiwic2lkIjoiZDA1YThjYjEtZTBmYS00MDJhLTgyYWQtNDVmNjc1NDE4YzQ5Iiwic3ViIjoiYjJiMmY3MTAtNTNjMC0xMWViLTg0MzgtMDI0MmU1OGY3MDczIn0.CP3eviXxMbXb_fv3OniclUAiTG-1vhEiErS6g_m5rMCEU4rhZfWFUQkqTIqN21oF2cODOH0reJ2a-98wLu-fEDyE6bSxs1e3jNNj6p_ypcscB16jwWBoOMYGB7IWead7WPNnxRAWEAei6N0BfOcmmMPITWzDV-fyblp_ur2Inn3ZEWnun9Qqik6X74UnbVxchoaL7tycL2i7S4zSKq8RMFlaZzEknsMyzfYN5tueSlilELQe0v1SXQpNIuEV0JoWrBCntmPtJjy5_eWxEDVxSAnBcYxAYZ4zqg59mDI4dKERcTcZkhT_MkPrzQYbXU5jMgzs233emCWjY0o1UTQTWddt_qkdFXFFd7DwWDkA6lC21ruRMQkFpG0GXJ0gprjDadzTt2Isp3uCWItTuYEVhRdqxhT3x85mAFgOkaYI6I3HuAPazJY2WZGCSGB1LSJ-aQ4HzGcfDOfe_7ZNlXx_DyH4NaACbJ8rvUwCn-JmsLCrUJd8xhSlMwMMMatgMYqaGRvnrIWSBgDPQhsKO0heZUgiLR9V_ekpCqsOW_hnYhB7-s95ctBnRXgOMxRunXka3DepWiQ7Xw8RjGThtiOcYjxyBvxELmNd36FPtYL-HiBIm2YCcJrx0id35w6roylBpbkl16bCpmwdQj3YRi_nAubtJ8BmDGYevPe4pB_SVJQ; io=2TYbWRleV2y8-qnLAAAD' \
  --compressed \
  --insecure
  
 // 正常
 curl 'https://192.168.79.79/api/sheet/v1/sheet/8c791d8b-53c1-11eb-a945-0242e58f7073/column' \
  -H 'Connection: keep-alive' \
  -H 'sec-ch-ua: "Google Chrome";v="87", " Not;A Brand";v="99", "Chromium";v="87"' \
  -H 'Authorization: Bearer xSXz9wgYaGh0E9owDYUa1em3fwMbZzi-J7jx_Lb5vjI.zwe9WQMjEyd0UaAohKFBcpIFbpZJUmvnegyx3PyjsjE' \
  -H 'Accept-Language: zh-CN' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36' \
  -H 'Accept: */*' \
  -H 'Sec-Fetch-Site: same-origin' \
  -H 'Sec-Fetch-Mode: cors' \
  -H 'Sec-Fetch-Dest: empty' \
  -H 'Referer: https://192.168.79.79/autosheets/excel?item_id=8c791d61-53c1-11eb-a945-0242e58f7073' \
  -H 'Cookie: lastVisitedOrigin=https%3A%2F%2F192.168.79.79%3A8080; _csrf=4rHXHIJUqIhgh6rcrqYMYXpG; lang=zh-cn; csrftoken=6I4gLPh8bLObEAZ1HePXOurWyuKxcl4G; clustersid=s%3AtOf8HBNnEXft9EGjYCh-AknNXr4jCqTX.gKOrNOD%2BfeENt7i7ZEtB%2BQb6xcMI52yZ%2F5oDzOPJMqg; clustertoken=uHIR0ZE-1h-wZQdUdQUFn0T6f7ThDO_EiFUKu8ULtHc.M3NN8hAtxfHbYCqyFKLXLXhmo9gfxMHkHuWiITT0KBc; sessionid=t56yu6e14maa7zucn29vmtaa2lcsm64j; consoletoken=MUvhZYEKPG28HJEjUb4gLuAqOKAtdKluwi1ITCDAXQ0.051v7v_KiStT0UhrLI0_GAToLoWSCd0vaZOFyNlB0V8; oauth2_authentication_csrf=MTYxMDM1MTIwMXxEdi1CQkFFQ180SUFBUkFCRUFBQVB2LUNBQUVHYzNSeWFXNW5EQVlBQkdOemNtWUdjM1J5YVc1bkRDSUFJR1F3Wm1WbU5HRXhaR0UxTURRMk1EWmhOakF6WWpWbE9XSXdaakk1T0RReHzU13kdszJdqgf9gJSv1qNnzwP5V0gCmE58yRsnSKVNjA==; oauth2_consent_csrf=MTYxMDM1MTIwNXxEdi1CQkFFQ180SUFBUkFCRUFBQVB2LUNBQUVHYzNSeWFXNW5EQVlBQkdOemNtWUdjM1J5YVc1bkRDSUFJRFk1T0RjNE5HTTRPV05oTVRReFpUZGhOV0ZqTUdGbU16bGxNR0ZtWkdRNHzn-cwm12Ssy40m6NRSXIIHyLijBdHXs6tWP9OEMIDe6Q==; token=xSXz9wgYaGh0E9owDYUa1em3fwMbZzi-J7jx_Lb5vjI.zwe9WQMjEyd0UaAohKFBcpIFbpZJUmvnegyx3PyjsjE; id_token=eyJhbGciOiJSUzI1NiIsImtpZCI6InB1YmxpYzowOWMwOTkwYS1jMGM5LTQyZWEtOTBjNy04OTAzM2NmNmFhMWQiLCJ0eXAiOiJKV1QifQ.eyJhdF9oYXNoIjoiT1JQRlRoRXV3enlmdXpFejhqN296QSIsImF1ZCI6WyIxZmVkZWJlMi01ODQ3LTQwZmEtOWMyMi1mNmM1NzA1MWI5YWMiXSwiYXV0aF90aW1lIjoxNjEwMzUxMjA1LCJleHAiOjE2MTAzNTQ4MDUsImlhdCI6MTYxMDM1MTIwNSwiaXNzIjoiaHR0cHM6Ly8xOTIuMTY4Ljc5Ljc5OjQ0My8iLCJqdGkiOiI0NjFkM2I5Zi04OWZlLTQyMjYtOTRmMC1kZGM5ZDE0YWQ1ZjIiLCJub25jZSI6IiIsInJhdCI6MTYxMDM1MTIwMiwic2lkIjoiZDA1YThjYjEtZTBmYS00MDJhLTgyYWQtNDVmNjc1NDE4YzQ5Iiwic3ViIjoiYjJiMmY3MTAtNTNjMC0xMWViLTg0MzgtMDI0MmU1OGY3MDczIn0.CP3eviXxMbXb_fv3OniclUAiTG-1vhEiErS6g_m5rMCEU4rhZfWFUQkqTIqN21oF2cODOH0reJ2a-98wLu-fEDyE6bSxs1e3jNNj6p_ypcscB16jwWBoOMYGB7IWead7WPNnxRAWEAei6N0BfOcmmMPITWzDV-fyblp_ur2Inn3ZEWnun9Qqik6X74UnbVxchoaL7tycL2i7S4zSKq8RMFlaZzEknsMyzfYN5tueSlilELQe0v1SXQpNIuEV0JoWrBCntmPtJjy5_eWxEDVxSAnBcYxAYZ4zqg59mDI4dKERcTcZkhT_MkPrzQYbXU5jMgzs233emCWjY0o1UTQTWddt_qkdFXFFd7DwWDkA6lC21ruRMQkFpG0GXJ0gprjDadzTt2Isp3uCWItTuYEVhRdqxhT3x85mAFgOkaYI6I3HuAPazJY2WZGCSGB1LSJ-aQ4HzGcfDOfe_7ZNlXx_DyH4NaACbJ8rvUwCn-JmsLCrUJd8xhSlMwMMMatgMYqaGRvnrIWSBgDPQhsKO0heZUgiLR9V_ekpCqsOW_hnYhB7-s95ctBnRXgOMxRunXka3DepWiQ7Xw8RjGThtiOcYjxyBvxELmNd36FPtYL-HiBIm2YCcJrx0id35w6roylBpbkl16bCpmwdQj3YRi_nAubtJ8BmDGYevPe4pB_SVJQ; io=2TYbWRleV2y8-qnLAAAD' \
  --compressed \
  --insecure
```

#### 4.2 某些请求503

以下请求在授权有效期会返回503，而在授权过期后返回授权无效。因此猜测为某些依赖未满足，待了解

```
curl 'https://192.168.79.79/api/shared-link/v1/sheet/file/8c791d61-53c1-11eb-a945-0242e58f7073' \
  -H 'Connection: keep-alive' \
  -H 'sec-ch-ua: "Google Chrome";v="87", " Not;A Brand";v="99", "Chromium";v="87"' \
  -H 'Authorization: Bearer xSXz9wgYaGh0E9owDYUa1em3fwMbZzi-J7jx_Lb5vjI.zwe9WQMjEyd0UaAohKFBcpIFbpZJUmvnegyx3PyjsjE' \
  -H 'Accept-Language: zh-CN' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36' \
  -H 'Accept: */*' \
  -H 'Sec-Fetch-Site: same-origin' \
  -H 'Sec-Fetch-Mode: cors' \
  -H 'Sec-Fetch-Dest: empty' \
  -H 'Referer: https://192.168.79.79/autosheets/excel?item_id=8c791d61-53c1-11eb-a945-0242e58f7073' \
  -H 'Cookie: lastVisitedOrigin=https%3A%2F%2F192.168.79.79%3A8080; _csrf=4rHXHIJUqIhgh6rcrqYMYXpG; lang=zh-cn; csrftoken=6I4gLPh8bLObEAZ1HePXOurWyuKxcl4G; clustersid=s%3AtOf8HBNnEXft9EGjYCh-AknNXr4jCqTX.gKOrNOD%2BfeENt7i7ZEtB%2BQb6xcMI52yZ%2F5oDzOPJMqg; clustertoken=uHIR0ZE-1h-wZQdUdQUFn0T6f7ThDO_EiFUKu8ULtHc.M3NN8hAtxfHbYCqyFKLXLXhmo9gfxMHkHuWiITT0KBc; sessionid=t56yu6e14maa7zucn29vmtaa2lcsm64j; consoletoken=MUvhZYEKPG28HJEjUb4gLuAqOKAtdKluwi1ITCDAXQ0.051v7v_KiStT0UhrLI0_GAToLoWSCd0vaZOFyNlB0V8; oauth2_authentication_csrf=MTYxMDM1MTIwMXxEdi1CQkFFQ180SUFBUkFCRUFBQVB2LUNBQUVHYzNSeWFXNW5EQVlBQkdOemNtWUdjM1J5YVc1bkRDSUFJR1F3Wm1WbU5HRXhaR0UxTURRMk1EWmhOakF6WWpWbE9XSXdaakk1T0RReHzU13kdszJdqgf9gJSv1qNnzwP5V0gCmE58yRsnSKVNjA==; oauth2_consent_csrf=MTYxMDM1MTIwNXxEdi1CQkFFQ180SUFBUkFCRUFBQVB2LUNBQUVHYzNSeWFXNW5EQVlBQkdOemNtWUdjM1J5YVc1bkRDSUFJRFk1T0RjNE5HTTRPV05oTVRReFpUZGhOV0ZqTUdGbU16bGxNR0ZtWkdRNHzn-cwm12Ssy40m6NRSXIIHyLijBdHXs6tWP9OEMIDe6Q==; token=xSXz9wgYaGh0E9owDYUa1em3fwMbZzi-J7jx_Lb5vjI.zwe9WQMjEyd0UaAohKFBcpIFbpZJUmvnegyx3PyjsjE; id_token=eyJhbGciOiJSUzI1NiIsImtpZCI6InB1YmxpYzowOWMwOTkwYS1jMGM5LTQyZWEtOTBjNy04OTAzM2NmNmFhMWQiLCJ0eXAiOiJKV1QifQ.eyJhdF9oYXNoIjoiT1JQRlRoRXV3enlmdXpFejhqN296QSIsImF1ZCI6WyIxZmVkZWJlMi01ODQ3LTQwZmEtOWMyMi1mNmM1NzA1MWI5YWMiXSwiYXV0aF90aW1lIjoxNjEwMzUxMjA1LCJleHAiOjE2MTAzNTQ4MDUsImlhdCI6MTYxMDM1MTIwNSwiaXNzIjoiaHR0cHM6Ly8xOTIuMTY4Ljc5Ljc5OjQ0My8iLCJqdGkiOiI0NjFkM2I5Zi04OWZlLTQyMjYtOTRmMC1kZGM5ZDE0YWQ1ZjIiLCJub25jZSI6IiIsInJhdCI6MTYxMDM1MTIwMiwic2lkIjoiZDA1YThjYjEtZTBmYS00MDJhLTgyYWQtNDVmNjc1NDE4YzQ5Iiwic3ViIjoiYjJiMmY3MTAtNTNjMC0xMWViLTg0MzgtMDI0MmU1OGY3MDczIn0.CP3eviXxMbXb_fv3OniclUAiTG-1vhEiErS6g_m5rMCEU4rhZfWFUQkqTIqN21oF2cODOH0reJ2a-98wLu-fEDyE6bSxs1e3jNNj6p_ypcscB16jwWBoOMYGB7IWead7WPNnxRAWEAei6N0BfOcmmMPITWzDV-fyblp_ur2Inn3ZEWnun9Qqik6X74UnbVxchoaL7tycL2i7S4zSKq8RMFlaZzEknsMyzfYN5tueSlilELQe0v1SXQpNIuEV0JoWrBCntmPtJjy5_eWxEDVxSAnBcYxAYZ4zqg59mDI4dKERcTcZkhT_MkPrzQYbXU5jMgzs233emCWjY0o1UTQTWddt_qkdFXFFd7DwWDkA6lC21ruRMQkFpG0GXJ0gprjDadzTt2Isp3uCWItTuYEVhRdqxhT3x85mAFgOkaYI6I3HuAPazJY2WZGCSGB1LSJ-aQ4HzGcfDOfe_7ZNlXx_DyH4NaACbJ8rvUwCn-JmsLCrUJd8xhSlMwMMMatgMYqaGRvnrIWSBgDPQhsKO0heZUgiLR9V_ekpCqsOW_hnYhB7-s95ctBnRXgOMxRunXka3DepWiQ7Xw8RjGThtiOcYjxyBvxELmNd36FPtYL-HiBIm2YCcJrx0id35w6roylBpbkl16bCpmwdQj3YRi_nAubtJ8BmDGYevPe4pB_SVJQ; io=2TYbWRleV2y8-qnLAAAD' \
  --compressed \
  --insecure
```

