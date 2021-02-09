# 一、安装行为

istioctl为[istio主项目](https://github.com/istio/istio)下的子项目。`istioctl install`命令为 `istio.io/istio/istioctl/pkg/install`包下生成的命令。安装命令的主要流程如下

1. 首先由`runrunApplyCmd`通过调用`getProfileAndEnabledComponents`获取各从指定的iop资源对象文件和manifest中读取需要安装的对象(仅用于检查和打印作为提示信息)，而后通过调用`InstallManifests`进行安装。
2. `InstallManifests`调用`manifest.GenerateConfig`根据指定的iop资源对象文件，从manifests中读取profile配置并与iop合并为对象`*iopv1alpha1.IstioOperator`，为后续加载chart和选择准备。再通过` reconciler.Reconcile()`完成对部署模板的渲染和资源对象安装。最后嗲用函数`saveIOPToCluster(reconciler, iopStr)`将iop对象作为k8s资源对象保存，完成整个安装行为。

整个安装主要思路为基于manifest定义默认的charts文件，通过iop对象决定各种参数值，并通过iop对象渲染对应的charts文件。而最终iop对象的定义由三个位置的配置决定，**manifest目录下的profile，用户输入的iop资源对象文件，用户输入的命令行参数**，这三处优先级逐渐升高，高优先级覆盖低优先级。

最终安装的控制面组件实体主要由`Base,pilot,Cni,IstiodRemote, IngressGateways,EgressGateways`5种组成前三种组件一个istio(一个从istio控制面也算一个istio)仅含一个，后两种组件则根据需要可以有多个。除了这5种组件外还可以有额外的插件例如istio提供了coredns。这些实体组件的行为一般由服务运行的环境变量和配置文件(或者说configmap)决定。

```go
// 代码位置/root/src/go/src/github.com/istio/operator/cmd/mesh/install.go


func runApplyCmd(cmd *cobra.Command, rootArgs *rootArgs, iArgs *installArgs, logOpts *log.Options) error {
	l := clog.NewConsoleLogger(cmd.OutOrStdout(), cmd.ErrOrStderr(), installerScope)
	setFlags := applyFlagAliases(iArgs.set, iArgs.manifestsPath, iArgs.revision)
	// Warn users if they use `istioctl install` without any config args.
	if !rootArgs.dryRun && !iArgs.skipConfirmation {
		profile, enabledComponents, err := getProfileAndEnabledComponents(setFlags, iArgs.inFilenames, iArgs.force, l)
		if err != nil {
			return fmt.Errorf("failed to get profile and enabled components: %v", err)
		}
		prompt := fmt.Sprintf("This will install the Istio %s profile with %q components into the cluster. Proceed? (y/N)", profile, enabledComponents)
		if profile == "empty" {
			prompt = fmt.Sprintf("This will install the Istio %s profile into the cluster. Proceed? (y/N)", profile)
		}
		if !confirm(prompt, cmd.OutOrStdout()) {
			cmd.Print("Cancelled.\n")
			os.Exit(1)
		}
	}
	if err := configLogs(logOpts); err != nil {
		return fmt.Errorf("could not configure logs: %s", err)
	}
	if err := InstallManifests(setFlags, iArgs.inFilenames, iArgs.force, rootArgs.dryRun,
		iArgs.kubeConfigPath, iArgs.context, iArgs.readinessTimeout, l); err != nil {
		return fmt.Errorf("failed to install manifests: %v", err)
	}

	return nil
}

// InstallManifests generates manifests from the given input files and --set flag overlays and applies them to the
// cluster. See GenManifests for more description of the manifest generation process.
//  force   validation warnings are written to logger but command is not aborted
//  dryRun  all operations are done but nothing is written
func InstallManifests(setOverlay []string, inFilenames []string, force bool, dryRun bool,
	kubeConfigPath string, context string, waitTimeout time.Duration, l clog.Logger) error {

	restConfig, clientset, client, err := K8sConfig(kubeConfigPath, context)
	if err != nil {
		return err
	}
	if err := k8sversion.IsK8VersionSupported(clientset, l); err != nil {
		return err
	}
    
    // 从指定的文件和manifests中读取配置并合并为对象*iopv1alpha1.IstioOperator，为后续加载chart和选择准备
	_, iop, err := manifest.GenerateConfig(inFilenames, setOverlay, force, restConfig, l)
	if err != nil {
		return err
	}

	if err := createNamespace(clientset, iop.Namespace); err != nil {
		return err
	}

	// Needed in case we are running a test through this path that doesn't start a new process.
	cache.FlushObjectCaches()
	opts := &helmreconciler.Options{DryRun: dryRun, Log: l, WaitTimeout: waitTimeout, ProgressLog: progress.NewLog(),
		Force: force}
	reconciler, err := helmreconciler.NewHelmReconciler(client, restConfig, iop, opts)
	if err != nil {
		return err
	}
    
    // 完成对部署模板的渲染和资源对象安装
	status, err := reconciler.Reconcile()
	if err != nil {
		return fmt.Errorf("errors occurred during operation: %v", err)
	}
	if status.Status != v1alpha1.InstallStatus_HEALTHY {
		return fmt.Errorf("errors occurred during operation")
	}

	opts.ProgressLog.SetState(progress.StateComplete)

	// Save a copy of what was installed as a CR in the cluster under an internal name.
	iop.Name = savedIOPName(iop)
	if iop.Annotations == nil {
		iop.Annotations = make(map[string]string)
	}
	iop.Annotations[istiocontrolplane.IgnoreReconcileAnnotation] = "true"
	iopStr, err := util.MarshalWithJSONPB(iop)
	if err != nil {
		return err
	}

    // 保存已安装的iop对象
	return saveIOPToCluster(reconciler, iopStr)
}
```



# 二、iop安装

`Reconcile()`中通过`RenderCharts`函数完成从manifest中加载各个component的charts，并通过iop对象进行渲染。

```go
// 代码位于/root/src/go/src/github.com/istio/operator/pkg/helmreconciler/reconciler.go
// Reconcile reconciles the associated resources.
func (h *HelmReconciler) Reconcile() (*v1alpha1.InstallStatus, error) {
    // 利用helm库完成模板渲染工作
	manifestMap, err := h.RenderCharts()
	if err != nil {
		return nil, err
	}

	// 完成对象的安装，根据对象hash值确定需要更新或创建的对象并执行相应工作
	status := h.processRecursive(manifestMap)
	
	h.opts.ProgressLog.SetState(progress.StatePruning)
	pruneErr := h.Prune(manifestMap, false)
	h.reportPrunedObjectKind()
	return status, pruneErr
}
```

```go
// 代码位于/root/src/go/src/github.com/istio/operator/pkg/helmreconciler/render.go

// RenderCharts renders charts for h.
func (h *HelmReconciler) RenderCharts() (name.ManifestMap, error) {
	iopSpec := h.iop.Spec
	if err := validate.CheckIstioOperatorSpec(iopSpec, false); err != nil {
		if !h.opts.Force {
			return nil, err
		}
		h.opts.Log.PrintErr(fmt.Sprintf("spec invalid; continuing because of --force: %v\n", err))
	}

	t := translate.NewTranslator()
	
	// 生成控制面所需资源对象
	cp, err := controlplane.NewIstioControlPlane(iopSpec, t)
	if err != nil {
		return nil, err
	}
	
	// 从指定的manifests中加载charts
	if err := cp.Run(); err != nil {
		return nil, fmt.Errorf("failed to create Istio control plane with spec: \n%v\nerror: %s", iopSpec, err)
	}
	
	// 完成charts的渲染
	manifests, errs := cp.RenderManifest()
	if errs != nil {
		err = errs.ToError()
	}
	
	h.manifests = manifests
	
	return manifests, err
}
```



## 三、iop生成与合并

`ReadYamlProfile`首先是读取iop资源对象文件得到用户生成的iop的yaml文本以及profile名称

```go
// 代码位于/root/src/go/src/github.com/istio/operator/pkg/manifest/shared.go

// GenerateConfig creates an IstioOperatorSpec from the following sources, overlaid sequentially:
// 1. Compiled in base, or optionally base from paths pointing to one or multiple ICP/IOP files at inFilenames.
// 2. Profile overlay, if non-default overlay is selected. This also comes either from compiled in or path specified in IOP contained in inFilenames.
// 3. User overlays stored in inFilenames.
// 4. setOverlayYAML, which comes from --set flag passed to manifest command.
//
// Note that the user overlay at inFilenames can optionally contain a file path to a set of profiles different from the
// ones that are compiled in. If it does, the starting point will be the base and profile YAMLs at that file path.
// Otherwise it will be the compiled in profile YAMLs.
// In step 3, the remaining fields in the same user overlay are applied on the resulting profile base.
// The force flag causes validation errors not to abort but only emit log/console warnings.
func GenerateConfig(inFilenames []string, setFlags []string, force bool, kubeConfig *rest.Config,
	l clog.Logger) (string, *iopv1alpha1.IstioOperator, error) {
	if err := validateSetFlags(setFlags); err != nil {
		return "", nil, err
	}

	fy, profile, err := ReadYamlProfile(inFilenames, setFlags, force, l)
	if err != nil {
		return "", nil, err
	}

	iopsString, iops, err := GenIOPFromProfile(profile, fy, setFlags, force, false, kubeConfig, l)

	if err != nil {
		return "", nil, err
	}

	errs, warning := validation.ValidateConfig(false, iops.Spec)
	if warning != "" {
		l.LogAndError(warning)
	}

	if errs.ToError() != nil {
		return "", nil, fmt.Errorf("generated config failed semantic validation: %v", errs)
	}
	return iopsString, iops, nil
}
```



1. `GetProfileYAML`读取profiles目录下所有的profile，并以文件名作为区分依据。当profile非`default`时，会继续读取`default`profile，并将两者通过`util.OverlayIOP`函数进行合并返回生成的iop的文本格式。

2. `overlaySetFlagValues`完成对用户定义的iop完成与命令行参数的合并，`TranslateK8SfromValueToIOP`完成iop转变为内部使用的iop对象的yaml格式。

3. 通过`util.OverlayIOP`将用户定义的iop合并到从profile中读取的iop。最后通过`unmarshalAndValidateIOP`完成内部iop对象的生成。

```go
// GenIOPFromProfile generates an IstioOperator from the given profile name or path, and overlay YAMLs from user
// files and the --set flag. If successful, it returns an IstioOperator string and struct.
func GenIOPFromProfile(profileOrPath, fileOverlayYAML string, setFlags []string, skipValidation, allowUnknownField bool,
	kubeConfig *rest.Config, l clog.Logger) (string, *iopv1alpha1.IstioOperator, error) {

	......
    
	// To generate the base profileOrPath for overlaying with user values, we need the installPackagePath where the profiles
	// can be found, and the selected profileOrPath. Both of these can come from either the user overlay file or --set flag.
	outYAML, err := helm.GetProfileYAML(installPackagePath, profileOrPath)
	if err != nil {
		return "", nil, err
	}

	......

	// Combine file and --set overlays and translate any K8s settings in values to IOP format. Users should not set
	// these but we have to support this path until it's deprecated.
	overlayYAML, err := overlaySetFlagValues(fileOverlayYAML, setFlags)
	if err != nil {
		return "", nil, err
	}
	t := translate.NewReverseTranslator()
	overlayYAML, err = t.TranslateK8SfromValueToIOP(overlayYAML)
	if err != nil {
		return "", nil, fmt.Errorf("could not overlay k8s settings from values to IOP: %s", err)
	}

	// Merge user file and --set flags.
	outYAML, err = util.OverlayIOP(outYAML, overlayYAML)
	if err != nil {
		return "", nil, fmt.Errorf("could not overlay user config over base: %s", err)
	}

	if err := name.ScanBundledAddonComponents(installPackagePath); err != nil {
		return "", nil, err
	}
	// If enablement came from user values overlay (file or --set), translate into addonComponents paths and overlay that.
	outYAML, err = translate.OverlayValuesEnablement(outYAML, overlayYAML, overlayYAML)
	if err != nil {
		return "", nil, err
	}

	finalIOP, err := unmarshalAndValidateIOP(outYAML, skipValidation, allowUnknownField, l)
	if err != nil {
		return "", nil, err
	}
	// InstallPackagePath may have been a URL, change to extracted to local file path.
	finalIOP.Spec.InstallPackagePath = installPackagePath
	return util.ToYAMLWithJSONPB(finalIOP), finalIOP, nil
}
```



# 四、IOP对象

```go
// IstioOperator is a CustomResourceDefinition (CRD) for an operator.
type IstioOperator struct {
	Kind                 string                      `protobuf:"bytes,5,opt,name=kind,proto3" json:"kind,omitempty"`
	ApiVersion           string                      `protobuf:"bytes,6,opt,name=apiVersion,proto3" json:"apiVersion,omitempty"`
	Spec                 *v1alpha1.IstioOperatorSpec `protobuf:"bytes,7,opt,name=spec,proto3" json:"spec,omitempty"`
	Status				 *v1alpha1.InstallStatus     `protobuf:"bytes,8,opt,name=status,proto3" json:"status,omitempty"`
	v11.ObjectMeta       `json:"metadata,omitempty" protobuf:"bytes,9,opt,name=metadata"`
	v11.TypeMeta         `json:",inline"`
	Placeholder          string   `protobuf:"bytes,111,opt,name=placeholder,proto3" json:"placeholder,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}


// IstioOperatorSpec defines the desired installed state of Istio components.
// The spec is a used to define a customization of the default profile values that are supplied with each Istio release.
// Because the spec is a customization API, specifying an empty IstioOperatorSpec results in a default Istio
// component values.
//
// ```yaml
// apiVersion: install.istio.io/v1alpha1
// kind: IstioOperator
// spec:
//   profile: default
//   hub: gcr.io/istio-testing
//   tag: latest
//   revision: 1-8-0
//   meshConfig:
//     accessLogFile: /dev/stdout
//     enableTracing: true
//   components:
//     egressGateways:
//     - name: istio-egressgateway
//       enabled: true
// ```
//
type IstioOperatorSpec struct {
	// Path or name for the profile e.g.
	//
	// * minimal (looks in profiles dir for a file called minimal.yaml)
	// * /tmp/istio/install/values/custom/custom-install.yaml (local file path)
	//
	// default profile is used if this field is unset.
	Profile string `protobuf:"bytes,10,opt,name=profile,proto3" json:"profile,omitempty"`
	// Path for the install package. e.g.
	//
	// * /tmp/istio-installer/nightly (local file path)
	//
	InstallPackagePath string `protobuf:"bytes,11,opt,name=install_package_path,json=installPackagePath,proto3" json:"installPackagePath,omitempty"`
	// Root for docker image paths e.g. `docker.io/istio`
	Hub string `protobuf:"bytes,12,opt,name=hub,proto3" json:"hub,omitempty"`
	// Version tag for docker images e.g. `1.7.2`
	Tag interface{} `protobuf:"bytes,13,opt,name=tag,proto3" json:"tag,omitempty"`
	// $hide_from_docs
	// Resource suffix is appended to all resources installed by each component.
	// Never implemented; replaced by revision.
	ResourceSuffix string `protobuf:"bytes,14,opt,name=resource_suffix,json=resourceSuffix,proto3" json:"resourceSuffix,omitempty"` // Deprecated: Do not use.
	// Namespace to install control plane resources into. If unset, Istio will be installed into the same namespace
	// as the `IstioOperator` CR.
	Namespace string `protobuf:"bytes,15,opt,name=namespace,proto3" json:"namespace,omitempty"`
	// Identify the revision this installation is associated with.
	// This option is currently experimental.
	Revision string `protobuf:"bytes,16,opt,name=revision,proto3" json:"revision,omitempty"`
	// Config used by control plane components internally.
	MeshConfig map[string]interface{} `protobuf:"bytes,40,opt,name=mesh_config,json=meshConfig,proto3" json:"meshConfig,omitempty"`
	// Kubernetes resource settings, enablement and component-specific settings that are not internal to the
	// component.
	Components *IstioComponentSetSpec `protobuf:"bytes,50,opt,name=components,proto3" json:"components,omitempty"`
	// Extra addon components which are not explicitly specified above.
	AddonComponents map[string]*ExternalComponentSpec `protobuf:"bytes,51,rep,name=addon_components,json=addonComponents,proto3" json:"addonComponents,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Overrides for default `values.yaml`. This is a validated pass-through to Helm templates.
	// See the [Helm installation options](https://istio.io/v1.5/docs/reference/config/installation-options/) for schema details.
	// Anything that is available in `IstioOperatorSpec` should be set above rather than using the passthrough. This
	// includes Kubernetes resource settings for components in `KubernetesResourcesSpec`.
	Values map[string]interface{} `protobuf:"bytes,100,opt,name=values,proto3" json:"values,omitempty"`
	// Unvalidated overrides for default `values.yaml`. Used for custom templates where new parameters are added.
	UnvalidatedValues    map[string]interface{} `protobuf:"bytes,101,opt,name=unvalidated_values,json=unvalidatedValues,proto3" json:"unvalidatedValues,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}


// IstioComponentSpec defines the desired installed state of Istio components.
type IstioComponentSetSpec struct {
	Base                 *BaseComponentSpec `protobuf:"bytes,29,opt,name=base,proto3" json:"base,omitempty"`
	Pilot                *ComponentSpec     `protobuf:"bytes,30,opt,name=pilot,proto3" json:"pilot,omitempty"`
	Cni                  *ComponentSpec     `protobuf:"bytes,38,opt,name=cni,proto3" json:"cni,omitempty"`
	IstiodRemote         *ComponentSpec     `protobuf:"bytes,39,opt,name=istiod_remote,json=istiodRemote,proto3" json:"istiodRemote,omitempty"`
	IngressGateways      []*GatewaySpec     `protobuf:"bytes,40,rep,name=ingress_gateways,json=ingressGateways,proto3" json:"ingressGateways,omitempty"`
	EgressGateways       []*GatewaySpec     `protobuf:"bytes,41,rep,name=egress_gateways,json=egressGateways,proto3" json:"egressGateways,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}
```

