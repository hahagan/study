# 一、现象描述

两个主节点的k8s集群，一个deployment在副本数为1时在pod控制面板下出现多个pod副本。其中一个为running，另一个为pending。如图

![](E:\study\distribute system\k8s\images\qa_helm_upgrade_multiPod.png)



# 二、原因

在遇到该表现时第一时间的猜测是由于滚动升级时，新的pod没有正常启动提供服务，此时旧的pod不会被删除，因此查看k8s提供的event日志。日志(超过一定期限的event被删除了，未来需要考虑对event进行采集)如下。

在上图中发现其中有一个pod副本是pending状态，这种状态往往意味着pod启动资源不满足，结合下文的event可以得出结论此时deployment的某个pod由于一个节点没有足够的`requests cpu`资源，一个节点不满足pod的亲和性导致该pod处于`pending`状态。

继续查看是否存在滚动升级，由于产品使用的`helm`进行部署的，因此查看部署状态，发现确实存在多个版本的`helm chart`。由此可以确定确实是发生了滚动升级。并且查看系统日志，确实发现了多次的`helm upgrade`行为。

由此可以得出的**结论**，由于滚动升级时，资源`cpu`不足导致新的版本无法正常部署处于`pending`状态，老版本不会被清理，导致一个deployment存在两个pod副本。

注：这台机器似乎并不文档，机器有时很卡，有时很流畅。原因不明

```shell
### helm
[root@node-1-24 pods]# helm ls 
NAME                       	REVISION	UPDATED                 	STATUS  	CHART                                                      	APP VERSION    	NAMESPACE
authentication             	3       	Wed Feb  3 02:10:25 2021	DEPLOYED	authentication-1.0.0-release-mission-alita-m2              	1.0.0          	default  
class-19121                	1       	Tue Feb  2 11:08:15 2021	DEPLOYED	nginx-ingress-controller-1.1.3                             	0.32.0         	anyshare 
class-19530                	1       	Tue Feb  2 11:08:13 2021	DEPLOYED	nginx-ingress-controller-1.1.3                             	0.32.0         	anyshare 
class-30002                	1       	Tue Feb  2 11:08:13 2021	DEPLOYED	nginx-ingress-controller-1.1.3                             	0.32.0         	anyshare 
class-32066                	1       	Tue Feb  2 11:08:17 2021	DEPLOYED	nginx-ingress-controller-1.1.3                             	0.32.0         	anyshare 
class-8500                 	1       	Tue Feb  2 11:08:12 2021	DEPLOYED	nginx-ingress-controller-1.1.3                             	0.32.0         	anyshare 
class-9001                 	1       	Tue Feb  2 11:08:16 2021	DEPLOYED	nginx-ingress-controller-1.1.3                             	0.32.0         	anyshare 
class-9200                 	1       	Tue Feb  2 11:08:14 2021	DEPLOYED	nginx-ingress-controller-1.1.3                             	0.32.0         	anyshare 
class-9301                 	1       	Tue Feb  2 11:08:18 2021	DEPLOYED	nginx-ingress-controller-1.1.3                             	0.32.0         	anyshare 
doc-sync                   	3       	Wed Feb  3 02:09:47 2021	DEPLOYED	doc-sync-1.0.0-release-mission-alita-m2                    	1.0.0          	default  
document                   	3       	Wed Feb  3 02:09:43 2021	DEPLOYED	document-1.0.0-release-mission-alita-m2                    	1.0.0          	default  
document-domain-management 	3       	Wed Feb  3 02:10:39 2021	DEPLOYED	document-domain-management-1.0.0-release-mission-console-m2	1.0.0          	default  
document-sync-scheduling   	3       	Wed Feb  3 02:10:23 2021	DEPLOYED	document-sync-scheduling-1.0.0-release-mission-console-m2  	1.0.0          	default  
ecron-analysis             	3       	Wed Feb  3 02:09:40 2021	DEPLOYED	ecron-analysis-1.0.0-release-mission-alita-m2              	1.0.0          	default  
ecron-management           	3       	Wed Feb  3 02:09:52 2021	DEPLOYED	ecron-management-1.0.0-release-mission-alita-m2            	1.0.0          	default  
hydra                      	3       	Wed Feb  3 02:09:45 2021	DEPLOYED	hydra-1.0.0                                                	v1.1.0_oryOS.13	anyshare 
ingress-manager            	1       	Tue Feb  2 11:07:53 2021	DEPLOYED	ingress-manager-1.1.3                                      	1.1.3-6        	default  
license                    	1       	Tue Feb  2 11:20:57 2021	DEPLOYED	license-1.0.0-release-mission-deployment-m2                	1.0            	default  
oauth2-ui                  	3       	Wed Feb  3 02:09:49 2021	DEPLOYED	oauth2-ui-1.0.0-release-mission-client-web-m2              	1.0.0          	default  
policy-management          	3       	Wed Feb  3 02:10:50 2021	DEPLOYED	policy-management-1.0.0-release-mission-console-m2         	1.0.0          	default  
proton-eceph-config-manager	3       	Wed Feb  3 02:10:57 2021	DEPLOYED	proton-eceph-config-manager-7.0.1                          	1.0            	default  
proton-eceph-config-web    	3       	Wed Feb  3 02:11:01 2021	DEPLOYED	proton-eceph-config-web-7.0.1                              	1.0            	default  
proton-mq-nsq              	1       	Tue Feb  2 11:08:35 2021	DEPLOYED	proton-mq-nsq-1.0.1                                        	1.0.1-4        	default  
proton-policy-engine       	3       	Wed Feb  3 02:10:44 2021	DEPLOYED	proton-policy-engine-1.0.0                                 	1.0            	default  
read-policy                	3       	Wed Feb  3 02:10:42 2021	DEPLOYED	read-policy-1.0.0-release-mission-console-m2               	1.0.0          	default  
shared-link                	3       	Wed Feb  3 02:10:47 2021	DEPLOYED	shared-link-1.0.0-release-mission-alita-m2                 	1.0.0          	default  
user-management            	3       	Wed Feb  3 02:10:28 2021	DEPLOYED	user-management-1.0.0-release-mission-alita-m2             	1.0.0          	default  
webportal                  	2       	Wed Feb  3 02:10:06 2021	DEPLOYED	webportal-1.0.0-release-mission-client-web-m2              	1.0.0          	default  
webservice                 	3       	Wed Feb  3 02:10:09 2021	DEPLOYED	webservice-1.0.0-release-mission-client-web-m2             	1.0.0          	default 
```



```shell
### k8s的event信息

LAST SEEN   TYPE      REASON              OBJECT                                            MESSAGE
50m         Warning   FailedScheduling    pod/authentication-58dd4bfdbd-qpxww               0/2 nodes are available: 1 Insufficient cpu, 1 node(s) didn't match Pod's node affinity.
23m         Normal    Scheduled           pod/authentication-58dd4bfdbd-qpxww               Successfully assigned anyshare/authentication-58dd4bfdbd-qpxww to node-1-24
23m         Normal    Pulled              pod/authentication-58dd4bfdbd-qpxww               Container image "registry.aishu.cn:15000/as/authentication-db:7.0.0-release-mission-alita-m2" already present on machine
23m         Normal    Created             pod/authentication-58dd4bfdbd-qpxww               Created container init-authentication-database
23m         Normal    Started             pod/authentication-58dd4bfdbd-qpxww               Started container init-authentication-database
23m         Normal    Pulled              pod/authentication-58dd4bfdbd-qpxww               Container image "registry.aishu.cn:15000/as/authentication:7.0.0-release-mission-alita-m2" already present on machine
23m         Normal    Created             pod/authentication-58dd4bfdbd-qpxww               Created container authentication
23m         Normal    Started             pod/authentication-58dd4bfdbd-qpxww               Started container authentication
22m         Normal    Killing             pod/authentication-5fcf6df884-hchjn               Stopping container authentication
22m         Normal    SuccessfulDelete    replicaset/authentication-5fcf6df884              Deleted pod: authentication-5fcf6df884-hchjn
22m         Normal    ScalingReplicaSet   deployment/authentication                         Scaled down replica set authentication-5fcf6df884 to 0
50m         Warning   FailedScheduling    pod/document-domain-management-5459cdd89b-w5v6j   0/2 nodes are available: 1 Insufficient cpu, 1 node(s) didn't match Pod's node affinity.
25m         Normal    Scheduled           pod/document-domain-management-5459cdd89b-w5v6j   Successfully assigned anyshare/document-domain-management-5459cdd89b-w5v6j to node-1-24
25m         Normal    Pulled              pod/document-domain-management-5459cdd89b-w5v6j   Container image "registry.aishu.cn:15000/as/document-domain-management:1.0.0-release-mission-console-m2" already present on machine
25m         Normal    Created             pod/document-domain-management-5459cdd89b-w5v6j   Created container document-domain-management
25m         Normal    Started             pod/document-domain-management-5459cdd89b-w5v6j   Started container document-domain-management
25m         Normal    Killing             pod/document-domain-management-85b854f45-vs99f    Stopping container document-domain-management
25m         Normal    SuccessfulDelete    replicaset/document-domain-management-85b854f45   Deleted pod: document-domain-management-85b854f45-vs99f
25m         Normal    ScalingReplicaSet   deployment/document-domain-management             Scaled down replica set document-domain-management-85b854f45 to 0
55m         Warning   Unhealthy           pod/document-sync-scheduling-756b79dfc4-5dm74     Readiness probe failed: Get "http://192.169.0.36:9800/api/document-sync-scheduling/v1/health/ready": dial tcp 192.169.0.36:9800: connect: connection refused
30m         Warning   BackOff             pod/document-sync-scheduling-756b79dfc4-5dm74     Back-off restarting failed container
27m         Normal    Killing             pod/document-sync-scheduling-c78c85bf4-vl28g      Stopping container document-sync-scheduling
27m         Normal    SuccessfulDelete    replicaset/document-sync-scheduling-c78c85bf4     Deleted pod: document-sync-scheduling-c78c85bf4-vl28g
27m         Normal    ScalingReplicaSet   deployment/document-sync-scheduling               Scaled down replica set document-sync-scheduling-c78c85bf4 to 0
38m         Warning   Unhealthy           pod/ecron-management-5598cd95d9-v865v             Liveness probe failed: Get "http://192.169.0.39:9800/health/alive": dial tcp 192.169.0.39:9800: connect: connection refused
63m         Warning   Unhealthy           pod/ecron-management-5598cd95d9-v865v             Readiness probe failed: Get "http://192.169.0.39:9800/health/ready": dial tcp 192.169.0.39:9800: connect: connection refused
28m         Warning   BackOff             pod/ecron-management-5598cd95d9-v865v             Back-off restarting failed container
25m         Normal    Killing             pod/ecron-management-5d74cdbcfc-bndvs             Stopping container ecron-management
25m         Normal    SuccessfulDelete    replicaset/ecron-management-5d74cdbcfc            Deleted pod: ecron-management-5d74cdbcfc-bndvs
25m         Normal    ScalingReplicaSet   deployment/ecron-management                       Scaled down replica set ecron-management-5d74cdbcfc to 0
18m         Warning   Unhealthy           pod/license-55xfm                                 Liveness probe failed: Get "http://192.169.1.14:8090/api/license/health": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
19m         Warning   Unhealthy           pod/nginx-ingress-controller-class-19530-lr7db    Liveness probe failed: Get "http://192.169.1.5:10254/healthz": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
19m         Warning   Unhealthy           pod/nginx-ingress-controller-class-8500-jgxwn     Readiness probe failed: Get "http://192.169.1.4:10254/healthz": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
19m         Warning   Unhealthy           pod/nginx-ingress-controller-class-9200-rqwq8     Liveness probe failed: Get "http://192.169.1.7:10254/healthz": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
24m         Normal    Killing             pod/policy-management-6ff88b4d9d-7hfh8            Stopping container policy-management
24m         Warning   Unhealthy           pod/policy-management-6ff88b4d9d-7hfh8            Liveness probe failed: Get "http://192.169.0.28:9800/api/policy-management/v1/health/alive": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
24m         Normal    SuccessfulDelete    replicaset/policy-management-6ff88b4d9d           Deleted pod: policy-management-6ff88b4d9d-7hfh8
50m         Warning   FailedScheduling    pod/policy-management-77dc9c55c7-grph4            0/2 nodes are available: 1 Insufficient cpu, 1 node(s) didn't match Pod's node affinity.
24m         Normal    Scheduled           pod/policy-management-77dc9c55c7-grph4            Successfully assigned anyshare/policy-management-77dc9c55c7-grph4 to node-1-24
24m         Normal    Pulled              pod/policy-management-77dc9c55c7-grph4            Container image "registry.aishu.cn:15000/as/policy-management:1.0.0-release-mission-console-m2" already present on machine
24m         Normal    Created             pod/policy-management-77dc9c55c7-grph4            Created container policy-management
24m         Normal    Started             pod/policy-management-77dc9c55c7-grph4            Started container policy-management
24m         Normal    ScalingReplicaSet   deployment/policy-management                      Scaled down replica set policy-management-6ff88b4d9d to 0
19m         Warning   Unhealthy           pod/proton-mq-nsq-nsqlookupd-0                    Liveness probe failed: Get "http://192.169.1.12:4161/ping": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
64m         Normal    Pulled              pod/proton-policy-engine-77bb94955b-2v82x         Container image "registry.aishu.cn:15000/proton/busybox:1.31.1" already present on machine
30m         Warning   BackOff             pod/proton-policy-engine-77bb94955b-2v82x         Back-off restarting failed container
25m         Normal    Killing             pod/read-policy-5978cf4d7f-9cvp5                  Stopping container read-policy
25m         Warning   Unhealthy           pod/read-policy-5978cf4d7f-9cvp5                  Readiness probe failed: Get "http://192.169.0.25:9800/api/read-policy/v1/health/ready": dial tcp 192.169.0.25:9800: i/o timeout (Client.Timeout exceeded while awaiting headers)
25m         Normal    SuccessfulDelete    replicaset/read-policy-5978cf4d7f                 Deleted pod: read-policy-5978cf4d7f-9cvp5
50m         Warning   FailedScheduling    pod/read-policy-7dff49cfcd-8hhpg                  0/2 nodes are available: 1 Insufficient cpu, 1 node(s) didn't match Pod's node affinity.
25m         Normal    Scheduled           pod/read-policy-7dff49cfcd-8hhpg                  Successfully assigned anyshare/read-policy-7dff49cfcd-8hhpg to node-1-24
25m         Normal    Pulled              pod/read-policy-7dff49cfcd-8hhpg                  Container image "registry.aishu.cn:15000/as/read-policy:1.0.0-release-mission-console-m2" already present on machine
25m         Normal    Created             pod/read-policy-7dff49cfcd-8hhpg                  Created container read-policy
25m         Normal    Started             pod/read-policy-7dff49cfcd-8hhpg                  Started container read-policy
25m         Normal    ScalingReplicaSet   deployment/read-policy                            Scaled down replica set read-policy-5978cf4d7f to 0
50m         Warning   FailedScheduling    pod/shared-link-bc67f6bb4-2txjl                   0/2 nodes are available: 1 Insufficient cpu, 1 node(s) didn't match Pod's node affinity.
25m         Normal    Scheduled           pod/shared-link-bc67f6bb4-2txjl                   Successfully assigned anyshare/shared-link-bc67f6bb4-2txjl to node-1-24
25m         Normal    Pulled              pod/shared-link-bc67f6bb4-2txjl                   Container image "registry.aishu.cn:15000/as/shared-link-db:7.0.0-release-mission-alita-m2" already present on machine
25m         Normal    Created             pod/shared-link-bc67f6bb4-2txjl                   Created container init-sharedlink-database
25m         Normal    Started             pod/shared-link-bc67f6bb4-2txjl                   Started container init-sharedlink-database
25m         Normal    Pulled              pod/shared-link-bc67f6bb4-2txjl                   Container image "registry.aishu.cn:15000/as/shared-link:7.0.0-release-mission-alita-m2" already present on machine
25m         Normal    Created             pod/shared-link-bc67f6bb4-2txjl                   Created container shared-link
25m         Normal    Started             pod/shared-link-bc67f6bb4-2txjl                   Started container shared-link
24m         Normal    Killing             pod/shared-link-c874f8bd8-hhw9k                   Stopping container shared-link
24m         Warning   Unhealthy           pod/shared-link-c874f8bd8-hhw9k                   Liveness probe failed: Get "http://192.169.0.27:9800/health/alive": dial tcp 192.169.0.27:9800: connect: connection refused
24m         Normal    SuccessfulDelete    replicaset/shared-link-c874f8bd8                  Deleted pod: shared-link-c874f8bd8-hhw9k
24m         Normal    ScalingReplicaSet   deployment/shared-link                            Scaled down replica set shared-link-c874f8bd8 to 0
21m         Normal    Killing             pod/user-management-6c6b88469f-pdvvb              Stopping container user-management
21m         Warning   Unhealthy           pod/user-management-6c6b88469f-pdvvb              Liveness probe failed: Get "http://192.169.0.23:9800/health/alive": dial tcp 192.169.0.23:9800: i/o timeout (Client.Timeout exceeded while awaiting headers)
21m         Normal    SuccessfulDelete    replicaset/user-management-6c6b88469f             Deleted pod: user-management-6c6b88469f-pdvvb
49m         Warning   FailedScheduling    pod/user-management-ff67c745f-6sdq7               0/2 nodes are available: 1 Insufficient cpu, 1 node(s) didn't match Pod's node affinity.
22m         Normal    Scheduled           pod/user-management-ff67c745f-6sdq7               Successfully assigned anyshare/user-management-ff67c745f-6sdq7 to node-1-24
22m         Normal    Pulled              pod/user-management-ff67c745f-6sdq7               Container image "registry.aishu.cn:15000/as/user-management-db:7.0.0-release-mission-alita-m2" already present on machine
22m         Normal    Created             pod/user-management-ff67c745f-6sdq7               Created container init-user-management-database
22m         Normal    Started             pod/user-management-ff67c745f-6sdq7               Started container init-user-management-database
22m         Normal    Pulled              pod/user-management-ff67c745f-6sdq7               Container image "registry.aishu.cn:15000/as/user-management:7.0.0-release-mission-alita-m2" already present on machine
22m         Normal    Created             pod/user-management-ff67c745f-6sdq7               Created container user-management
22m         Normal    Started             pod/user-management-ff67c745f-6sdq7               Started container user-management
21m         Normal    ScalingReplicaSet   deployment/user-management                        Scaled down replica set user-management-6c6b88469f to 0
```



# 三、可能的调整建议

该问题的原因是由于`requests.cpu`资源不满足导致pod无法部署。而一般的应用场景中`requests`代表的是系统上负载的下限，一般情况下系统的负载应该保持在整体负载的70%-80%，超过这个部分应该触发告警进行资源扩充或其他操作，考虑这一点所有的`requests`和不应该超过系统的70%(如果考虑其他非POD形式的应用存在，应该进一步下调这个比例。当然特殊情况除外)。

k8s的的资源计算如果没有特殊配置则默认使用对应节点的总资源(例如当前环境)，出现该问题后参考上文所述，可以考虑对`request`进行调整。

`resources`分为`requests`和`limits`，分别代表下限和运行是分配比例。对于特殊应用(比较重要的应用)则应该考虑将两者设为相同值，防止在系统负载较高时被驱逐。对于一般应用则需要考虑应用的最低资源需求以设置`requests`。以及运行时要求设置`limit`。例如如果应用时java程序，配置pod资源时需要考虑结合jvm配置，否则很容易造成oom。

更详细的`resources`说明可以查看[官方文档](https://kubernetes.io/zh/docs/concepts/configuration/manage-resources-containers/)