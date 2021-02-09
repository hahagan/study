### 问题描述

根据istio多网络的主从结构[安装文档](https://istio.io/latest/docs/setup/install/multicluster/multi-primary_multi-network/#set-the-default-network-for-cluster2)安装istio，在第二个网络安装remote的istio后发现istiod没有正确设置istio的CLUSTER_ID

https://istio.io/latest/docs/setup/install/multicluster/multi-primary_multi-network/#set-the-default-network-for-cluster2



### 操作流程

如文档中执行，在部署cluster2时，但对配置进行调整，降低了istiod中的resource.requests。两个集群配置如下分别对应文档中的cluster1和cluster2

```
### primary/cluster1 cluster
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  values:
    global:
      meshID: pri
      multiCluster:
        clusterName: local
      network: pri
```

```
### cluster2 cluster
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  profile: remote
  values:
    global:
      meshID: pri
      multiCluster:
        clusterName: cluster2
      network: network2
      remotePilotAddress: 192.168.79.78
  components:
    ingressGateways:
    - name: istio-ingressgateway
      k8s:
        resources:
          requests:
            cpu: 0m
            memory: 40Mi

    pilot:
      k8s:
        env:
          - name: PILOT_TRACE_SAMPLING
            value: "100"
        resources:
          requests:
            cpu: 0m
            memory: 100Mi
  values:
    global:
      imagePullPolicy: "IfNotPresent"
      proxy:
        resources:
          requests:
            cpu: 0m
            memory: 40Mi

```



### 问题发现过程

1. 在部署cluster2的east-west gateway的集群时，对应pod通过健康检查。日志如下

2. 根据日志提示查看istiod日志信息，日志如下
3. 根据istiod日志提示，查看istiod源码，发现为`istio.io/istio/pilot/pkg/xds/authentication.go`代码`func (s *DiscoveryServer) authenticate(ctx context.Context) ([]string, error)`方法在无法通过认证后打印的日志。
4. 在

