### 主要代码

`sourceconfig`反序列化

```rust
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

`#[typetag::serde(tag = "type")]`属性使得所有实现了`trait SourceConfig`的对象，





```rust
Box::new(future::lazy(move || {
        info!(message = "Starting file server.", ?include, ?exclude);

        // sizing here is just a guess
        let (tx, rx) = futures01::sync::mpsc::channel(100);

        // This closure is overcomplicated because of the compatibility layer.
        let wrap_with_line_agg = |rx, config| {
            // 转换rx为特定格式的输出流
            let rx = StreamExt::filter_map(Compat01As03::new(rx), |val| {
                futures::future::ready(val.ok())
            });
            let logic = line_agg::Logic::new(config);
            Box::new(Compat::new(
                LineAgg::new(rx.map(|(line, src)| (src, line, ())), logic)
                    .map(|(src, line, _context)| (line, src))
                    .map(Ok),
            ))
        };
        let messages: Box<dyn Stream<Item = (Bytes, String), Error = ()> + Send> =
            if let Some(ref multiline_config) = multiline_config {
                wrap_with_line_agg(
                    rx,
                    multiline_config.try_into().unwrap(), // validated in build
                )
            } else if let Some(msi) = message_start_indicator {
                wrap_with_line_agg(
                    rx,
                    line_agg::Config::for_legacy(
                        Regex::new(&msi).unwrap(), // validated in build
                        multi_line_timeout,
                    ),
                )
            } else {
                Box::new(rx)
            };

        // Once file server ends this will run until it has finished processing remaining
        // logs in the queue.
        let span = current_span();
        let span2 = span.clone();
        tokio::spawn(
            messages
                .map(move |(msg, file): (Bytes, String)| {
                    let _enter = span2.enter();
                    create_event(msg, file, &host_key, &hostname, &file_key)
                })
                .forward(out.sink_map_err(|e| error!(%e)))
                .map(|_| ())
                .compat()
                .instrument(span),
        );

        let span = info_span!("file_server");
        spawn_blocking(move || {
            let _enter = span.enter();
            let result = file_server.run(Compat01As03Sink::new(tx), shutdown.compat());
            // Panic if we encounter any error originating from the file server.
            // We're at the `spawn_blocking` call, the panic will be caught and
            // passed to the `JoinHandle` error, similar to the usual threads.
            result.unwrap();
        })
        .boxed()
        .compat()
        .map_err(|error| error!(message="file server unexpectedly stopped.",%error))
    }))
```

`let (tx, rx) = futures01::sync::mpsc::channel(100);`创建了一个大小为100的有限管道，其输入端用于传递从文件中读取的数据，输出端封装为多行合并数据流，并最终封装为代表`source`的数据流

`wrap_with_line_agg`根据多行合并配置将`rx`进行封装，从`rx`中读取数据和数据来源，即代码行`LineAgg::new(rx.map(|(line, src)| (src, line, ())), logic)`，`logic`基于`let logic = line_agg::Logic::new(config);`创建，`config`变量为`multiline_config.try_into().unwrap()`转换，而`impl TryFrom<&MultilineConfig> for line_agg::Config`使得`try_into()`能够自动使用`try_from`完成转换。

此时的`message`变量代表了从`rx`读取数据并多行合并的处理逻辑，目前返回的数据为数据内容和数据所属文件，因此在进一步对`message`进行封装，将message的输出内容通过`create_event`生成事件，并将事件数据流重定向到`ouput`也就是创建数据源时创建的`pipeline`

在最后的`spawn_blocking`代码处，创建了一个可阻塞任务线程，用于执行文件数据的读取，其输出结果为`tx`，即将文件数据流重定向到多行合并处理

```
file_server(tx) --->  LineAgg(rx) ---> create_event ---forward---> output(pipeline)
```

