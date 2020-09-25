## 概述
用于检测Rust程序以收集结构化、基于事件的诊断信息的一个框架。
普通日志追踪难以处理多任务情况下的日志行为，例如一个线程中多个任务混合进行，日志行会被混合到一起。tracing使用`span`和`event`的概念将任务和任务事件结合并结构化日志

## 核心概念
* `spans`: 代表一段时间，包含开始和结束时间，当程序执行一段内容会一个工作单元时，会进入span的上下文，在任务结束后退出span。线程正则执行时指向的span称为线程的当前span
* `event`: 标识当一个trace被记录时发生的某些事情，在span的上下文中数显
* `subscribers`: `subscribers`的实现在`span`和`event`发生时，将它们进行记录或聚合了。在`span`进入或离开且`event`发生时，`subscriber`将会被触发
    * `event method`: `event`发生时被调用
    * `enter method`: 进入`span`时被调用
    * `exit method`: 离开`span`时被调用

## 重要宏、属性与函数
* `span!(target, parent, Level, name)`: 获取一个span，如果target和parent span为指定，则默认值由宏被调用处或当前span决定
    * 可设置span的特殊字段，字段数量最多32个
    * `pub fn enter()`: 进入span
```rust
// 字段设置与简写
let user = "ferris";

span!(Level::TRACE, "login", user);
// is equivalent to:
span!(Level::TRACE, "login", user = user);
let email = "ferris@rust-lang.org";
span!(Level::TRACE, "login", user, user.email = email);

let user = User {
    name: "ferris",
    email: "ferris@rust-lang.org",
};
// the span will have the fields `user.name = "ferris"` and
// `user.email = "ferris@rust-lang.org"`.
span!(Level::TRACE, "login", user.name, user.email);
```
* `event!(Level, ctx, field="...")`: 触发一个event，可以设置字段值，规则类似`span`
* `#[instruct]`:  将一个函数便捷的进入span中
* `?`: 指定字段应该被`fmt::Debug`记录的简写
    * `event!(Level::TRACE, greeting = ?my_struct);`等效于`event!(Level::TRACE, greeting = tracing::field::debug(&my_struct));`
* `%`: 指定字段应该被`fmt::Displat`记录的简写
* `trace!, debug!, info!, warn!, and error!`: 与`log`crate同名的宏用于生成`event`,便于对`log`crate的替换


