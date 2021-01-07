[官网](https://istio.io/latest/docs/reference/config/networking/virtual-service/#VirtualService)

[toc]

# VirtualService

## hosts::string[]
## gateways::string[]
## http::HTTPRoute[]

### name::string
### match::HTTPMatchRequest[]

#### name::string
#### uri::StringMatch
#### scheme::StringMatch
#### method::StringMatch
#### authority::StringMatch
#### headers::map<string, StringMatch>
#### port::uint32
#### sourceLabels::map<string, string>
#### gateways::string[]
#### queryParams::map<string, StringMatch>
#### ignoreUriCase::bool
#### withoutHeaders::map<string, StringMatch>
#### sourceNamespace::string



### route::HTTPRouteDestination[]
### redirect::HTTPRedirect
### delegate::Delegate
### rewrite::HTTPRewrite
### timeout::Duration
### retries::HTTPRetry
### fault::HTTPFaultInjection
### mirror::Destination
### mirrorPercentage::Percent
### corsPolicy::CorsPolicy
### headers::Headers
### mirrorPercent::UInt32Value

## tls::TLSRoute[]

## tcp::TCPRoute[]
## exportTo::string[]









## VirtualService

```mermaid
graph TB;
VirtualService --> hosts
VirtualService --指定选择--> gateways::string_list
VirtualService --> http::HTTPRoute_list
VirtualService --> tls::TLSRoute_list
VirtualService --> tcp::TCPRoute_list
VirtualService --可访问命名空间--> exportTo::string_list



```

### HTTPRoute




```mermaid
graph LR;
    http::HTTPRoute_list --> name::string
    http::HTTPRoute_list --> match::HTTPMatchRequest_list
    http::HTTPRoute_list --> route::HTTPRouteDestination_list
    http::HTTPRoute_list --> redirect::HTTPRedirect
    http::HTTPRoute_list --> delegate::Delegate
    http::HTTPRoute_list --> rewrite::HTTPRewrite
    http::HTTPRoute_list --> timeout::Duration
    http::HTTPRoute_list --> retries::HTTPRetry
    http::HTTPRoute_list --> fault::HTTPFaultInjection
    http::HTTPRoute_list --> mirror::Destination
    http::HTTPRoute_list --> mirrorPercentage::Percent
    http::HTTPRoute_list --> corsPolicy::CorsPolicy
    http::HTTPRoute_list --> mirrorPercent::UInt32Value
    http::HTTPRoute_list --> headers::Headers
    
subgraph HTTPFaultInjection
fault::HTTPFaultInjection --> delay::Delay
fault::HTTPFaultInjection --> abort::Abort
end

    
subgraph HTTPMatchRequest_list
match::HTTPMatchRequest_list --> HTTPMatchRequesName::string
match::HTTPMatchRequest_list --> uri::StringMatch
match::HTTPMatchRequest_list --> scheme::StringMatch
match::HTTPMatchRequest_list --> method::StringMatch
match::HTTPMatchRequest_list --> authority::StringMatch
match::HTTPMatchRequest_list --> headers::map&ltstring,StringMatch&gt
match::HTTPMatchRequest_list --> port::uint32
match::HTTPMatchRequest_list --> sourceLabels::map&ltstring,string&gt
match::HTTPMatchRequest_list --> gateways::string_list
match::HTTPMatchRequest_list --> queryParams::map&ltstring,StringMatch&gt
match::HTTPMatchRequest_list --> ignoreUriCase::bool
match::HTTPMatchRequest_list --> withoutHeaders::map&ltstring,StringMatch&gt
match::HTTPMatchRequest_list --> sourceNamespace::string
end

subgraph HTTPRouteDestination
	route::HTTPRouteDestination_list --> destination::Destination
    route::HTTPRouteDestination_list --> weight::int32
    route::HTTPRouteDestination_list --> headers::Headers
    destination::Destination --> host::string
    destination::Destination --> subset::string
    destination::Destination --> port::PortSelector
   	end
   	
subgraph HTTPRedirect
redirect::HTTPRedirect --> uri::string
redirect::HTTPRedirect --> authority::string
redirect::HTTPRedirect --> redirectCode::uint32
end

subgraph Delegate
delegate::Delegate --> DelegateName::string
delegate::Delegate --> namespace::string
end

subgraph HTTPRewrite
rewrite::HTTPRewrite --> uriHTTPRewrite::string
rewrite::HTTPRewrite --> authorityHTTPRewrite::string
end


```

#### HTTPMatchReques

```mermaid
graph LR;
HTTPMatchReques --> name::string
HTTPMatchReques --> uri::StringMatch
HTTPMatchReques --> scheme::StringMatch
HTTPMatchReques --> method::StringMatch
HTTPMatchReques --> authority::StringMatch
HTTPMatchReques --> headers::map&ltstring,StringMatch&gt
HTTPMatchReques --> port::uint32
HTTPMatchReques --> sourceLabels::map&ltstring,string&gt
HTTPMatchReques --> gateways::string_list
HTTPMatchReques --> queryParams::map&ltstring,StringMatch&gt
HTTPMatchReques --> ignoreUriCase::bool
HTTPMatchReques --> withoutHeaders::map&ltstring,StringMatch&gt
HTTPMatchReques --> sourceNamespace::string
```

