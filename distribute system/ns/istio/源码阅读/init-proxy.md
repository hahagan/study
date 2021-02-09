# 一、初始化行为

在istio的1.8.1版本中，init-proxy行为由`pilot-agent`程序通过`istio-iptables`命令完成对容器的iptables规则初始化。

```
Script responsible for setting up port forwarding for Istio sidecar.

Usage:
  pilot-agent istio-iptables [flags]

Flags:
  -n, --dry-run                                     Do not call any external dependencies like iptables
  -p, --envoy-port string                           Specify the envoy port to which redirect all TCP traffic (default $ENVOY_PORT = 15001)
  -h, --help                                        help for istio-iptables
  -z, --inbound-capture-port string                 Port to which all inbound TCP traffic to the pod/VM should be redirected to (default $INBOUND_CAPTURE_PORT = 15006)
  -e, --inbound-tunnel-port string                  Specify the istio tunnel port for inbound tcp traffic (default $INBOUND_TUNNEL_PORT = 15008)
      --iptables-probe-port string                  set listen port for failure detection (default "15002")
  -m, --istio-inbound-interception-mode string      The mode used to redirect inbound connections to Envoy, either "REDIRECT" or "TPROXY"
  -b, --istio-inbound-ports string                  Comma separated list of inbound ports for which traffic is to be redirected to Envoy (optional). The wildcard character "*" can be used to configure redirection for all ports. An empty list will disable
  -t, --istio-inbound-tproxy-mark string            
  -r, --istio-inbound-tproxy-route-table string     
  -d, --istio-local-exclude-ports string            Comma separated list of inbound ports to be excluded from redirection to Envoy (optional). Only applies  when all inbound traffic (i.e. "*") is being redirected (default to $ISTIO_LOCAL_EXCLUDE_PORTS)
  -o, --istio-local-outbound-ports-exclude string   Comma separated list of outbound ports to be excluded from redirection to Envoy
  -q, --istio-outbound-ports string                 Comma separated list of outbound ports to be explicitly included for redirection to Envoy
  -i, --istio-service-cidr string                   Comma separated list of IP ranges in CIDR form to redirect to envoy (optional). The wildcard character "*" can be used to redirect all outbound traffic. An empty list will disable all outbound
  -x, --istio-service-exclude-cidr string           Comma separated list of IP ranges in CIDR form to be excluded from redirection. Only applies when all  outbound traffic (i.e. "*") is being redirected (default to $ISTIO_SERVICE_EXCLUDE_CIDR)
  -k, --kube-virt-interfaces string                 Comma separated list of virtual interfaces whose inbound traffic (from VM) will be treated as outbound
      --probe-timeout duration                      failure detection timeout (default 5s)
  -g, --proxy-gid string                            Specify the GID of the user for which the redirection is not applied. (same default value as -u param)
  -u, --proxy-uid string                            Specify the UID of the user for which the redirection is not applied. Typically, this is the UID of the proxy container
  -f, --restore-format                              Print iptables rules in iptables-restore interpretable format (default true)
      --run-validation                              Validate iptables
      --skip-rule-apply                             Skip iptables apply
```

下文初始化命令结合上命令行工具说明，常规的init-proxy。可以得出该规则的目的是为了达成"将所有非uid 1337用户，除15090，15021，15020端口外的所有流量重定向到15006端口，由15001端口负责重定向"，这里之所以排除了15090,15021,15020三个端口是因为这三个端口分别用于envoy prometheus监控、pilot-proxy健康检查、应用和envoy以及pilot-agent和应用的prometheus合并监控遥测。

```
istio-iptables 
	  -p 15001				## 由15001端口负责重定向流量
      -z 15006      		## inbuond流量转发到15006端口
      -u 1337       		## UID为1337的用户不需要执行流量重定向
      -m REDIRECT   		## 流量重定向模式采用REDIRECT，意味着udp转发可能存在问题
      -i *					## 所有ip都需要被重定向
      -x  					## 所有ip都需要被重定向
      -b *					## 所有端口都需要被重定向
      -d 15090,15021,15020	## 15090，15021，15020端口不需要重定向
```

在实际的iptables规则中入口流量规则如下，满足了上文对入口流量的排除

```shell
Chain PREROUTING (policy ACCEPT 35509 packets, 2130540 bytes)
    pkts      bytes target     prot opt in     out     source               destination         
   35510  2130580 ISTIO_INBOUND  tcp  --  *      *       0.0.0.0/0            0.0.0.0/0
   
Chain ISTIO_IN_REDIRECT (3 references)
    pkts      bytes target     prot opt in     out     source               destination         
       1       40 REDIRECT   tcp  --  *      *       0.0.0.0/0            0.0.0.0/0            redir ports 15006
   
Chain ISTIO_INBOUND (1 references)
    pkts      bytes target     prot opt in     out     source               destination         
       0        0 RETURN     tcp  --  *      *       0.0.0.0/0            0.0.0.0/0            tcp dpt:15008
       0        0 RETURN     tcp  --  *      *       0.0.0.0/0            0.0.0.0/0            tcp dpt:22
       0        0 RETURN     tcp  --  *      *       0.0.0.0/0            0.0.0.0/0            tcp dpt:15090
   35508  2130480 RETURN     tcp  --  *      *       0.0.0.0/0            0.0.0.0/0            tcp dpt:15021
       1       60 RETURN     tcp  --  *      *       0.0.0.0/0            0.0.0.0/0            tcp dpt:15020
       1       40 ISTIO_IN_REDIRECT  tcp  --  *      *       0.0.0.0/0            0.0.0.0/0 
```

出口流量如下，`ISTIO_OUTPUT`链第一调规则使得源自127.0.0.6发往回环网卡的数据包直接通过。第二条将发往回环网卡且目的地址非127.0.0.1的istio-proxy流量进行重定向。第三条允许普通的回环地址的流量直接通过。第四条结合第二条则是允许所有proxy流量从lo的127.0.0.1或其他网卡的流量通过。后8条与前四条类似，不过判断依据为流量的GID。

第二条规则的存在的原因是为了提供envoy到自身envoy的数据流量管理。第一条规则使用于envoy对upstream的访问。第三条规则允许非envoy的回环访问不需要被envoy重定向。第四条则是为了开放envoy的对外流量。如果是应用的对外流量则会触发该链的最后一条被重定向到envoy。

```shell
Chain OUTPUT (policy ACCEPT 15104 packets, 1351718 bytes)
    pkts      bytes target     prot opt in     out     source               destination         
     234    14119 ISTIO_OUTPUT  tcp  --  *      *       0.0.0.0/0            0.0.0.0/0 

Chain ISTIO_IN_REDIRECT (3 references)
    pkts      bytes target     prot opt in     out     source               destination         
       1       40 REDIRECT   tcp  --  *      *       0.0.0.0/0            0.0.0.0/0            redir ports 15006

Chain ISTIO_OUTPUT (1 references)
    pkts      bytes target     prot opt in     out     source               destination         
       0        0 RETURN     all  --  *      lo      127.0.0.6            0.0.0.0/0           
       0        0 ISTIO_IN_REDIRECT  all  --  *      lo      0.0.0.0/0           !127.0.0.1            owner UID match 1337
       4      240 RETURN     all  --  *      lo      0.0.0.0/0            0.0.0.0/0            ! owner UID match 1337
     172    10353 RETURN     all  --  *      *       0.0.0.0/0            0.0.0.0/0            owner UID match 1337
       0        0 ISTIO_IN_REDIRECT  all  --  *      lo      0.0.0.0/0           !127.0.0.1            owner GID match 1337
       0        0 RETURN     all  --  *      lo      0.0.0.0/0            0.0.0.0/0            ! owner GID match 1337
       0        0 RETURN     all  --  *      *       0.0.0.0/0            0.0.0.0/0            owner GID match 1337
       0        0 RETURN     all  --  *      *       0.0.0.0/0            127.0.0.1           
      58     3526 ISTIO_REDIRECT  all  --  *      *       0.0.0.0/0            0.0.0.0/0
```





# 二、源码

该代码显式的完成了`output`规则链相关的iptables规则生成，通过函数`handleInboundIpv4Rules`和`handleInboundIpv6Rules`完成了`input`链规则生成，最终通过`iptConfigurator.executeCommands()`完成规则的应用。

有趣的是在`iptConfigurator`创建时其成员变量`iptConfigurator.ext`代表了iptables规则所依赖的平台，将平台依赖抽象为一层，有利于未来扩展，但未来扩展时是否还能使用类似的iptables规则或者通过转换兼容。因为不同的工具最终的规则可能完全不同。

```go
// /root/src/go/src/github.com/istio/tools/istio-iptables/pkg/cmd/run.go

func (iptConfigurator *IptablesConfigurator) run() {
	defer func() {
		// Best effort since we don't know if the commands exist
		_ = iptConfigurator.ext.Run(constants.IPTABLESSAVE)
		if iptConfigurator.cfg.EnableInboundIPv6 {
			_ = iptConfigurator.ext.Run(constants.IP6TABLESSAVE)
		}
	}()

	// Since OUTBOUND_IP_RANGES_EXCLUDE could carry ipv4 and ipv6 ranges
	// need to split them in different arrays one for ipv4 and one for ipv6
	// in order to not to fail
	ipv4RangesExclude, ipv6RangesExclude, err := iptConfigurator.separateV4V6(iptConfigurator.cfg.OutboundIPRangesExclude)
	if err != nil {
		panic(err)
	}
	if ipv4RangesExclude.IsWildcard {
		panic("Invalid value for OUTBOUND_IP_RANGES_EXCLUDE")
	}
	// FixMe: Do we need similar check for ipv6RangesExclude as well ??

	ipv4RangesInclude, ipv6RangesInclude, err := iptConfigurator.separateV4V6(iptConfigurator.cfg.OutboundIPRangesInclude)
	if err != nil {
		panic(err)
	}

	redirectDNS := false
	if dnsCaptureByAgent {
		redirectDNS = true
	}
	iptConfigurator.logConfig()

	if iptConfigurator.cfg.EnableInboundIPv6 {
		//TODO: (abhide): Move this out of this method
		iptConfigurator.ext.RunOrFail(constants.IP, "-6", "addr", "add", "::6/128", "dev", "lo")
	}

	// Do not capture internal interface.
	iptConfigurator.shortCircuitKubeInternalInterface()

	// Create a new chain for to hit tunnel port directly. Envoy will be listening on port acting as VPN tunnel.
	iptConfigurator.iptables.AppendRuleV4(constants.ISTIOINBOUND, constants.NAT, "-p", constants.TCP, "--dport",
		iptConfigurator.cfg.InboundTunnelPort, "-j", constants.RETURN)

	if redirectDNS {
		// redirect all TCP dns traffic on port 53 to the agent on port 15053
		iptConfigurator.iptables.AppendRuleV4(
			constants.ISTIOREDIRECT, constants.NAT, "-p", constants.TCP, "--dport", "53", "-j", constants.REDIRECT, "--to-ports", constants.IstioAgentDNSListenerPort)
		// the rest of the IPtables rules will take care of ensuring that the traffic does not loop, among other things.
	}

	// Create a new chain for redirecting outbound traffic to the common Envoy port.
	// In both chains, '-j RETURN' bypasses Envoy and '-j ISTIOREDIRECT'
	// redirects to Envoy.
	iptConfigurator.iptables.AppendRuleV4(
		constants.ISTIOREDIRECT, constants.NAT, "-p", constants.TCP, "-j", constants.REDIRECT, "--to-ports", iptConfigurator.cfg.ProxyPort)

	// Use this chain also for redirecting inbound traffic to the common Envoy port
	// when not using TPROXY.

	iptConfigurator.iptables.AppendRuleV4(constants.ISTIOINREDIRECT, constants.NAT, "-p", constants.TCP, "-j", constants.REDIRECT,
		"--to-ports", iptConfigurator.cfg.InboundCapturePort)

	iptConfigurator.handleInboundPortsInclude()

	// TODO: change the default behavior to not intercept any output - user may use http_proxy or another
	// iptablesOrFail wrapper (like ufw). Current default is similar with 0.1
	// Jump to the ISTIOOUTPUT chain from OUTPUT chain for all tcp traffic, and UDP dns (if enabled)
	iptConfigurator.iptables.AppendRuleV4(constants.OUTPUT, constants.NAT, "-p", constants.TCP, "-j", constants.ISTIOOUTPUT)
	// Apply port based exclusions. Must be applied before connections back to self are redirected.
	if iptConfigurator.cfg.OutboundPortsExclude != "" {
		for _, port := range split(iptConfigurator.cfg.OutboundPortsExclude) {
			iptConfigurator.iptables.AppendRuleV4(constants.ISTIOOUTPUT, constants.NAT, "-p", constants.TCP, "--dport", port, "-j", constants.RETURN)
		}
	}

	// 127.0.0.6 is bind connect from inbound passthrough cluster
	iptConfigurator.iptables.AppendRuleV4(constants.ISTIOOUTPUT, constants.NAT, "-o", "lo", "-s", "127.0.0.6/32", "-j", constants.RETURN)

	for _, uid := range split(iptConfigurator.cfg.ProxyUID) {
		// Redirect app calls back to itself via Envoy when using the service VIP
		// e.g. appN => Envoy (client) => Envoy (server) => appN.
		// nolint: lll
		iptConfigurator.iptables.AppendRuleV4(constants.ISTIOOUTPUT, constants.NAT, "-o", "lo", "!", "-d", "127.0.0.1/32", "-m", "owner", "--uid-owner", uid, "-j", constants.ISTIOINREDIRECT)

		// Do not redirect app calls to back itself via Envoy when using the endpoint address
		// e.g. appN => appN by lo
		// If loopback explicitly set via OutboundIPRangesInclude, then don't return.
		if !ipv4RangesInclude.HasLoopBackIP && !ipv6RangesInclude.HasLoopBackIP {
			iptConfigurator.iptables.AppendRuleV4(constants.ISTIOOUTPUT, constants.NAT, "-o", "lo", "-m", "owner", "!", "--uid-owner", uid, "-j", constants.RETURN)
		}

		// Avoid infinite loops. Don't redirect Envoy traffic directly back to
		// Envoy for non-loopback traffic.
		iptConfigurator.iptables.AppendRuleV4(constants.ISTIOOUTPUT, constants.NAT, "-m", "owner", "--uid-owner", uid, "-j", constants.RETURN)
	}

	for _, gid := range split(iptConfigurator.cfg.ProxyGID) {
		// Redirect app calls back to itself via Envoy when using the service VIP
		// e.g. appN => Envoy (client) => Envoy (server) => appN.
		// nolint: lll
		iptConfigurator.iptables.AppendRuleV4(constants.ISTIOOUTPUT, constants.NAT, "-o", "lo", "!", "-d", "127.0.0.1/32", "-m", "owner", "--gid-owner", gid, "-j", constants.ISTIOINREDIRECT)

		// Do not redirect app calls to back itself via Envoy when using the endpoint address
		// e.g. appN => appN by lo
		// If loopback explicitly set via OutboundIPRangesInclude, then don't return.
		if !ipv4RangesInclude.HasLoopBackIP && !ipv6RangesInclude.HasLoopBackIP {
			iptConfigurator.iptables.AppendRuleV4(constants.ISTIOOUTPUT, constants.NAT, "-o", "lo", "-m", "owner", "!", "--gid-owner", gid, "-j", constants.RETURN)
		}

		// Avoid infinite loops. Don't redirect Envoy traffic directly back to
		// Envoy for non-loopback traffic.
		iptConfigurator.iptables.AppendRuleV4(constants.ISTIOOUTPUT, constants.NAT, "-m", "owner", "--gid-owner", gid, "-j", constants.RETURN)
	}
	// Skip redirection for Envoy-aware applications and
	// container-to-container traffic both of which explicitly use
	// localhost.
	iptConfigurator.iptables.AppendRuleV4(constants.ISTIOOUTPUT, constants.NAT, "-d", "127.0.0.1/32", "-j", constants.RETURN)
	// Apply outbound IPv4 exclusions. Must be applied before inclusions.
	for _, cidr := range ipv4RangesExclude.IPNets {
		iptConfigurator.iptables.AppendRuleV4(constants.ISTIOOUTPUT, constants.NAT, "-d", cidr.String(), "-j", constants.RETURN)
	}

	iptConfigurator.handleOutboundPortsInclude()

	iptConfigurator.handleInboundIpv4Rules(ipv4RangesInclude)
	if iptConfigurator.cfg.EnableInboundIPv6 {
		iptConfigurator.handleInboundIpv6Rules(ipv6RangesExclude, ipv6RangesInclude)
	}

	if redirectDNS {
		// Make sure that upstream DNS requests from agent/envoy dont get captured.
		for _, uid := range split(iptConfigurator.cfg.ProxyUID) {
			iptConfigurator.iptables.AppendRuleV4(constants.OUTPUT, constants.NAT,
				"-p", "udp", "--dport", "53", "-m", "owner", "--uid-owner", uid, "-j", constants.RETURN)
		}
		for _, gid := range split(iptConfigurator.cfg.ProxyGID) {
			// TODO: add ip6 as well
			iptConfigurator.iptables.AppendRuleV4(constants.OUTPUT, constants.NAT,
				"-p", "udp", "--dport", "53", "-m", "owner", "--gid-owner", gid, "-j", constants.RETURN)
		}

		// from app to agent/envoy - dnat to 127.0.0.1:port
		iptConfigurator.iptables.AppendRuleV4(constants.OUTPUT, constants.NAT,
			"-p", "udp", "--dport", "53",
			"-j", "DNAT", "--to-destination", "127.0.0.1:"+constants.IstioAgentDNSListenerPort)
		// overwrite the source IP so that when envoy/agent responds to the DNS request
		// it responds to localhost on same interface. Otherwise, the connection will not
		// match in the kernel. Note that the dest port here should be the rewritten port.
		iptConfigurator.iptables.AppendRuleV4(constants.POSTROUTING, constants.NAT,
			"-p", "udp", "--dport", constants.IstioAgentDNSListenerPort, "-j", "SNAT", "--to-source", "127.0.0.1")
	}

	if iptConfigurator.cfg.InboundInterceptionMode == constants.TPROXY {
		// mark outgoing packets from 127.0.0.1/32 with 1337, match it to policy routing entry setup for TPROXY mode
		iptConfigurator.iptables.AppendRuleV4(constants.OUTPUT, constants.MANGLE,
			"-p", constants.TCP, "-s", "127.0.0.1/32", "!", "-d", "127.0.0.1/32",
			"-j", constants.MARK, "--set-mark", iptConfigurator.cfg.InboundTProxyMark)
	}
	iptConfigurator.executeCommands()
}
```

