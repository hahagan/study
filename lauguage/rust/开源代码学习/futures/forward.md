Forward为实现了futures的，用于从上游stream中读取数据，并发送到下游sink中，实现将两个数据管道的同步传递。

```rust
/// src/stream/forward.rs
/// Future for the `Stream::forward` combinator, which sends a stream of values
/// to a sink and then waits until the sink has fully flushed those values.
#[derive(Debug)]
#[must_use = "futures do nothing unless polled"]
pub struct Forward<T: Stream, U> {
    sink: Option<U>,
    stream: Option<Fuse<T>>,
    buffered: Option<T::Item>,
}


pub fn new<T, U>(stream: T, sink: U) -> Forward<T, U>
    where U: Sink<SinkItem=T::Item>,
          T: Stream,
          T::Error: From<U::SinkError>,
{
    Forward {
        sink: Some(sink),
        stream: Some(stream.fuse()),
        buffered: None,
    }
}

impl<T, U> Future for Forward<T, U>
    where U: Sink<SinkItem=T::Item>,
          T: Stream,
          T::Error: From<U::SinkError>,
{
    type Item = (T, U);
    type Error = T::Error;

    fn poll(&mut self) -> Poll<(T, U), T::Error> {
        // If we've got an item buffered already, we need to write it to the
        // sink before we can do anything else
        if let Some(item) = self.buffered.take() {
            try_ready!(self.try_start_send(item))
        }

        loop {
            match self.stream_mut()
                .expect("Attempted to poll Forward after completion")
                .poll()?
            {
                Async::Ready(Some(item)) => try_ready!(self.try_start_send(item)),
                Async::Ready(None) => {
                    try_ready!(self.sink_mut().expect("Attempted to poll Forward after completion").close());
                    return Ok(Async::Ready(self.take_result()))
                }
                Async::NotReady => {
                    try_ready!(self.sink_mut().expect("Attempted to poll Forward after completion").poll_complete());
                    return Ok(Async::NotReady)
                }
            }
        }
    }
}
```

在数据同步时首先会从自身缓存中获取数据往sink发送再从stream中获取新数据。

调用`stream.poll()`获取上游数据并发送给sink。Forward的`poll`返回结果由`stream`决定，如果`stream`未关闭则返回`Async::NotReady`，如果`stream`已关闭，且其数据已经完全消费到sink，则返回`Async::Ready`。在这个过程中如果调用sink的`start_send`返回的是`Async::NotReady`则缓存发送的数据确保在下一次继续发送，并返回`Async::NotReady`