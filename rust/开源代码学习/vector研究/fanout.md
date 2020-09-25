## ## 概念

实现了 `futrues::Sink`的结构体，Fanout内部维护了多个sink，用于判断这些sink是否全都已经进入`Async::Ready`状态，是对多个`sink`的组合包装。

## ## 数据结构

```rust
pub struct Fanout {
    sinks: Vec<(String, RouterSink)>,		// 内部维护的Sink
    i: usize,
    control_channel: mpsc::UnboundedReceiver<ControlMessage>,	// 接收Fanout的信号处理，指导Fanout行为
}
```

## ## 行为分析

`Fanout`的出现的目的是为了同时维护对多个`Sink`的同步行为，确保所有`sink`都处于`就绪状态`才会触发`Fanout`的就绪。

每次`Fanout`的`poll_complete`和`start_send`都会首先触发`process_control_messages`从控制信号管道中获取控制信号，并根据信号对`Fanout`内部的sinks进行调整

```rust
pub fn process_control_messages(&mut self) {
    while let Ok(Async::Ready(Some(message))) = self.control_channel.poll() {
        match message {
            ControlMessage::Add(name, sink) => self.add(name, sink),
            ControlMessage::Remove(name) => self.remove(&name),
            ControlMessage::Replace(name, sink) => self.replace(name, sink),
        }
    }
}
```



