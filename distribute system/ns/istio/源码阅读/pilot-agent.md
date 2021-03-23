[toc]



# 一、配置项

## role::Proxy配置

istio以`Proxy`结构体表示代理的类型和相关信息

```go
// Proxy contains information about an specific instance of a proxy (envoy sidecar, gateway,
// etc). The Proxy is initialized when a sidecar connects to Pilot, and populated from
// 'node' info in the protocol as well as data extracted from registries.
//
// In current Istio implementation nodes use a 4-parts '~' delimited ID.
// Type~IPAddress~ID~Domain
type Proxy struct {
	sync.RWMutex

	// Type specifies the node type. First part of the ID.
	Type NodeType

	// IPAddresses is the IP addresses of the proxy used to identify it and its
	// co-located service instances. Example: "10.60.1.6". In some cases, the host
	// where the poxy and service instances reside may have more than one IP address
	IPAddresses []string

	// ID is the unique platform-specific sidecar proxy ID. For k8s it is the pod ID and
	// namespace.
	ID string

	// Locality is the location of where Envoy proxy runs. This is extracted from
	// the registry where possible. If the registry doesn't provide a locality for the
	// proxy it will use the one sent via ADS that can be configured in the Envoy bootstrap
	Locality *core.Locality

	// DNSDomain defines the DNS domain suffix for short hostnames (e.g.
	// "default.svc.cluster.local")
	DNSDomain string

	// ConfigNamespace defines the namespace where this proxy resides
	// for the purposes of network scoping.
	// NOTE: DO NOT USE THIS FIELD TO CONSTRUCT DNS NAMES
	ConfigNamespace string

	// Metadata key-value pairs extending the Node identifier
	Metadata *NodeMetadata

	// the sidecarScope associated with the proxy
	SidecarScope *SidecarScope

	// the sidecarScope associated with the proxy previously
	PrevSidecarScope *SidecarScope

	// The merged gateways associated with the proxy if this is a Router
	MergedGateway *MergedGateway

	// service instances associated with the proxy
	ServiceInstances []*ServiceInstance

	// Istio version associated with the Proxy
	IstioVersion *IstioVersion

	// VerifiedIdentity determines whether a proxy had its identity verified. This
	// generally occurs by JWT or mTLS authentication. This can be false when
	// connecting over plaintext. If this is set to true, we can verify the proxy has
	// access to ConfigNamespace namespace. However, other options such as node type
	// are not part of an Istio identity and thus are not verified.
	VerifiedIdentity *spiffe.Identity

	// Indicates whether proxy supports IPv6 addresses
	ipv6Support bool

	// Indicates whether proxy supports IPv4 addresses
	ipv4Support bool

	// GlobalUnicastIP stores the global unicast IP if available, otherwise nil
	GlobalUnicastIP string

	// XdsResourceGenerator is used to generate resources for the node, based on the PushContext.
	// If nil, the default networking/core v2 generator is used. This field can be set
	// at connect time, based on node metadata, to trigger generation of a different style
	// of configuration.
	XdsResourceGenerator XdsResourceGenerator

	// WatchedResources contains the list of watched resources for the proxy, keyed by the DiscoveryRequest TypeUrl.
	WatchedResources map[string]*WatchedResource
}
```

在pilot-agent的启动中以变量`role`作为`Proxy`对象实例，并且以sidecar形式与应用绑定的服务，因此启动代码中显式的设置默认值`role.Type = model.SidecarProxy`或命令行的第一个参数。其他`role`变量配置代码摘要如下

```go
// No IP addresses provided, append 127.0.0.1 for ipv4 and ::1 for ipv6
if len(role.IPAddresses) == 0 {
    role.IPAddresses = append(role.IPAddresses, "127.0.0.1")
    role.IPAddresses = append(role.IPAddresses, "::1")
}
role.ID = podName + "." + podNamespace
role.DNSDomain = getDNSDomain(podNamespace, role.DNSDomain)
```

### 涉及环境变量

| 环境变量      | 说明        | 默认值 |
| ------------- | ----------- | ------ |
| INSTANCE_IP   | 实例标识IP  |        |
| POD_NAMESPACE | POD命名空间 |        |

### 涉及范围



## proxyConfig::meshconfig.ProxyConfig配置

`proxyConfig`在istio-proxy中可以用于启动agent和envoy，而`proxyConfig`由`func constructProxyConfig() (meshconfig.ProxyConfig, error)`决定

```go
// ProxyConfig defines variables for individual Envoy instances.
type ProxyConfig struct {
	// Path to the generated configuration file directory.
	// Proxy agent generates the actual configuration and stores it in this directory.
	ConfigPath string `protobuf:"bytes,1,opt,name=config_path,json=configPath,proto3" json:"configPath,omitempty"`
	// Path to the proxy binary
	BinaryPath string `protobuf:"bytes,2,opt,name=binary_path,json=binaryPath,proto3" json:"binaryPath,omitempty"`
	// Service cluster defines the name for the `service_cluster` that is
	// shared by all Envoy instances. This setting corresponds to
	// `--service-cluster` flag in Envoy.  In a typical Envoy deployment, the
	// `service-cluster` flag is used to identify the caller, for
	// source-based routing scenarios.
	//
	// Since Istio does not assign a local `service/service` version to each
	// Envoy instance, the name is same for all of them.  However, the
	// source/caller's identity (e.g., IP address) is encoded in the
	// `--service-node` flag when launching Envoy.  When the RDS service
	// receives API calls from Envoy, it uses the value of the `service-node`
	// flag to compute routes that are relative to the service instances
	// located at that IP address.
	ServiceCluster string `protobuf:"bytes,3,opt,name=service_cluster,json=serviceCluster,proto3" json:"serviceCluster,omitempty"`
	// The time in seconds that Envoy will drain connections during a hot
	// restart. MUST be >=1s (e.g., _1s/1m/1h_)
	// Default drain duration is `45s`.
	DrainDuration *types.Duration `protobuf:"bytes,4,opt,name=drain_duration,json=drainDuration,proto3" json:"drainDuration,omitempty"`
	// The time in seconds that Envoy will wait before shutting down the
	// parent process during a hot restart. MUST be >=1s (e.g., `1s/1m/1h`).
	// MUST BE greater than `drain_duration` parameter.
	// Default shutdown duration is `60s`.
	ParentShutdownDuration *types.Duration `protobuf:"bytes,5,opt,name=parent_shutdown_duration,json=parentShutdownDuration,proto3" json:"parentShutdownDuration,omitempty"`
	// Address of the discovery service exposing xDS with mTLS connection.
	// The inject configuration may override this value.
	DiscoveryAddress string `protobuf:"bytes,6,opt,name=discovery_address,json=discoveryAddress,proto3" json:"discoveryAddress,omitempty"`
	// $hide_from_docs
	DiscoveryRefreshDelay *types.Duration `protobuf:"bytes,7,opt,name=discovery_refresh_delay,json=discoveryRefreshDelay,proto3" json:"discoveryRefreshDelay,omitempty"` // Deprecated: Do not use.
	// Address of the Zipkin service (e.g. _zipkin:9411_).
	// DEPRECATED: Use [tracing][istio.mesh.v1alpha1.ProxyConfig.tracing] instead.
	ZipkinAddress string `protobuf:"bytes,8,opt,name=zipkin_address,json=zipkinAddress,proto3" json:"zipkinAddress,omitempty"` // Deprecated: Do not use.
	// IP Address and Port of a statsd UDP listener (e.g. `10.75.241.127:9125`).
	StatsdUdpAddress string `protobuf:"bytes,10,opt,name=statsd_udp_address,json=statsdUdpAddress,proto3" json:"statsdUdpAddress,omitempty"`
	// $hide_from_docs
	EnvoyMetricsServiceAddress string `protobuf:"bytes,20,opt,name=envoy_metrics_service_address,json=envoyMetricsServiceAddress,proto3" json:"envoyMetricsServiceAddress,omitempty"` // Deprecated: Do not use.
	// Port on which Envoy should listen for administrative commands.
	// Default port is `15000`.
	ProxyAdminPort int32 `protobuf:"varint,11,opt,name=proxy_admin_port,json=proxyAdminPort,proto3" json:"proxyAdminPort,omitempty"`
	// $hide_from_docs
	AvailabilityZone string `protobuf:"bytes,12,opt,name=availability_zone,json=availabilityZone,proto3" json:"availabilityZone,omitempty"` // Deprecated: Do not use.
	// AuthenticationPolicy defines how the proxy is authenticated when it connects to the control plane.
	// Default is set to `MUTUAL_TLS`.
	ControlPlaneAuthPolicy AuthenticationPolicy `protobuf:"varint,13,opt,name=control_plane_auth_policy,json=controlPlaneAuthPolicy,proto3,enum=istio.mesh.v1alpha1.AuthenticationPolicy" json:"controlPlaneAuthPolicy,omitempty"`
	// File path of custom proxy configuration, currently used by proxies
	// in front of Mixer and Pilot.
	CustomConfigFile string `protobuf:"bytes,14,opt,name=custom_config_file,json=customConfigFile,proto3" json:"customConfigFile,omitempty"`
	// Maximum length of name field in Envoy's metrics. The length of the name field
	// is determined by the length of a name field in a service and the set of labels that
	// comprise a particular version of the service. The default value is set to 189 characters.
	// Envoy's internal metrics take up 67 characters, for a total of 256 character name per metric.
	// Increase the value of this field if you find that the metrics from Envoys are truncated.
	StatNameLength int32 `protobuf:"varint,15,opt,name=stat_name_length,json=statNameLength,proto3" json:"statNameLength,omitempty"`
	// The number of worker threads to run.
	// If unset, this will be automatically determined based on CPU requests/limits.
	// If set to 0, all cores on the machine will be used.
	// Default is 2 worker threads.
	Concurrency *types.Int32Value `protobuf:"bytes,16,opt,name=concurrency,proto3" json:"concurrency,omitempty"`
	// Path to the proxy bootstrap template file
	ProxyBootstrapTemplatePath string `protobuf:"bytes,17,opt,name=proxy_bootstrap_template_path,json=proxyBootstrapTemplatePath,proto3" json:"proxyBootstrapTemplatePath,omitempty"`
	// The mode used to redirect inbound traffic to Envoy.
	InterceptionMode ProxyConfig_InboundInterceptionMode `protobuf:"varint,18,opt,name=interception_mode,json=interceptionMode,proto3,enum=istio.mesh.v1alpha1.ProxyConfig_InboundInterceptionMode" json:"interceptionMode,omitempty"`
	// Tracing configuration to be used by the proxy.
	Tracing *Tracing `protobuf:"bytes,19,opt,name=tracing,proto3" json:"tracing,omitempty"`
	// Secret Discovery Service(SDS) configuration to be used by the proxy.
	Sds *SDS `protobuf:"bytes,21,opt,name=sds,proto3" json:"sds,omitempty"`
	// Address of the service to which access logs from Envoys should be
	// sent. (e.g. `accesslog-service:15000`). See [Access Log
	// Service](https://www.envoyproxy.io/docs/envoy/latest/api-v2/config/accesslog/v2/als.proto)
	// for details about Envoy's gRPC Access Log Service API.
	EnvoyAccessLogService *RemoteService `protobuf:"bytes,22,opt,name=envoy_access_log_service,json=envoyAccessLogService,proto3" json:"envoyAccessLogService,omitempty"`
	// Address of the Envoy Metrics Service implementation (e.g. `metrics-service:15000`).
	// See [Metric Service](https://www.envoyproxy.io/docs/envoy/latest/api-v2/config/metrics/v2/metrics_service.proto)
	// for details about Envoy's Metrics Service API.
	EnvoyMetricsService *RemoteService `protobuf:"bytes,23,opt,name=envoy_metrics_service,json=envoyMetricsService,proto3" json:"envoyMetricsService,omitempty"`
	// $hide_from_docs
	// Additional env variables for the proxy.
	// Names starting with ISTIO_META_ will be included in the generated bootstrap and sent to the XDS server.
	ProxyMetadata map[string]string `protobuf:"bytes,24,rep,name=proxy_metadata,json=proxyMetadata,proto3" json:"proxyMetadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Port on which the agent should listen for administrative commands such as readiness probe.
	// Default is set to port `15020`.
	StatusPort int32 `protobuf:"varint,26,opt,name=status_port,json=statusPort,proto3" json:"statusPort,omitempty"`
	// An additional list of tags to extract from the in-proxy Istio telemetry. These extra tags can be
	// added by configuring the telemetry extension. Each additional tag needs to be present in this list.
	// Extra tags emitted by the telemetry extensions must be listed here so that they can be processed
	// and exposed as Prometheus metrics.
	ExtraStatTags []string `protobuf:"bytes,27,rep,name=extra_stat_tags,json=extraStatTags,proto3" json:"extraStatTags,omitempty"`
	// $hide_from_docs
	// Topology encapsulates the configuration which describes where the proxy is
	// located i.e. behind a (or N) trusted proxy (proxies) or directly exposed
	// to the internet. This configuration only effects gateways and is applied
	// to all the gateways in the cluster unless overriden via annotations of the
	// gateway workloads.
	GatewayTopology *Topology `protobuf:"bytes,28,opt,name=gateway_topology,json=gatewayTopology,proto3" json:"gatewayTopology,omitempty"`
	// The amount of time allowed for connections to complete on proxy shutdown.
	// On receiving `SIGTERM` or `SIGINT`, `istio-agent` tells the active Envoy to start draining,
	// preventing any new connections and allowing existing connections to complete. It then
	// sleeps for the `termination_drain_duration` and then kills any remaining active Envoy processes.
	// If not set, a default of `5s` will be applied.
	TerminationDrainDuration *types.Duration `protobuf:"bytes,29,opt,name=termination_drain_duration,json=terminationDrainDuration,proto3" json:"terminationDrainDuration,omitempty"`
	// The unique identifier for the [service mesh](https://istio.io/docs/reference/glossary/#service-mesh)
	// All control planes running in the same service mesh should specify the same mesh ID.
	// Mesh ID is used to label telemetry reports for cases where telemetry from multiple meshes is mixed together.
	MeshId string `protobuf:"bytes,30,opt,name=mesh_id,json=meshId,proto3" json:"meshId,omitempty"`
	// VM Health Checking readiness probe. This health check config exactly mirrors the
	// kubernetes readiness probe configuration both in schema and logic.
	// Only one health check method of 3 can be set at a time.
	ReadinessProbe *v1alpha3.ReadinessProbe `protobuf:"bytes,31,opt,name=readiness_probe,json=readinessProbe,proto3" json:"readinessProbe,omitempty"`
	// Proxy stats matcher defines configuration for reporting custom Envoy stats.
	// To reduce memory and CPU overhead from Envoy stats system, Istio proxies by
	// default create and expose only a subset of Envoy stats. This option is to
	// control creation of additional Envoy stats with prefix, suffix, and regex
	// expressions match on the name of the stats. This replaces the stats
	// inclusion annotations
	// (`sidecar.istio.io/statsInclusionPrefixes`,
	// `sidecar.istio.io/statsInclusionRegexps`, and
	// `sidecar.istio.io/statsInclusionSuffixes`). For example, to enable stats
	// for circuit breaker, retry, and upstream connections, you can specify stats
	// matcher as follow:
	// ```yaml
	// proxyStatsMatcher:
	//   inclusionRegexps:
	//     - .*circuit_breakers.*
	//   inclusionPrefixes:
	//     - upstream_rq_retry
	//     - upstream_cx
	// ```
	// Note including more Envoy stats might increase number of time series
	// collected by prometheus significantly. Care needs to be taken on Prometheus
	// resource provision and configuration to reduce cardinality.
	ProxyStatsMatcher *ProxyConfig_ProxyStatsMatcher `protobuf:"bytes,32,opt,name=proxy_stats_matcher,json=proxyStatsMatcher,proto3" json:"proxyStatsMatcher,omitempty"`
	// Boolean flag for enabling/disabling the holdApplicationUntilProxyStarts behavior.
	// This feature adds hooks to delay application startup until the pod proxy
	// is ready to accept traffic, mitigating some startup race conditions.
	// Default value is 'false'.
	HoldApplicationUntilProxyStarts *types.BoolValue `protobuf:"bytes,33,opt,name=hold_application_until_proxy_starts,json=holdApplicationUntilProxyStarts,proto3" json:"holdApplicationUntilProxyStarts,omitempty"`
	XXX_NoUnkeyedLiteral            struct{}         `json:"-"`
	XXX_unrecognized                []byte           `json:"-"`
	XXX_sizecache                   int32            `json:"-"`
}
```



`ProxyConfig`可以有三种来源，第一种是默认值，第二种是`./etc/istio/config/mesh`中的`meshconfig.MeshConfig.DefaultConfig`的，第三种是命令行flag，第四种为annotations中指定，三种值优先级逐渐变高，高优先级覆盖低优先级。

代码中`readPodAnnotations()`会从`"./etc/istio/pod/annotations"`路径中读取annotations信息，由于这个值为常量，因此如果非容器启动则应该在工作目录准备好这些配置。在k8s中istio注入会使用k8s的`downloadApi`将pod的annotations和label注入到对应的文件中。

`meshConfigFile`从命令行`--meshConfig`中获取，默认值为`./etc/istio/config/mesh`用于读取`meshconfig.MeshConfig`。该配置会覆盖默认`MeshConfig`,并被annotations所指定的config覆盖。最简化的sidecar注入`--meshConfig`和annotations不指定，所以默认值为`func DefaultMeshConfig() meshconfig.MeshConfig`生成的值。



```go
// /root/src/go/src/github.com/istio/pilot/cmd/pilot-agent/config.go

// return proxyConfig and trustDomain
func constructProxyConfig() (meshconfig.ProxyConfig, error) {
	annotations, err := readPodAnnotations()
	if err != nil {
		log.Warnf("failed to read pod annotations: %v", err)
	}
	var fileMeshContents string
	if fileExists(meshConfigFile) {
		contents, err := ioutil.ReadFile(meshConfigFile)
		if err != nil {
			return meshconfig.ProxyConfig{}, fmt.Errorf("failed to read mesh config file %v: %v", meshConfigFile, err)
		}
		fileMeshContents = string(contents)
	}
	meshConfig, err := getMeshConfig(fileMeshContents, annotations[annotation.ProxyConfig.Name])
	if err != nil {
		return meshconfig.ProxyConfig{}, err
	}
	proxyConfig := mesh.DefaultProxyConfig()
	if meshConfig.DefaultConfig != nil {
		proxyConfig = *meshConfig.DefaultConfig
	}

	proxyConfig.Concurrency = &types.Int32Value{Value: int32(concurrency)}
	proxyConfig.ServiceCluster = serviceCluster
	// resolve statsd address
	if proxyConfig.StatsdUdpAddress != "" {
		addr, err := network.ResolveAddr(proxyConfig.StatsdUdpAddress)
		if err != nil {
			log.Warnf("resolve StatsdUdpAddress failed: %v", err)
			proxyConfig.StatsdUdpAddress = ""
		} else {
			proxyConfig.StatsdUdpAddress = addr
		}
	}
	if err := validation.ValidateProxyConfig(&proxyConfig); err != nil {
		return meshconfig.ProxyConfig{}, err
	}
	return applyAnnotations(proxyConfig, annotations), nil
}


// /root/src/go/src/github.com/istio/pkg/config/mesh/mesh.go
// DefaultProxyConfig for individual proxies
func DefaultProxyConfig() meshconfig.ProxyConfig {
	// TODO: include revision based on REVISION env
	// TODO: set default namespace based on POD_NAMESPACE env
	return meshconfig.ProxyConfig{
		// missing: ConnectTimeout: 10 * time.Second,
		ConfigPath:               constants.ConfigPathDir,
		ServiceCluster:           constants.ServiceClusterName,
		DrainDuration:            types.DurationProto(45 * time.Second),
		ParentShutdownDuration:   types.DurationProto(60 * time.Second),
		TerminationDrainDuration: types.DurationProto(5 * time.Second),
		ProxyAdminPort:           15000,
		Concurrency:              &types.Int32Value{Value: 2},
		ControlPlaneAuthPolicy:   meshconfig.AuthenticationPolicy_MUTUAL_TLS,
		DiscoveryAddress:         "istiod.istio-system.svc:15012",
		Tracing: &meshconfig.Tracing{
			Tracer: &meshconfig.Tracing_Zipkin_{
				Zipkin: &meshconfig.Tracing_Zipkin{
					Address: "zipkin.istio-system:9411",
				},
			},
		},

		// Code defaults
		BinaryPath:            constants.BinaryPathFilename,
		StatsdUdpAddress:      "",
		EnvoyMetricsService:   &meshconfig.RemoteService{Address: ""},
		EnvoyAccessLogService: &meshconfig.RemoteService{Address: ""},
		CustomConfigFile:      "",
		StatNameLength:        189,
		StatusPort:            15020,
	}
}
```

### 涉及范围

### 涉及命令行flag

| flag名称       | 描述                                                         | 默认值                    |
| -------------- | ------------------------------------------------------------ | ------------------------- |
| meshConfig     | `meshconfig.MeshConfig`配置文件路径                          | `./etc/istio/config/mesh` |
| concurrency    | 并发数量                                                     | 0                         |
| serviceCluster | xDS调用时使用的服务名，在sidecar注入模式下为<pod>.<namspace> | istio-proxy               |



## secOpts::security.Options

```go
// Options provides all of the configuration parameters for secret discovery service
// and CA configuration. Used in both Istiod and Agent.
// TODO: ProxyConfig should have most of those, and be passed to all components
// (as source of truth)
type Options struct {
	// WorkloadUDSPath is the unix domain socket through which SDS server communicates with workload proxies.
	WorkloadUDSPath string

	// IngressGatewayUDSPath is the unix domain socket through which SDS server communicates with
	// ingress gateway proxies.
	GatewayUDSPath string

	// CertFile is the path of Cert File for gRPC server TLS settings.
	CertFile string

	// KeyFile is the path of Key File for gRPC server TLS settings.
	KeyFile string

	// CAEndpoint is the CA endpoint to which node agent sends CSR request.
	CAEndpoint string

	// The CA provider name.
	CAProviderName string

	// TrustDomain corresponds to the trust root of a system.
	// https://github.com/spiffe/spiffe/blob/master/standards/SPIFFE-ID.md#21-trust-domain
	TrustDomain string

	// The Vault CA address.
	VaultAddress string

	// The Vault auth path.
	VaultAuthPath string

	// The Vault role.
	VaultRole string

	// The Vault sign CSR path.
	VaultSignCsrPath string

	// The Vault TLS root certificate.
	VaultTLSRootCert string

	// GrpcServer is an already configured (shared) grpc server. If set, the agent will just register on the server.
	GrpcServer *grpc.Server

	// Recycle job running interval (to clean up staled sds client connections).
	RecycleInterval time.Duration

	// Debug server port from which node_agent serves SDS configuration dumps
	DebugPort int

	// EnableWorkloadSDS indicates whether node agent works as SDS server for workload proxies.
	EnableWorkloadSDS bool

	// EnableGatewaySDS indicates whether node agent works as ingress gateway agent.
	EnableGatewaySDS bool

	// UseLocalJWT is set when the sds server should use its own local JWT, and not expect one
	// from the UDS caller. Used when it runs in the same container with Envoy.
	UseLocalJWT bool

	// Whether to generate PKCS#8 private keys.
	Pkcs8Keys bool

	// Location of JWTPath to connect to CA.
	JWTPath string

	// OutputKeyCertToDir is the directory for output the key and certificate
	OutputKeyCertToDir string

	// ProvCert is the directory for client to provide the key and certificate to server
	// when do mtls
	ProvCert string

	// whether  ControlPlaneAuthPolicy is MUTUAL_TLS
	TLSEnabled bool

	// ClusterID is the cluster where the agent resides.
	// Normally initialized from ISTIO_META_CLUSTER_ID - after a tortuous journey it
	// makes its way into the ClusterID metadata of Citadel gRPC request to create the cert.
	// Didn't find much doc - but I suspect used for 'central cluster' use cases - so should
	// match the cluster name set in the MC setup.
	ClusterID string

	// The type of Elliptical Signature algorithm to use
	// when generating private keys. Currently only ECDSA is supported.
	ECCSigAlg string

	// FileMountedCerts indicates whether the proxy is using file
	// mounted certs created by a foreign CA. Refresh is managed by the external
	// CA, by updating the Secret or VM file. We will watch the file for changes
	// or check before the cert expires. This assumes the certs are in the
	// well-known ./etc/certs location.
	FileMountedCerts bool

	// PilotCertProvider is the provider of the Pilot certificate (PILOT_CERT_PROVIDER env)
	// Determines the root CA file to use for connecting to CA gRPC:
	// - istiod
	// - kubernetes
	// - custom
	PilotCertProvider string

	// secret TTL.
	SecretTTL time.Duration

	// The initial backoff time in millisecond to avoid the thundering herd problem.
	InitialBackoffInMilliSec int64

	// secret should be rotated if:
	// time.Now.After(<secret ExpireTime> - <secret TTL> * SecretRotationGracePeriodRatio)
	SecretRotationGracePeriodRatio float64

	// Key rotation job running interval.
	RotationInterval time.Duration

	// Cached secret will be removed from cache if (time.now - secretItem.CreatedTime >= evictionDuration), this prevents cache growing indefinitely.
	EvictionDuration time.Duration

	// authentication provider specific plugins, will exchange the token
	// For example exchange long lived refresh with access tokens.
	// Used by the secret fetcher when signing CSRs.
	TokenExchangers []TokenExchanger

	// CSR requires a token. This is a property of the CA.
	// The default value is false because Istiod does not require a token in CSR.
	UseTokenForCSR bool

	// credential fetcher.
	CredFetcher CredFetcher

	// credential identity provider
	CredIdentityProvider string

	// Namespace corresponding to workload
	WorkloadNamespace string

	// Name of the Service Account
	ServiceAccount string
}
```

### 涉及环境变量

| 环境变量                               | 描述                                                         | 默认值              |
| -------------------------------------- | ------------------------------------------------------------ | ------------------- |
| ENABLE_INGRESS_GATEWAY_SDS             | Enable provisioning gateway secrets. Requires Secret read permission | false               |
| CA_PROVIDER                            | 认证提供者名称                                               | Citadel             |
| TRUST_DOMAIN                           | spiffe的信任域                                               | cluster.local       |
| PKCS8_KEY                              | Whether to generate PKCS#8 private keys                      | false               |
| ECC_SIGNATURE_ALGORITHM                | The type of ECC signature algorithm to use when generating private keys |                     |
| STALED_CONNECTION_RECYCLE_RUN_INTERVAL | The ticker to detect and close stale connections. 以ns为单位 | 5m                  |
| SECRET_TTL                             | The cert lifetime requested by istio agent. 以ns为单位       | 24h                 |
| SECRET_GRACE_PERIOD_RATIO              | The grace period ratio for the cert rotation, by default 0.5. | 0.5                 |
| SECRET_ROTATION_CHECK_INTERVAL         | The ticker to detect and rotate the certificates, by default 5 minutes | 5m                  |
| INITIAL_BACKOFF_MSEC                   |                                                              | 0                   |
| CREDENTIAL_IDENTITY_PROVIDER           | The identity provider for credential. Currently default supported identity provider is GoogleComputeEngine | GoogleComputeEngine |
| CREDENTIAL_FETCHER_TYPE                | The type of the credential fetcher. Currently supported types include GoogleComputeEngine |                     |
| JWT_POLICY                             | The JWT validation policy，决定jwt路径                       | third-party-jwt     |



```go
			secOpts := &security.Options{
				PilotCertProvider:  pilotCertProvider,
				OutputKeyCertToDir: outputKeyCertToDir,
				ProvCert:           provCert,
				JWTPath:            jwtPath,
				ClusterID:          clusterIDVar.Get(),
				FileMountedCerts:   fileMountedCertsEnv,
				CAEndpoint:         caEndpointEnv,
				UseTokenForCSR:     useTokenForCSREnv,
				CredFetcher:        nil,
				WorkloadNamespace:  podNamespace,
				ServiceAccount:     serviceAccountVar.Get(),
			}
			// If not set explicitly, default to the discovery address.
			if caEndpointEnv == "" {
				secOpts.CAEndpoint = proxyConfig.DiscoveryAddress
			}

			secOpts.EnableWorkloadSDS = true
			secOpts.EnableGatewaySDS = enableGatewaySDSEnv
			secOpts.CAProviderName = caProviderEnv

			secOpts.TrustDomain = trustDomainEnv
			secOpts.Pkcs8Keys = pkcs8KeysEnv
			secOpts.ECCSigAlg = eccSigAlgEnv
			secOpts.RecycleInterval = staledConnectionRecycleIntervalEnv
			secOpts.SecretTTL = secretTTLEnv
			secOpts.SecretRotationGracePeriodRatio = secretRotationGracePeriodRatioEnv
			secOpts.RotationInterval = secretRotationIntervalEnv
			secOpts.InitialBackoffInMilliSec = int64(initialBackoffInMilliSecEnv)
			// Disable the secret eviction for istio agent.
			secOpts.EvictionDuration = 0

			// TODO (liminw): CredFetcher is a general interface. In 1.7, we limit the use on GCE only because
			// GCE is the only supported plugin at the moment.
			if credFetcherTypeEnv == security.GCE {
				secOpts.CredIdentityProvider = credIdentityProvider
				credFetcher, err := credentialfetcher.NewCredFetcher(credFetcherTypeEnv, secOpts.TrustDomain, jwtPath, secOpts.CredIdentityProvider)
				if err != nil {
					return fmt.Errorf("failed to create credential fetcher: %v", err)
				}
				log.Infof("Start credential fetcher of %s type in %s trust domain", credFetcherTypeEnv, secOpts.TrustDomain)
				secOpts.CredFetcher = credFetcher
			}
```



## agentConfig::istio_agent.AgentConfig

### 初始化

```go
			agentConfig := &istio_agent.AgentConfig{
				XDSRootCerts: xdsRootCA,
				CARootCerts:  caRootCA,
				XDSHeaders:   map[string]string{},
			}
			extractXDSHeadersFromEnv(agentConfig)
			if proxyXDSViaAgent {
				agentConfig.ProxyXDSViaAgent = true
				agentConfig.DNSCapture = dnsCaptureByAgent
				agentConfig.ProxyNamespace = podNamespace
				agentConfig.ProxyDomain = role.DNSDomain
			}
```

### 依赖项

| 名称                   | 类型 | 描述                                                         | 默认值 |
| ---------------------- | :--- | ------------------------------------------------------------ | ------ |
| XDS_ROOT_CA            | env  | Explicitly set the root CA to expect for the XDS connection. |        |
| PROXY_XDS_VIA_AGENT    | env  | If set to true, envoy will proxy XDS calls via the agent instead of directly connecting to istiod. This option will be removed once the feature is stabilized. | true   |
| POD_NAMESPACE          |      |                                                              |        |
| CA_ROOT_CA             | env  | Explicitly set the root CA to expect for the CA connection.  |        |
| ISTIO_META_DNS_CAPTURE | env  | If set to true, enable the capture of outgoing DNS packets on port 53, redirecting to istio-agent on :15053 | false  |
| XDS_HEADER_*           | env  | 其他xdx头部环境变量                                          |        |



# 二、运行实体

## SDS与xDS服务代理端

### 初始化

```go
			agentConfig := &istio_agent.AgentConfig{
				XDSRootCerts: xdsRootCA,
				CARootCerts:  caRootCA,
				XDSHeaders:   map[string]string{},
			}
			extractXDSHeadersFromEnv(agentConfig)
			if proxyXDSViaAgent {
				agentConfig.ProxyXDSViaAgent = true
				agentConfig.DNSCapture = dnsCaptureByAgent
				agentConfig.ProxyNamespace = podNamespace
				agentConfig.ProxyDomain = role.DNSDomain
			}
			sa := istio_agent.NewAgent(&proxyConfig, agentConfig, secOpts)
```

### 启动

代码中将会启动通过`server, err := sds.NewServer(sa.secOpts, sa.WorkloadSecrets, gatewaySecretCache)`代码启动SDS服务，通过`sa.xdsProxy, err = initXdsProxy(sa)`启动对envoy的xDS代理，由istio-agent为envoy提供服务端，实际是istiod服务器的xDS代理。两者都是通过本地的unix socket文件为envoy提供入口，对应文件位于**/etc/*istio/proxy**

```go
// /root/src/go/src/github.com/istio/pkg/istio-agent/agent.go

// Simplified SDS setup. This is called if and only if user has explicitly mounted a K8S JWT token, and is not
// using a hostPath mounted or external SDS server.
//
// 1. External CA: requires authenticating the trusted JWT AND validating the SAN against the JWT.
//    For example Google CA
//
// 2. Indirect, using istiod: using K8S cert.
//
// 3. Monitor mode - watching secret in same namespace ( Ingress)
//
// 4. TODO: File watching, for backward compat/migration from mounted secrets.
func (sa *Agent) Start(isSidecar bool, podNamespace string) (*sds.Server, error) {

	// TODO: remove the caching, workload has a single cert
	if sa.WorkloadSecrets == nil {
		sa.WorkloadSecrets, _ = sa.newWorkloadSecretCache()
	}

	var gatewaySecretCache *cache.SecretCache
	if !isSidecar {
		if gatewaySdsExists() {
			log.Infof("Starting gateway SDS")
			sa.secOpts.EnableGatewaySDS = true
			// TODO: what is the setting for ingress ?
			sa.secOpts.GatewayUDSPath = strings.TrimPrefix(model.CredentialNameSDSUdsPath, "unix:")
			gatewaySecretCache = sa.newSecretCache(podNamespace)
		} else {
			log.Infof("Skipping gateway SDS")
			sa.secOpts.EnableGatewaySDS = false
		}
	}

	server, err := sds.NewServer(sa.secOpts, sa.WorkloadSecrets, gatewaySecretCache)
	if err != nil {
		return nil, err
	}

	// Start the local XDS generator.
	if sa.localXDSGenerator != nil {
		err = sa.startXDSGenerator(sa.proxyConfig, sa.WorkloadSecrets, podNamespace)
		if err != nil {
			return nil, fmt.Errorf("failed to start local xds generator: %v", err)
		}
	}

	if err = sa.initLocalDNSServer(isSidecar); err != nil {
		return nil, fmt.Errorf("failed to start local DNS server: %v", err)
	}
	if sa.cfg.ProxyXDSViaAgent {
		sa.xdsProxy, err = initXdsProxy(sa)
		if err != nil {
			return nil, fmt.Errorf("failed to start xds proxy: %v", err)
		}
	}
	return server, nil
}
```

#### SDS服务代理

istio-agent通过本地文件**/etc/istio/proxy/SDS**为envoy提供了sds的grpc服务，在envoy中定义的sds服务如下，而istio-agent则实现了`StreamSecrets`和`FetchSecrets`

```go
// /root/src/go/pkg/mod/github.com/envoyproxy/go-control-plane@v0.9.8-0.20201019204000-12785f608982/envoy/service/secret/v3/sds.pb.go

var _SecretDiscoveryService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "envoy.service.secret.v3.SecretDiscoveryService",
	HandlerType: (*SecretDiscoveryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FetchSecrets",
			Handler:    _SecretDiscoveryService_FetchSecrets_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "DeltaSecrets",
			Handler:       _SecretDiscoveryService_DeltaSecrets_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "StreamSecrets",
			Handler:       _SecretDiscoveryService_StreamSecrets_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "envoy/service/secret/v3/sds.proto",
}
```



在`FetchSecrets`中获取的token为k8s的audience `istio-token` token(k8s的audience为istio-ca且过期自动乱转的serviceAccountToken)或从请求上下文中的header的`istio_sds_credentail_header`或`authorization`值。

通过`s.st.GenerateSecret(ctx, connID, resourceName, token)`完成密钥的重新生成。

`s.st`实际为初始化部分代码`sa.newWorkloadSecretCache()`创建的一个`SecretCache`实例，`SecretCache`实现了`SecretManager`接口。

在`SecretCache`生成密钥的实现中通过与istiod通信，通过csr验签等完成密钥和证书的生成。详细代码查看`/root/src/go/src/github.com/istio/security/pkg/nodeagent/cache/secretcache.go`的`func (sc *SecretCache) generateSecret(ctx context.Context, token string, connKey ConnKey, t time.Time) (*security.SecretItem, error)`

与istiod的连接与csr的签名代码查看`/root/src/go/src/github.com/istio/security/pkg/nodeagent/caclient/providers/citadel/client.go`。

整个接口流程整理后如下，与官网说明的istio pki流程符合

1. envoy通过sds API 分发证书和密钥请求

2. istio-agent接收sds请求，创建密钥和csr。并发往istiod进行签名，这个过程中使用的是为k8s分配的serviceAccountToken `istio-token` 

3. istiod的ca对csr进行签名生成证书。

4. istio-agent接收来自istiod的证书，并将证书与密钥通过grpc发给envoy

   

```go
// /root/src/go/src/github.com/istio/security/pkg/nodeagent/sds/sdsservice.go

// FetchSecrets generates and returns secret from SecretManager in response to DiscoveryRequest
func (s *sdsservice) FetchSecrets(ctx context.Context, discReq *discovery.DiscoveryRequest) (*discovery.DiscoveryResponse, error) {
	token := ""
	if s.localJWT {
		t, err := s.getToken()
		if err != nil {
			sdsServiceLog.Errorf("Failed to get credential token: %v", err)
			return nil, err
		}
		token = t
	} else if !s.skipToken {
		t, err := getCredentialToken(ctx)
		if err != nil {
			sdsServiceLog.Errorf("Failed to get credential token: %v", err)
			return nil, err
		}
		token = t
	}

	resourceName, err := getResourceName(discReq)
	if err != nil {
		sdsServiceLog.Errorf("Close connection. Error: %v", err)
		return nil, err
	}

	connID := constructConnectionID(discReq.Node.Id)
	secret, err := s.st.GenerateSecret(ctx, connID, resourceName, token)
	if err != nil {
		sdsServiceLog.Errorf("Failed to get secret for proxy %q from secret cache: %v", connID, err)
		return nil, err
	}

	// Output the key and cert to a directory, if some applications need to read them from local file system.
	if err = nodeagentutil.OutputKeyCertToDir(s.outputKeyCertToDir, secret.PrivateKey,
		secret.CertificateChain, secret.RootCert); err != nil {
		sdsServiceLog.Errorf("(%v) error when output the key and cert: %v",
			connID, err)
		return nil, err
	}
	return sdsDiscoveryResponse(secret, resourceName, discReq.TypeUrl)
}
```

ps: 

1. 卷istiod-ca-cert路径为`/var/run/secrets/istio/root-cert.pem`为istiod的服务器证书，而卷istiod-token为`/var/run/secrets/token`为分配给istio-agent用于请求istiod的jwt token



#### xDS服务代理

istio-agent的xds代理模块中`func initXdsProxy(ia *Agent) (*XdsProxy, error)`完成对下游(envoy)提供xds服务，而实际上是在envoy通过流式grpc向istio-agent进行xds请求时，istio-agent创建与istiod的连接从而获取对应的xds对象。envoy通过本地unix socket文件**/etc/istio/proxy/XDS**与istio-agent通信。

`func initXdsProxy(ia *Agent) (*XdsProxy, error)`创建了本地xds服务端的unix socket文件**/etc/istio/proxy/XDS**。

与istiod的连接在envoy连入istio-agent时进行，代码如下所示，在接收到grpc连接后会创建子线程用于接收来自envoy的请求，同时当前线程将会创建与istiod的连接。通过一个`ProxyConnection`对象表示两个连接的状态，在istiod连接和envoy连接中传递状态、请求和响应。

相关代码位于`github.com/istio/pkg/istio-agent/xds_proxy.go`

```go
// /root/src/go/src/github.com/istio/pkg/istio-agent/xds_proxy.go

// Every time envoy makes a fresh connection to the agent, we reestablish a new connection to the upstream xds
// This ensures that a new connection between istiod and agent doesn't end up consuming pending messages from envoy
// as the new connection may not go to the same istiod. Vice versa case also applies.
func (p *XdsProxy) StreamAggregatedResources(downstream discovery.AggregatedDiscoveryService_StreamAggregatedResourcesServer) error {
	proxyLog.Infof("Envoy ADS stream established")

	con := &ProxyConnection{
		upstreamError:   make(chan error),
		downstreamError: make(chan error),
		requestsChan:    make(chan *discovery.DiscoveryRequest, 10),
		responsesChan:   make(chan *discovery.DiscoveryResponse, 10),
		stopChan:        make(chan struct{}),
		downstream:      downstream,
	}

	p.RegisterStream(con)

	// Handle downstream xds
	firstNDSSent := false
	go func() {
		for {
			// From Envoy
			req, err := downstream.Recv()
			if err != nil {
				con.downstreamError <- err
				return
			}
			// forward to istiod
			con.requestsChan <- req
			if p.localDNSServer != nil && !firstNDSSent && req.TypeUrl == v3.ListenerType {
				// fire off an initial NDS request
				con.requestsChan <- &discovery.DiscoveryRequest{
					TypeUrl: v3.NameTableType,
				}
				firstNDSSent = true
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	upstreamConn, err := grpc.DialContext(ctx, p.istiodAddress, p.istiodDialOptions...)
	if err != nil {
		proxyLog.Errorf("failed to connect to upstream %s: %v", p.istiodAddress, err)
		metrics.IstiodConnectionFailures.Increment()
		return err
	}
	defer upstreamConn.Close()

	xds := discovery.NewAggregatedDiscoveryServiceClient(upstreamConn)
	ctx = metadata.AppendToOutgoingContext(context.Background(), "ClusterID", p.clusterID)
	if p.agent.cfg.XDSHeaders != nil {
		for k, v := range p.agent.cfg.XDSHeaders {
			ctx = metadata.AppendToOutgoingContext(ctx, k, v)
		}
	}
	// We must propagate upstream termination to Envoy. This ensures that we resume the full XDS sequence on new connection
	return p.HandleUpstream(ctx, con, xds)
}
```

##### 补充

istio-agent的xds服务端定义的service代码如下，是第三方的envoyproxy库提供的，从代码可以看出其要求istio-agent实现`StreamAggregatedResources`和`DeltaAggregatedResources`用于流式grpc与Delta xDS的实现，但是在istio-agent中仅实现了`func (p *XdsProxy) StreamAggregatedResources(downstream discovery.AggregatedDiscoveryService_StreamAggregatedResourcesServer) error`

```go
// /root/src/go/pkg/mod/github.com/envoyproxy/go-control-plane@v0.9.8-0.20201019204000-12785f608982/envoy/service/discovery/v3/ads.pb.go

var _AggregatedDiscoveryService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "envoy.service.discovery.v3.AggregatedDiscoveryService",
	HandlerType: (*AggregatedDiscoveryServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamAggregatedResources",
			Handler:       _AggregatedDiscoveryService_StreamAggregatedResources_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "DeltaAggregatedResources",
			Handler:       _AggregatedDiscoveryService_DeltaAggregatedResources_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "envoy/service/discovery/v3/ads.proto",
}
```



### 依赖项

依赖于`role`,`secOpts`,`proxyConfig`





## STS

### 初始化

```go
			// If security token service (STS) port is not zero, start STS server and
			// listen on STS port for STS requests. For STS, see
			// https://tools.ietf.org/html/draft-ietf-oauth-token-exchange-16.
			if stsPort > 0 {
				localHostAddr := localHostIPv4
				if proxyIPv6 {
					localHostAddr = localHostIPv6
				}
				tokenManager := tokenmanager.CreateTokenManager(tokenManagerPlugin,
					tokenmanager.Config{CredFetcher: secOpts.CredFetcher, TrustDomain: secOpts.TrustDomain})
				stsServer, err := stsserver.NewServer(stsserver.Config{
					LocalHostAddr: localHostAddr,
					LocalPort:     stsPort,
				}, tokenManager)
				if err != nil {
					return err
				}
				defer stsServer.Stop()
			}
```

### 依赖项

| 名称               | 类型              | 描述                                                         | 默认值              |
| ------------------ | ----------------- | ------------------------------------------------------------ | ------------------- |
| tokenManagerPlugin | flag              | Token provider specific plugin name.                         | GoogleTokenExchange |
| secOpts            | &security.Options |                                                              |                     |
| stsPort            | flag              | HTTP Port on which to serve Security Token Service (STS). If zero, STS service will not be provided. | 0                   |



## envoy

### 初始化

在完成envoy实例的初始化后，通过`agent.restart`完成对envoy的启动，该函数本意为实现热更新，但实际实现上函数未监控任何文件，仅在启动时完成当前轮次的注册，并最终通过调用`func (e *envoy) Run(config interface{}, epoch int, abort <-chan error) error `函数完成子进程envoy实例的启动(详细代码可查看**/root/src/go/src/github.com/istio/pkg/envoy/agent.go**)。这个的agent更多指的父进程对子进程envoy实例的控制。

```go
			envoyProxy := envoy.NewProxy(envoy.ProxyConfig{
				Config:              proxyConfig,
				Node:                role.ServiceNode(),
				LogLevel:            proxyLogLevel,
				ComponentLogLevel:   proxyComponentLogLevel,
				PilotSubjectAltName: pilotSAN,
				NodeIPs:             role.IPAddresses,
				STSPort:             stsPort,
				OutlierLogPath:      outlierLogPath,
				PilotCertProvider:   pilotCertProvider,
				ProvCert:            sa.FindRootCAForXDS(),
				Sidecar:             role.Type == model.SidecarProxy,
				ProxyViaAgent:       agentConfig.ProxyXDSViaAgent,
				CallCredentials:     callCredentials.Get(),
			})

			drainDuration, _ := types.DurationFromProto(proxyConfig.TerminationDrainDuration)
			if ds, f := features.TerminationDrainDuration.Lookup(); f {
				// Legacy environment variable is set, us that instead
				drainDuration = time.Second * time.Duration(ds)
			}

			agent := envoy.NewAgent(envoyProxy, drainDuration)

			// Watcher is also kicking envoy start.
			// 向watcher注册agent.restart函数，企图实现基于配置文件的热更新，但实际上未实现热更新而是envoy的启动
			watcher := envoy.NewWatcher(agent.Restart)
			go watcher.Run(ctx)

			// On SIGINT or SIGTERM, cancel the context, triggering a graceful shutdown
			go cmd.WaitSignalFunc(cancel)

			// 阻塞等待envoy当前轮次停止，并清除活动轮次记录
			return agent.Run(ctx)
```

### 运行

代码中通过`func (e *envoy) Run(config interface{}, epoch int, abort <-chan error) error`完成envoy组件所需的配置和启动。代码中的`CreateFileForEpoch`完成从模板文件(/var/lib/istio/envoy/envoy_bootstrap_tmpl.json)读取go template，再配合参数完成对envoy启动配置文件(/etc/istio/proxy/envoy-rev0.json)的设置。最终通过`exec.Command`创建envoy进程。

envoy的配置文件名称含有当前轮次，但由于未实现热更新，所以目前一定为0。

```go
// /root/src/go/src/github.com/istio/pkg/envoy/proxy.go
func (e *envoy) Run(config interface{}, epoch int, abort <-chan error) error {
	var fname string
	// Note: the cert checking still works, the generated file is updated if certs are changed.
	// We just don't save the generated file, but use a custom one instead. Pilot will keep
	// monitoring the certs and restart if the content of the certs changes.
	if len(e.Config.CustomConfigFile) > 0 {
		// there is a custom configuration. Don't write our own config - but keep watching the certs.
		fname = e.Config.CustomConfigFile
	} else {
		discHost := strings.Split(e.Config.DiscoveryAddress, ":")[0]
		out, err := bootstrap.New(bootstrap.Config{
			Node:                e.Node,
			Proxy:               &e.Config,
			PilotSubjectAltName: e.PilotSubjectAltName,
			LocalEnv:            os.Environ(),
			NodeIPs:             e.NodeIPs,
			STSPort:             e.STSPort,
			ProxyViaAgent:       e.ProxyViaAgent,
			OutlierLogPath:      e.OutlierLogPath,
			PilotCertProvider:   e.PilotCertProvider,
			ProvCert:            e.ProvCert,
			CallCredentials:     e.CallCredentials,
			DiscoveryHost:       discHost,
		}).CreateFileForEpoch(epoch)
		if err != nil {
			log.Errora("Failed to generate bootstrap config: ", err)
			os.Exit(1) // Prevent infinite loop attempting to write the file, let k8s/systemd report
		}
		fname = out
	}

	// spin up a new Envoy process
	args := e.args(fname, epoch, istioBootstrapOverrideVar.Get())
	log.Infof("Envoy command: %v", args)

	/* #nosec */
	cmd := exec.Command(e.Config.BinaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-abort:
		log.Warnf("Aborting epoch %d", epoch)
		if errKill := cmd.Process.Kill(); errKill != nil {
			log.Warnf("killing epoch %d caused an error %v", epoch, errKill)
		}
		return err
	case err := <-done:
		return err
	}
}
```



```go
// /root/src/go/src/github.com/istio/pkg/bootstrap/instance.go
func (i *instance) CreateFileForEpoch(epoch int) (string, error) {
	// Create the output file.
	if err := os.MkdirAll(i.Proxy.ConfigPath, 0700); err != nil {
		return "", err
	}

	templateFile := getEffectiveTemplatePath(i.Proxy)

	outputFilePath := configFile(i.Proxy.ConfigPath, templateFile, epoch)
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = outputFile.Close() }()

	// Write the content of the file.
	if err := i.WriteTo(templateFile, outputFile); err != nil {
		return "", err
	}

	return outputFilePath, err
}
```





### 依赖项

| 名称                               | 类型                    | 描述                                                         | 默认值     |
| ---------------------------------- | ----------------------- | ------------------------------------------------------------ | ---------- |
| proxyConfig                        | meshconfig.ProxyConfig  |                                                              |            |
| role                               | model.Proxy             |                                                              |            |
| proxyLogLevel                      | flag                    | The log level used to start the Envoy proxy                  | warning    |
| proxyComponentLogLevel             | flag                    | The component log level used to start the Envoy proxy        | misc:error |
| stsPort                            | flag                    | HTTP Port on which to serve Security Token Service (STS). If zero, STS service will not be provided. | 0          |
| outlierLogPath                     | flag                    | The log path for outlier detection                           |            |
| PILOT_CERT_PROVIDER                | env                     | The provider of Pilot DNS certificate.                       | istiod     |
| SDS                                |                         |                                                              |            |
| agentConfig                        | istio_agent.AgentConfig |                                                              |            |
| CALL_CREDENTIALS                   | env                     | Use JWT directly instead of MTLS                             | false      |
| TERMINATION_DRAIN_DURATION_SECONDS | env                     |                                                              | 5          |



# 三、statusServer

在proxy启动中存在以下一段代码，表示服务启动时若提供了`StatusPort`则启动状态服务，该变量为`func constructProxyConfig() (meshconfig.ProxyConfig, error)`函数从annotation中的`status.sidecar.istio.io/port`获取或`func DefaultProxyConfig() meshconfig.ProxyConfig`中直接设置为默认值15020。

```go
			// If a status port was provided, start handling status probes.
			if proxyConfig.StatusPort > 0 {
				if err := initStatusServer(ctx, proxyIPv6, proxyConfig); err != nil {
					return err
				}
			}
```

## 配置statusServer

代码中通过读取环境变量`ISTIO_PROMETHEUS_ANNOTATIONS`决定是否启动prometheus。环境变量`ISTIO_KUBE_APP_PROBERS`决定了pod的服务探针，并且目前仅支持`httpGet`模式。

在创建`statusServer`的过程中完成对负载应用的prometheus的配置、对负载的http服务的探针客户端创建、envoy的管理端口、statusServer服务端口的设置。

```go
// /root/src/go/src/github.com/istio/pilot/cmd/pilot-agent/main.go

func initStatusServer(ctx context.Context, proxyIPv6 bool, proxyConfig meshconfig.ProxyConfig) error {
	localHostAddr := localHostIPv4
	if proxyIPv6 {
		localHostAddr = localHostIPv6
	}
	prober := kubeAppProberNameVar.Get()
	statusServer, err := status.NewServer(status.Config{
		LocalHostAddr:  localHostAddr,
		AdminPort:      uint16(proxyConfig.ProxyAdminPort),
		StatusPort:     uint16(proxyConfig.StatusPort),
		KubeAppProbers: prober,
		NodeType:       role.Type,
	})
	if err != nil {
		return err
	}
	go statusServer.Run(ctx)
	return nil
}


// /root/src/go/src/github.com/istio/pilot/cmd/pilot-agent/status/server.go

const (
	// readyPath is for the pilot agent readiness itself.
	readyPath = "/healthz/ready"
	// quitPath is to notify the pilot agent to quit.
	quitPath = "/quitquitquit"
	// KubeAppProberEnvName is the name of the command line flag for pilot agent to pass app prober config.
	// The json encoded string to pass app HTTP probe information from injector(istioctl or webhook).
	// For example, ISTIO_KUBE_APP_PROBERS='{"/app-health/httpbin/livez":{"httpGet":{"path": "/hello", "port": 8080}}.
	// indicates that httpbin container liveness prober port is 8080 and probing path is /hello.
	// This environment variable should never be set manually.
	KubeAppProberEnvName = "ISTIO_KUBE_APP_PROBERS"
)


// NewServer creates a new status server.
func NewServer(config Config) (*Server, error) {
	s := &Server{
		statusPort: config.StatusPort,
		ready: &ready.Probe{
			LocalHostAddr: config.LocalHostAddr,
			AdminPort:     config.AdminPort,
			NodeType:      config.NodeType,
		},
		envoyStatsPort: 15090,
	}

	// Enable prometheus server if its configured and a sidecar
	// Because port 15020 is exposed in the gateway Services, we cannot safely serve this endpoint
	// If we need to do this in the future, we should use envoy to do routing or have another port to make this internal
	// only. For now, its not needed for gateway, as we can just get Envoy stats directly, but if we
	// want to expose istio-agent metrics we may want to revisit this.
	if cfg, f := PrometheusScrapingConfig.Lookup(); config.NodeType == model.SidecarProxy && f {
		var prom PrometheusScrapeConfiguration
		if err := json.Unmarshal([]byte(cfg), &prom); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s: %v", PrometheusScrapingConfig.Name, err)
		}
		log.Infof("Prometheus scraping configuration: %v", prom)
		s.prometheus = &prom
		if s.prometheus.Path == "" {
			s.prometheus.Path = "/metrics"
		}
		if s.prometheus.Port == "" {
			s.prometheus.Port = "80"
		}
		if s.prometheus.Port == strconv.Itoa(int(config.StatusPort)) {
			return nil, fmt.Errorf("invalid prometheus scrape configuration: "+
				"application port is the same as agent port, which may lead to a recursive loop. "+
				"Ensure pod does not have prometheus.io/port=%d label, or that injection is not happening multiple times", config.StatusPort)
		}
	}

	if config.KubeAppProbers == "" {
		return s, nil
	}
	if err := json.Unmarshal([]byte(config.KubeAppProbers), &s.appKubeProbers); err != nil {
		return nil, fmt.Errorf("failed to decode app prober err = %v, json string = %v", err, config.KubeAppProbers)
	}

	s.appProbeClient = make(map[string]*http.Client, len(s.appKubeProbers))
	// Validate the map key matching the regex pattern.
	for path, prober := range s.appKubeProbers {
		if !appProberPattern.Match([]byte(path)) {
			return nil, fmt.Errorf(`invalid key, must be in form of regex pattern ^/app-health/[^\/]+/(livez|readyz)$`)
		}
		if prober.HTTPGet == nil {
			return nil, fmt.Errorf(`invalid prober type, must be of type httpGet`)
		}
		if prober.HTTPGet.Port.Type != intstr.Int {
			return nil, fmt.Errorf("invalid prober config for %v, the port must be int type", path)
		}
		// Construct a http client and cache it in order to reuse the connection.
		s.appProbeClient[path] = &http.Client{
			Timeout: time.Duration(prober.TimeoutSeconds) * time.Second,
			// We skip the verification since kubelet skips the verification for HTTPS prober as well
			// https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/#configure-probes
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}

	return s, nil
}
```



## statusServer的启动与能力

`statusServer`通过http服务提供了四个restful api接口，分别用于对envoy状态检查(是否异常，是否已经准备完成)、prometheus状态获取(envoy、agent和应用的指标的合并)、服务退出和应用健康检查。

```go
// Run opens a the status port and begins accepting probes.
func (s *Server) Run(ctx context.Context) {
	log.Infof("Opening status port %d\n", s.statusPort)

	mux := http.NewServeMux()

	// Add the handler for ready probes.
	mux.HandleFunc(readyPath, s.handleReadyProbe)
	mux.HandleFunc(`/stats/prometheus`, s.handleStats)
	mux.HandleFunc(quitPath, s.handleQuit)
	mux.HandleFunc("/app-health/", s.handleAppProbe)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.statusPort))
	if err != nil {
		log.Errorf("Error listening on status port: %v", err.Error())
		return
	}
	// for testing.
	if s.statusPort == 0 {
		addrs := strings.Split(l.Addr().String(), ":")
		allocatedPort, _ := strconv.Atoi(addrs[len(addrs)-1])
		s.mutex.Lock()
		s.statusPort = uint16(allocatedPort)
		s.mutex.Unlock()
	}
	defer l.Close()

	go func() {
		if err := http.Serve(l, mux); err != nil {
			log.Errora(err)
			select {
			case <-ctx.Done():
				// We are shutting down already, don't trigger SIGTERM
				return
			default:
				// If the server errors then pilot-agent can never pass readiness or liveness probes
				// Therefore, trigger graceful termination by sending SIGTERM to the binary pid
				notifyExit()
			}
		}
	}()

	// Wait for the agent to be shut down.
	<-ctx.Done()
	log.Info("Status server has successfully terminated")
}

```



