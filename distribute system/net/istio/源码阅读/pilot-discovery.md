```
configStoreCache.queue <--event--- kube controller <---watch/list-- kube crd

configStoreCache.queue.run <--event-- configStoreCache.queue
 	|----> configStoreCache.handler(event)---- rep ---> DiscoveryServer.pushChannel

pushQueue <----equeue PushRequest----- `handleUpdates` <---pull PushRequest------ DiscoveryServer.pushChannel

con.pushChannel <--------`doSendPushes` <---dequue PushRequest ------ pushQueue

`pushConnection`<---StreamAggregatedResources <---discovery.DiscoveryRequest

	

```



### 笔记

istio通过创建controller对象监听kubectl中的对象，将对象的变化转换为event对象，存入configStoreCache的队列queImpl::queue.Instance中。

crdclient运行会启动队列queImpl的运行，队列queImpl在实现上会不断的从队列中读取event，并调用`cacheHandler.onEvent`。

`cacheHandler.onEvent`会调用注册的handler，而handler将imforer提交的对象转换为request，推送到discovery的请求队列中。

最终dsicovery将队列中的请求推送到下游的envoy。

