hello,

When the gateway uses wildcards, the route domain name of the envoy will fix its port, so that it must be accessed through a specific port. 

I want to know how to configure to remove this restriction.

In the gateway configuration, I need to use wildcards, but I can't to limit the ports that clients can access. Because the istio ingress gateway port may be changed by the external load, it will make the user unable to access it.How can I solve this problem.



## 1. Use wildcard in Gateway

My gateway configuration is as follows.

```yaml
apiVersion: v1
items:
- apiVersion: networking.istio.io/v1beta1
  kind: Gateway
  metadata:
    annotations:
      meta.helm.sh/release-name: test
      meta.helm.sh/release-namespace: anyshare
    name: as
    namespace: anyshare
  spec:
    selector:
      istio: ingressgateway
    servers:
    - hosts:
      - '*'
      port:
        name: http
        number: 80
        protocol: HTTP
    - hosts:
      - '*.pre.eisoo.com'
      port:
        name: dns1
        number: 443
        protocol: HTTPS
      tls:
        credentialName: as-credential0
        mode: SIMPLE
    - hosts:
      - '*.eisoo.com'
      port:
        name: dns0
        number: 443
        protocol: HTTPS
      tls:
        credentialName: as-credential
        mode: SIMPLE
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""

```

My virtualservice configuration is as follows.

```yaml
apiVersion: v1
items:
- apiVersion: networking.istio.io/v1beta1
  kind: VirtualService
  metadata:
    annotations:
      meta.helm.sh/release-name: test
      meta.helm.sh/release-namespace: anyshare
    name: autosheet
    namespace: anyshare
  spec:
    gateways:
    - anyshare/as
    hosts:
    - '*'
    http:
    - match:
      - uri:
          prefix: /api/collaboration
      route:
      - destination:
          host: collaboration
          port:
            number: 9800
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""

```

The actual route configuration of the envoy of istio ingress gateway is as follows.

```json
[
    {
        "name": "https.443.dns0.as.anyshare",
        "virtualHosts": [
            {
                "name": "*.eisoo.com:443",
                "domains": [
                    "*.eisoo.com",
                    "*.eisoo.com:443"
                ],
                "routes": [
                    {
                        "match": {
                            "prefix": "/api/collaboration",
                            "caseSensitive": true
                        },
                        "route": {
                            "cluster": "outbound|9800||collaboration.anyshare.svc.cluster.local",
                            "timeout": "0s",
                            "retryPolicy": {
                                "retryOn": "connect-failure,refused-stream,unavailable,cancelled,retriable-status-codes",
                                "numRetries": 2,
                                "retryHostPredicate": [
                                    {
                                        "name": "envoy.retry_host_predicates.previous_hosts"
                                    }
                                ],
                                "hostSelectionRetryMaxAttempts": "5",
                                "retriableStatusCodes": [
                                    503
                                ]
                            },
                            "maxStreamDuration": {
                                "maxStreamDuration": "0s"
                            }
                        },
                        "metadata": {
                            "filterMetadata": {
                                "istio": {
                                    "config": "/apis/networking.istio.io/v1alpha3/namespaces/anyshare/virtual-service/autosheet"
                                }
                            }
                        },
                        "decorator": {
                            "operation": "collaboration.anyshare.svc.cluster.local:9800/api/collaboration*"
                        }
                    }
                ],
                "includeRequestAttemptCount": true
            }
        ],
        "validateClusters": false
    }
]
```

So I have to access it through port 443



## 2. Do not use wildcards in Gateway

I tried not to use wildcards on gateway configuration and looked at the routing of envoy in this state. 

I found that the routing configuration here does not lock the port, so I can access the client through any port (the port accessed by the client is mapped to the gateway specific port).

My new gateway configuration is as follows.

```yaml
apiVersion: v1
items:
- apiVersion: networking.istio.io/v1beta1
  kind: Gateway
  metadata:
    annotations:
      meta.helm.sh/release-name: test
      meta.helm.sh/release-namespace: anyshare
    name: as
    namespace: anyshare
  spec:
    selector:
      istio: ingressgateway
    servers:
    - hosts:
      - '*'
      port:
        name: http
        number: 80
        protocol: HTTP
    - hosts:
      - '*.pre.eisoo.com'
      port:
        name: dns1
        number: 443
        protocol: HTTPS
      tls:
        credentialName: as-credential0
        mode: SIMPLE
    - hosts:
      - 'test.eisoo.com'
      port:
        name: dns0
        number: 443
        protocol: HTTPS
      tls:
        credentialName: as-credential
        mode: SIMPLE
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
```

The actual route configuration of the envoy of istio ingress gateway is as follows.

```json
[
    {
        "name": "https.443.dns0.as.anyshare",
        "virtualHosts": [
            {
                "name": "test.eisoo.com:443",
                "domains": [
                    "test.eisoo.com",
                    "test.eisoo.com:*"
                ],
                "routes": [
                    {
                        "match": {
                            "prefix": "/api/collaboration",
                            "caseSensitive": true
                        },
                        "route": {
                            "cluster": "outbound|9800||collaboration.anyshare.svc.cluster.local",
                            "timeout": "0s",
                            "retryPolicy": {
                                "retryOn": "connect-failure,refused-stream,unavailable,cancelled,retriable-status-codes",
                                "numRetries": 2,
                                "retryHostPredicate": [
                                    {
                                        "name": "envoy.retry_host_predicates.previous_hosts"
                                    }
                                ],
                                "hostSelectionRetryMaxAttempts": "5",
                                "retriableStatusCodes": [
                                    503
                                ]
                            },
                            "maxStreamDuration": {
                                "maxStreamDuration": "0s"
                            }
                        },
                        "metadata": {
                            "filterMetadata": {
                                "istio": {
                                    "config": "/apis/networking.istio.io/v1alpha3/namespaces/anyshare/virtual-service/autosheet"
                                }
                            }
                        },
                        "decorator": {
                            "operation": "collaboration.anyshare.svc.cluster.local:9800/api/collaboration*"
                        }
                    }
                ],
                "includeRequestAttemptCount": true
            }
        ],
        "validateClusters": false
    }
]
```



Comparing the two route configurations, we can find that their domain names are `"test.eisoo.com:443"` and `"test.eisoo.com:*`respectively。