### 创建config对象

```rust
#[derive(Deserialize, Serialize, Debug, Clone)]
// TODO: add back when https://github.com/serde-rs/serde/issues/1358 is addressed
// #[serde(deny_unknown_fields)]
pub struct SocketConfig {
    #[serde(flatten)]
    pub mode: Mode,
}

#[derive(Deserialize, Serialize, Debug, Clone)]
#[serde(tag = "mode", rename_all = "snake_case")]
pub enum Mode {
    Tcp(tcp::TcpConfig),
    Udp(udp::UdpConfig),
    #[cfg(unix)]
    Unix(unix::UnixConfig),
}
```

在反序列化到type类型为`socket`后，此时会继续识别`Mode`值，并将`mode`对象字段根据当前配置继续解析处tcp、udp或unix类型的socket配置。



### 创建socket类型的source

```rust
// source/socket/mod.rs
#[typetag::serde(name = "socket")]
impl SourceConfig for SocketConfig
```

根据配置反序列化得到的SocketConfig,mode类型创建udp，tcp或socket类型的source。



### tcp接收数据与事件生成

tcp根据source配置生成了一个绑定本地网络地址的tcp监听者，每个连接会创建一个`handler_stream`任务，处理连接，并将`handler_stream`加入异步计算任务中。这部分工作同样作为一个`future`作为`Source`任务返回给框架，在框架启动`source`任务时完成真正的`tcp::bind`和连接处理，与`handle_stream`任务创建。

`MaybeTlsListener`为vector在`tikio::net::tcp::TcpListener`上加上`tls`的封装对象，在

```rust
pub(crate) struct MaybeTlsListener {
    listener: TcpListener,			// tikio::net::tcp::TcpListener
    acceptor: Option<SslAcceptor>,
}

	fn run(
        self,
        addr: SocketListenAddr,
        shutdown_timeout_secs: u64,
        tls: MaybeTlsSettings,
        shutdown: ShutdownSignal,
        out: Pipeline,
    ) -> crate::Result<crate::sources::Source> {
		......
        ......
		listener.incoming()
                .take_until(shutdown.clone().compat())
                .for_each(|connection| {
                    let source = self.clone();
                    let out = out.clone();

                    async move {
                        let socket = match connection {
                            Ok(socket) => socket,
                            Err(error) => {
                                error!(
                                    message = "failed to accept socket",
                                    %error
                                );
                                return;
                            }
                        };

                        let peer_addr = socket.peer_addr().ip().to_string();
                        let span = info_span!("connection", %peer_addr);
                        let host = Bytes::from(peer_addr);

                        span.in_scope(|| {
                            let peer_addr = socket.peer_addr();
                            debug!(message = "accepted a new connection", %peer_addr);

                            let fut = handle_stream(shutdown, socket, source, tripwire, host, out);
                            tokio::spawn(fut.instrument(span.clone()));
                        });
                    }
                })
                .map(Ok)
                .await
        };
	......
	......
```

