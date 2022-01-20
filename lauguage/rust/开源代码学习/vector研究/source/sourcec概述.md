## 配置对象SourceConfig生成

```rust
#[derive(Deserialize, Serialize, Debug, Default)]
#[serde(deny_unknown_fields)]
pub struct ConfigBuilder {
    #[serde(flatten)]
    pub global: GlobalOptions,
    #[cfg(feature = "api")]
    #[serde(default)]
    pub api: api::Options,
    #[serde(default)]
    pub sources: IndexMap<String, Box<dyn SourceConfig>>,
    #[serde(default)]
    pub sinks: IndexMap<String, SinkOuter>,
    #[serde(default)]
    pub transforms: IndexMap<String, TransformOuter>,
    #[serde(default)]
    pub tests: Vec<TestDefinition>,
}

#[async_trait::async_trait]
#[typetag::serde(tag = "type")]
pub trait SourceConfig: core::fmt::Debug + Send + Sync {
    fn build(
        &self,
        name: &str,
        globals: &GlobalOptions,
        shutdown: ShutdownSignal,
        out: Pipeline,
    ) -> crate::Result<sources::Source>;

    async fn build_async(
        &self,
        name: &str,
        globals: &GlobalOptions,
        shutdown: ShutdownSignal,
        out: Pipeline,
    ) -> crate::Result<sources::Source> {
        self.build(name, globals, shutdown, out)
    }

    fn output_type(&self) -> DataType;

    fn source_type(&self) -> &'static str;
}
```

`#[typetag::serde(tag = "type")]`属性在反序列化文本的`sources`配置时，实现了`trait SourceConfig`的且结构体名称满足字段`type`的值时，则根据反序列化对应config对象，最后通过`SourceConfig.build_async`创建`Source`对象



### 框架创建source

```rust
// Build sources
    for (name, source) in config
        .sources
        .iter()
        .filter(|(name, _)| diff.sources.contains_new(&name))
    {
        let (tx, rx) = mpsc::channel(1000);
        let pipeline = Pipeline::from_sender(tx);

        let typetag = source.source_type();

        let (shutdown_signal, force_shutdown_tripwire) = shutdown_coordinator.register_source(name);

        let server = match source
            .build_async(&name, &config.global, shutdown_signal, pipeline)
            .await
        {
            Err(error) => {
                errors.push(format!("Source \"{}\": {}", name, error));
                continue;
            }
            Ok(server) => server,
        };

        let (output, control) = Fanout::new();
        let pump = rx.forward(output).map(|_| ()).compat();
        let pump = Task::new(name, typetag, pump);

        // The force_shutdown_tripwire is a Future that when it resolves means that this source
        // has failed to shut down gracefully within its allotted time window and instead should be
        // forcibly shut down.  We accomplish this by select()-ing on the server Task with the
        // force_shutdown_tripwire.  That means that if the force_shutdown_tripwire resolves while
        // the server Task is still running the Task will simply be dropped on the floor.
        let server = server
            .select(force_shutdown_tripwire)
            .map(|_| debug!("Finished"))
            .map_err(|_| ())
            .compat();
        let server = Task::new(name, typetag, server);

        outputs.insert(name.clone(), control);
        tasks.insert(name.clone(), pump);
        source_tasks.insert(name.clone(), server);
    }
```

souce实例代码为`source.build_async(&name, &config.global, shutdown_signal, pipeline)`，`build_async`默认实现为调用`source.build`。`pipeline`为source的输出管道，而pipeline在代码上下文可以看出是`mpsc::channel(1000)`的输入端。

`let pump = rx.forward(output).map(|_| ()).compat();`将source的输出端，重定向到到`fanout`。

最终`source`的数据会输出到所有的`fanout.sink`中，而`fanout`的`control`仅用于为`fanout`组装/卸载`fanout.sink`。在`fanout`的`process_control_messge`占用了该实例内部工作的大部分时间，用与根据`control`写入的控制信息装载/卸载`fanout.sink`，实际上可以考虑在重载或启动时进行，没有必要在每次数据输出`start_send`时调用并判断。

最终的`source_task`为`source`与`shutdown_coordinator`的组合任务。

```
source ---> pipeline(tx) ---> rx ---forward---> fanout ---> fanout.sinks

source_task = source.select(force_shutdown_tripwire)
```



`pub type Source = Box<dyn Future<Item = (), Error = ()> + Send>;`

source的构建中`build_async`默认实现调用了`build`创建一个`source`对象，而一个source对象代表了一个`futures01::Future`对象



