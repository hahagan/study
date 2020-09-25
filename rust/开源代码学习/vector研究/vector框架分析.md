## 数据流组装
vector的数据流结构由source，transform和output三种组件组合得到。每个组件间都会通过一个channel管道将组件连接。channel分为输入端和输出端。
数据流中的数据流动方向为 source -> transform -> output
vector的配置使用的时toml格式，每种组件在使用者配置数据流时，需要为每个组件声明一个名字。
每个组件通过inputs配置决定该与哪些上游组件组合。

下方的配置模板代表了一个数据流，每个`--->`都代表了一个channel，两端为其输入和输出端
```
                                                    +---> transforms.apache_sampler ---> sinks.es_cluster
sources.apache_logs ---> transforms.apache_parser---+
                                                    +---> sinks.s3_archives
```

### 数据流生成

1. 在组件拓扑结构调整阶段
    1. 首先对source组件的依赖进行调整
        * 将以source为上游的transform组件进行调整，使得transfrom组件的channel有原source调整为新的source组件的channel输出端
    2. 对transform组件进行调整
    3. 对sink组件进行调整

2. 在数据流启动阶段
    1. 首先对source组件进行启动
        * 已有source如果需要变动则首先重建
        * 对已有source重建完成后创建新增加的source
    2. 其次对Transform组件进行启动处理
    3. 最后对sink进行处理
    注: 首先启动input，再启动transform和sink，这个意味这数据接收在数据处理之前进行，如果这个过程中没有对数据进行缓存，期间出现异常(例如output启动失败等)，会导致数据丢失，在logstash早期版本也是这么设计，但是在最新版本中已经将input放到最后启动。在vector中数据的输入输出通过rust的异步数据能力，可以将数据缓存在


## vector的启动
1. 获取cli，解析命令行参数与环境变量
2. `trace::init(color, json, levels.as_str())`使用[tracing](../tracing.md) crate作为程序的追踪与日志输出系统
3. `metrics::init()`初始化一个全局[metrics_runtime::receiver](../metrics_runtime.md)，并设置全局`controller`以获取监控信息
4. 识别是否包含子命令(Validate,ListTest,Generate)，如果是则执行子命令并退出
5. 获取"vector.toml"配置,启动配置监控重载异步行为
6. 读取配置
```rust
    // 获取命令行参数
    let opts = root_opts.root;
    let sub_command = root_opts.sub_command;

    // 建立日志输出
    trace::init(color, json, levels.as_str());
    // 建立监控信息仓库
    metrics::init().expect("metrics initialization failed");

    // 执行子命令
    if let Some(s) = sub_command {
        std::process::exit(match s {
            SubCommand::Validate(v) => validate::validate(&v, color).await,
            SubCommand::List(l) => list::cmd(&l),
            SubCommand::Test(t) => unit_test::cmd(&t),
            SubCommand::Generate(g) => generate::cmd(&g),
        })
    };

    // 获取配置文件路径
    let config_paths = config::process_paths(&opts.config_paths).unwrap_or_else(|| {
            std::process::exit(exitcode::CONFIG);
    });

    // 启动子线程，基于文件系统事件驱动库notify监控配置文件变化
    if opts.watch_config {
        config::watcher::spawn_thread(&config_paths, None)
    }

    // 读取配置,该函数主要是打开配置文件进行toml读取并合并环境变量，生成一个包ConfigBuilder对象实例
    let config = config::load_from_paths(&config_paths)

    // 比较新旧配置，获得新旧配置差异，主要时通过比较配置的名称决定差异，相同名称比较配置项
    let diff = ConfigDiff::initial(&config);

    // 根据配置创建各个组件。 
	// 重要，详情查看"异步框架"
    let pieces = topology::build_or_log_errors(&config, &diff)


```
### 主要结构体
```rust
// config/builder.rs
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

// config/mod.rs
#[derive(Deserialize, Serialize, Debug)]
pub struct SinkOuter {
    #[serde(default)]
    pub buffer: crate::buffers::BufferConfig,
    #[serde(default = "healthcheck_default")]
    pub healthcheck: bool,
    pub inputs: Vec<String>,
    #[serde(flatten)]
    pub inner: Box<dyn SinkConfig>,     // SinkConfig为一个trait，定义了每个Sink配置项需要实现的方法
}

// config/mod.rs
#[derive(Deserialize, Serialize, Debug)]
pub struct TransformOuter {
    pub inputs: Vec<String>,
    #[serde(flatten)]
    pub inner: Box<dyn TransformConfig>,    // TransformConfig为一个trait，定义了每个transform配置项需要实现的方法
}

// buffers/mods.rs
// 定义了一个buffer的类型和行为，其存储容量和当buff满时的行为
#[derive(Deserialize, Serialize, Debug)]
#[serde(tag = "type")]
#[serde(rename_all = "snake_case")]
pub enum BufferConfig {
    Memory {
        #[serde(default = "BufferConfig::memory_max_events")]
        max_events: usize,
        #[serde(default)]
        when_full: WhenFull,
    },
    #[cfg(feature = "leveldb")]
    Disk {
        max_size: usize,
        #[serde(default)]
        when_full: WhenFull,
    },
}

// config/diff.rs
// 新旧配置项差异
pub struct ConfigDiff {
    pub sources: Difference,
    pub transforms: Difference,
    pub sinks: Difference,
}
```

### ref
* https://docs.rs/tracing/0.1.19/tracing/

## transform组件分析
### transform执行调度
```rust
pub trait Transform: Send {
    fn transform(&mut self, event: Event) -> Option<Event>;

    fn transform_into(&mut self, output: &mut Vec<Event>, event: Event) {
        if let Some(transformed) = self.transform(event) {
            output.push(transformed);
        }
    }

    fn transform_stream(
        self: Box<Self>,
        input_rx: Box<dyn Stream<Item = Event, Error = ()> + Send>,
    ) -> Box<dyn Stream<Item = Event, Error = ()> + Send>
    where
        Self: 'static,
    {
        let mut me = self;
        Box::new(
            input_rx
                .map(move |event| {
                    let mut output = Vec::with_capacity(1);
                    me.transform_into(&mut output, event);
                    futures01::stream::iter_ok(output.into_iter())
                })
                .flatten(),
        )
    }
}
```
从代码分析Send本质上是一个unsafe trait，Transfrom以Send作为super trait表明其同样为unsafe trait。
transform trait 默认实现了stransform_into和transform_stream方法，**trnsform**方法为每个Transform产品(代表Transfrom的结构体)各自event处理逻辑的实现部分。
transform_stream从transform的接收端获取待处理event，通过调用transform_into方法处理event并输出，而transform_into则调用transform形成处理完毕的event。

### Transform创建
每个Transfrom产品都需要实现`TransfromConfig trait`，用于声明自己的transfrom名称以及输入输出类型，并用于生成一个实现了`Transform trait`的结构体实例。
```rust
pub trait TransformConfig: core::fmt::Debug + Send + Sync {
    // 创建transfrom对象
    fn build(&self, cx: TransformContext) -> crate::Result<Box<dyn transforms::Transform>>;

    async fn build_async(
        &self,
        cx: TransformContext,
    ) -> crate::Result<Box<dyn transforms::Transform>> {
        self.build(cx)
    }

    // 接收的日志类型Log或metric
    fn input_type(&self) -> DataType;

    // 输出的日志类型Log或metric
    fn output_type(&self) -> DataType;

    // transfrom类型
    fn transform_type(&self) -> &'static str;

    /// Allows a transform configuration to expand itself into multiple "child"
    /// transformations to replace it. This allows a transform to act as a macro
    /// for various patterns.
    fn expand(&mut self) -> crate::Result<Option<IndexMap<String, Box<dyn TransformConfig>>>> {
        Ok(None)
    }
}
```