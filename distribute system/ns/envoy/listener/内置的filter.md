### ref

- HTTP Inspector
  - [Example](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/http_inspector#example)
  - [Statistics](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/http_inspector#statistics)
- [Original Destination](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/original_dst_filter)
- Original Source
  - [Interaction with Proxy Protocol](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/original_src_filter#interaction-with-proxy-protocol)
  - [IP Version Support](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/original_src_filter#ip-version-support)
  - [Extra Setup](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/original_src_filter#extra-setup)
  - [Example Listener configuration](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/original_src_filter#example-listener-configuration)
- Proxy Protocol
  - [Statistics](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/proxy_protocol#statistics)
- TLS Inspector
  - [Example](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/tls_inspector#example)
  - [Statistics](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/tls_inspector#statistics)



### HTTP inspector

该filter可应用协议进行识别，并进一步识别出http协议为http/1或http/2。往往用于通过协议选择filterchain



### Original Destination

从套接字中读取SO_ORIGINAL_DST 选项并保存。用于恢复被iptables REDIRECT 或 TPROXY 改写的目的地址。

[参考补充](https://www.ichenfu.com/2019/04/09/istio-inbond-interception-and-linux-transparent-proxy/)

### Original Source

与upstream连接时，使用downstream的源IP将envoy伪装为downstream。