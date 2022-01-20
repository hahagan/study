## task概念

tikio中计算执行单元。类似于线程，但task的调度由tikio的runtime控制，且task是不可抢占的。

task必须是轻量级、协作式(不可抢占)和非阻塞的。

## task创建

创建task有四种方式，分别为`spwan`,`spwan_blocking`,`block_in_place`和`yield_now`

### spawn

`tikio::spawn`函数可以接收一个`async`代码块或一个`future`对象，并返回一个`JoinHandle`。`JoinHandle`为一个`future`结构，用于等待获取结果。通过`spawn`创建的任务不应该是阻塞型任务，否则将会导致线程阻塞

```rust
use tokio::task;

task::spawn(async {
    // perform some work here...
});
```

启用feature: `spawn`, `JoinHandle`, and `JoinError` are present when the "rt-core" feature flag is enabled.

### spwan_blocking

与`spwan`类似，区别在于`spwan_blocking`用于在一个线程中运行阻塞型任务。

### block_in_place

该方法与`spawn_blocking`类似，主要用于执行阻塞任务，区别在于该方法将会阻塞当前worker线程，并将该线程上的其他任务移动到其他线程。该方法有利于减少上下文切换带来的开销

### yield_now

`yield_now`方法用于将当前任务让出线程，允许线程执行其他任务



## 源码阅读

### spawn

以下代码为`spawn`方法主要涉及代码和数据结构。`spawn`会根据运行时使用的features决定使用那种`Spawner`对象的`spawn`方法，生成`task`，随后将其通过`self.shared.schedule(task);`将任务调度存入`Shared`对象中。

```rust

/// Handle to the runtime.
///
/// The handle is internally reference-counted and can be freely cloned. A handle can be
/// obtained using the [`Runtime::handle`] method.
///
/// [`Runtime::handle`]: crate::runtime::Runtime::handle()
#[derive(Debug, Clone)]
pub struct Handle {
    pub(super) spawner: Spawner,

    /// Handles to the I/O drivers
    pub(super) io_handle: io::Handle,

    /// Handles to the time drivers
    pub(super) time_handle: time::Handle,

    /// Source of `Instant::now()`
    pub(super) clock: time::Clock,

    /// Blocking pool spawner
    pub(super) blocking_spawner: blocking::Spawner,
}

#[derive(Debug, Clone)]
pub(crate) enum Spawner {
    Shell,
    #[cfg(feature = "rt-core")]
    Basic(basic_scheduler::Spawner),
    #[cfg(feature = "rt-threaded")]
    ThreadPool(thread_pool::Spawner),
}


/// Submit futures to the associated thread pool for execution.
///
/// A `Spawner` instance is a handle to a single thread pool that allows the owner
/// of the handle to spawn futures onto the thread pool.
///
/// The `Spawner` handle is *only* used for spawning new futures. It does not
/// impact the lifecycle of the thread pool in any way. The thread pool may
/// shutdown while there are outstanding `Spawner` instances.
///
/// `Spawner` instances are obtained by calling [`ThreadPool::spawner`].
///
/// [`ThreadPool::spawner`]: method@ThreadPool::spawner
#[derive(Clone)]
pub(crate) struct Spawner {
    shared: Arc<worker::Shared>,
}

struct Tasks {
    /// Collection of all active tasks spawned onto this executor.
    owned: LinkedList<Task<Arc<Shared>>>,

    /// Local run queue.
    ///
    /// Tasks notified from the current thread are pushed into this queue.
    queue: VecDeque<task::Notified<Arc<Shared>>>,
}
```

```rust
// ==== impl Spawner =====
impl Spawner {
    /// Spawns a future onto the thread pool
    pub(crate) fn spawn<F>(&self, future: F) -> JoinHandle<F::Output>
    where
        F: Future + Send + 'static,
        F::Output: Send + 'static,
    {
        let (task, handle) = task::joinable(future);
        self.shared.schedule(task, false);
        handle
    }
}

impl Shared {
    pub(super) fn schedule(&self, task: Notified, is_yield: bool) {
        CURRENT.with(|maybe_cx| {
            if let Some(cx) = maybe_cx {
                // Make sure the task is part of the **current** scheduler.
                if self.ptr_eq(&cx.worker.shared) {
                    // And the current thread still holds a core
                    if let Some(core) = cx.core.borrow_mut().as_mut() {
                        self.schedule_local(core, task, is_yield);
                        return;
                    }
                }
            }

            // Otherwise, use the inject queue
            self.inject.push(task);
            self.notify_parked();
        });
    }
    ......
    ......
}


```







