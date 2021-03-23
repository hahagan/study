[toc]



### 背景

截至2019年初，Envoy仍然是一个静态编译的二进制文件，其所有扩展都在构建时编译。 这意味着提供自定义扩展名的项目（例如Istio）必须维护和分发自己的二进制文件，而不是使用官方的和未经修改的Envoy二进制文件。

对于无法控制其部署的项目，这甚至会带来更多问题，因为对扩展的任何更新或错误修复都需要构建新的二进制文件，生成版本，进行发布，更重要的是需要在生产中重新部署它。

这也意味着在已部署的扩展和配置它们的控制平面之间经常会有版本漂移



### 解决方案

尽管可以使用可动态加载的C++扩展解决部分问题，但目前尚不可行，因为由于Envoy开发的步伐很快，扩展插件并没有稳定的ABI甚至API可用，并且通常来说，更新Envoy需要更改代码，这使得更新版本成为一个手动过程。

相反，我们决定通过使用稳定的ABI在WebAssembly中编写和交付Envoy扩展来解决此问题，因为它带来了许多其他好处（如下所述）。



### 什么是WebAssembly？

WebAssembly（Wasm）是一种新兴的可移植二进制格式，用于执行代码。 在具有明确定义的资源约束以及用于与嵌入式主机环境（即代理）进行通信的API的情况下，该代码以近似原生的速度在内存安全的（对于主机）沙箱中执行。

### 收益

- 敏捷。 扩展可以在运行时直接从控制平面交付和重新加载。 这意味着不仅每个人都可以使用正式的和未修改的代理版本来加载自定义扩展，而且还意味着可以在运行时推送或测试任何错误修复和更新，而无需更新或重新部署新的二进制文件。
- 可靠性和隔离性。 由于扩展是在具有资源限制的沙箱内部署的，因此它们可以崩溃或泄漏内存，而不会导致整个代理宕机。 此外，还可以限制CPU和内存的使用。
- 安全。 因为扩展部署在具有明确定义的API的沙箱内部，用于与代理进行通信，所以扩展具有有限访问权限，并且只能修改连接和请求中的有限数量的属性。 此外，由于代理可以协调这种交互行为，因此它可以隐藏或清除扩展中的敏感信息（例如，HTTP标头中的“Authorization”和“ Cookie”属性，或客户端的IP地址）。
- 多样性。 可以将30多种编程语言编译为WebAssembly模块，从而允许来自所有背景（C，Go，Rust，Java，TypeScript等）的开发人员以他们选择的语言编写Proxy-Wasm扩展。
- 可维护性。 由于扩展是使用独立于代理代码库的标准库编写的，因此我们可以提供稳定的ABI。
- 可移植性。 由于主机环境和扩展之间的接口是与代理无关的，因此可以在各种代理中执行使用Proxy-Wasm编写的扩展，例如 Envoy，NGINX，ATS甚至在gRPC库中（假设它们都实现了标准）。

### 缺点

- 由于需要启动许多虚拟机，每个虚拟机都有自己的内存块，因此内存使用率更高。
- 性能约为c++实现envoy新版本的70%
- 由于需要将大量数据复制进出沙箱，因此对负载进行转码类扩展的性能较低。
- CPU绑定类型扩展的性能降低。 与原生代码相比，速度下降预计将在2倍不到。
- 由于需要引用Wasm运行时，因此二进制大小增加了。 WAVM约为20MB，V8约为10MB。
- WebAssembly生态系统仍处于起步阶段，目前的开发重点是浏览器内部使用，其中JavaScript被视为宿主环境。



### Envoy Proxy WASM SDK扩展

目前提供以下几种语言的sdk

- [C++](https://github.com/proxy-wasm/proxy-wasm-cpp-sdk)
- [Rust](https://github.com/proxy-wasm/proxy-wasm-rust-sdk)
- [AssemblyScript](https://github.com/solo-io/proxy-runtime)
- [Go](https://github.com/mathetake/proxy-wasm-go) - still experimental



主要以rust为例进行基本介绍。在实现filter时需要实现`trait Context`和`trait RootContext`。在wasm插件加载时，RootContext实例会被创建，且声明周期与vm实例相同。RootContext实例将会用于执行filter逻辑，并用于envoy proxy和代码初始化和对存活请求进行处理。

#### API介绍

##### 1. onConfigure

```rust
fn on_configure(&mut self, _plugin_configuration_size: usize) -> bool
```

该接口等价于C++的SDK的`void onConfigure(std::unique_ptr<WasmData> configuration)`,在rust的SDK中定义于`trait RootContext`。

该接口在主机加载wasm模块时调用，传递的参数为WasmData格式的配置。在模块运行在未配置的vm时，会被调用两次，第一次调用传递VM配置，第二次传递模块配置。当VM被几个filter共享时，如果VM已经被配置，则仅会被调用一次，并传递模块配置。该接口仅会在RootContext中调用







### REF

https://lupeier.com/post/webassembly-in-envoy

https://banzaicloud.com/blog/envoy-wasm-filter/

https://pretired.dazwilkin.com/posts/200723/

https://github.com/proxy-wasm/proxy-wasm-cpp-sdk/blob/istio-release/v1.8/docs/wasm_filter.md

https://github.com/proxy-wasm/proxy-wasm-rust-sdk