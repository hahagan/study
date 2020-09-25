## opts参数
```rust
#[derive(StructOpt, Debug)]
#[structopt(rename_all = "kebab-case")]
pub struct RootOpts {
    /// Read configuration from one or more files. Wildcard paths are supported.
    /// If zero files are specified the default config path
    /// `/etc/vector/vector.toml` will be targeted.
    #[structopt(name = "config", short, long, env = "VECTOR_CONFIG")]
    pub config_paths: Vec<PathBuf>,

    /// Exit on startup if any sinks fail healthchecks
    #[structopt(short, long, env = "VECTOR_REQUIRE_HEALTHY")]
    pub require_healthy: bool,

    /// Number of threads to use for processing (default is number of available cores)
    #[structopt(short, long, env = "VECTOR_THREADS")]
    pub threads: Option<usize>,

    /// Enable more detailed internal logging. Repeat to increase level. Overridden by `--quiet`.
    #[structopt(short, long, parse(from_occurrences))]
    pub verbose: u8,

    /// Reduce detail of internal logging. Repeat to reduce further. Overrides `--verbose`.
    #[structopt(short, long, parse(from_occurrences))]
    pub quiet: u8,

    /// Set the logging format
    #[structopt(long, default_value = "text", possible_values = &["text", "json"])]
    pub log_format: LogFormat,

    /// Control when ANSI terminal formatting is used.
    ///
    /// By default `vector` will try and detect if `stdout` is a terminal, if it is
    /// ANSI will be enabled. Otherwise it will be disabled. By providing this flag with
    /// the `--color always` option will always enable ANSI terminal formatting. `--color never`
    /// will disable all ANSI terminal formatting. `--color auto` will attempt
    /// to detect it automatically.
    #[structopt(long, default_value = "auto", possible_values = &["auto", "always", "never"])]
    pub color: Color,

    /// Watch for changes in configuration file, and reload accordingly.
    #[structopt(short, long, env = "VECTOR_WATCH_CONFIG")]
    pub watch_config: bool,
}
```
* quite: 代表级别日志级别"warn", "error", "off"
* verbose: 代表日志级别"info","debug","trace", 会被quiet覆盖
* config_paths: 指向vector配置vector.toml
* threads: 处理线程，默认未cpu核数

## 子命令
```rust
#[derive(StructOpt, Debug)]
#[structopt(rename_all = "kebab-case")]
pub enum SubCommand {
    /// Validate the target config, then exit.
    Validate(validate::Opts),

    /// Generate a Vector configuration containing a list of components.
    Generate(generate::Opts),

    /// List available components, then exit.
    List(list::Opts),

    /// Run Vector config unit tests, then exit. This command is experimental and therefore subject to change.
    /// For guidance on how to write unit tests check out: https://vector.dev/docs/setup/guides/unit-testing/
    Test(unit_test::Opts),
}
```
