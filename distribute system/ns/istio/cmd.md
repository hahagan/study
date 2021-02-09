### 访问策略

#### get

```shell
kubectl get istiooperator installed-state -n istio-system -o jsonpath='{.spec.meshConfig.outboundTrafficPolicy.mode}'

kubectl get configmap istio -n istio-system -o yaml | grep -o "mode: ALLOW_ANY"
```

#### set

```shell
kubectl get configmap istio -n istio-system -o yaml | sed 's/mode: REGISTRY_ONLY/mode: ALLOW_ANY/g' | kubectl replace -n istio-system -f -
```

