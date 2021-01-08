### 由来

在cpython中为了线程安全的内存区读写，引入GIL锁。确保同一时间内只有一个线程在运行。因此如果需要多处理器的运行需要使用进程。

GIL锁影响了程序对多处理器的利用率。所以一些高性能的库则绕过了GIL。

但是即使性能瓶颈不在gil，仍然会造成性能下降。因为多核硬件下的系统调用开销很大。gil会导致io绑定在cpu绑定调度之前，这阻止了信号量的分发(啥意思啊，个人的理解是类似于cas中由于需要申请锁修改某个变量引发总线流量风暴，导致其他线程被阻塞)。



### python线程内幕

#### 线程特性

* 在python中的线程仍然是真实的操作系统原生提供的线程，

* 并且由操作系统进行调度但是在线程上的使用需要使用gil保证线程的安全。
* cpython使用了gil来保证线程安全

#### 线程创建

* 创建一个struct保存解释器状态
* 创建并启动一个`pthread`
* `pthread`调用 `PyEval_CallObject`
* 调用c函数

#### 线程状态（PyThreadState)

* 线程栈
* 递归深度
* 线程ID
* 线程执行信息
* hook

```
typedef struct _ts {
 struct _ts *next;
 PyInterpreterState *interp;
 struct _frame *frame;
 int recursion_depth;
 int tracing;
 int use_tracing;
 Py_tracefunc c_profilefunc;
 Py_tracefunc c_tracefunc;
 PyObject *c_profileobj;
 PyObject *c_traceobj;
 PyObject *curexc_type;
 PyObject *curexc_value;
 PyObject *curexc_traceback;
 PyObject *exc_type;
 PyObject *exc_value;
 PyObject *exc_traceback;
 PyObject *dict;
 int tick_counter;
 int gilstate_counter;
 PyObject *async_exc;
 long thread_id;
} PyThreadState;
```

这些数据结构小与100字节

#### 执行

[gil](http://www.dabeaz.com/python/GIL.pdf)

解释器持有一个当前线程状态的全局变量。`PyThreadState *_PyThreadState_Current `。因此解释器通过这个变量可以知道当前的执行线程。

gil保证了每个线程都能够独占解释器。在每次线程执行时都会申请gil，在进行IO时释放gil。

只有在IO时线程才会主动适当gil，因此对于cpu密集型的线程，为了保证gil锁能够分配到其他线程上，解释器会定期的进行检测。这个定期检查的实现依赖与一个全局计数器完全依赖于线程调度。在定期检查中会释放和申请gil，如果主线程有任何pending信号，那么主线程会执行信号处理。而这个计数器代表的是`Tick`的数量，默认为100。python解释器会将`tick`松散的插入线程执行指令中。即一个函数的执行时间会被划分为多个`Tick`。而cpython解释器则会基于已运行的`Tick`计数器确定GIL的释放与申请。

* `Tick`并不是基于时间的，而是基于解释器如何将`Tick`插入和分配到执行函数中。所以如果出现一个中断信号，主进程需要获得在gil后才能对中断信号进行处理。在发送中断信号后，解释器在每个tick后都会进行check，直到主线程获得gil。

* python的没有线程调度器，所有线程的调度完全依靠操作系统，python仅仅会对gil进行获取，通过gil实现线程安全。

* 解释器通过信号量进行加锁，获取gil，如果gil不可用则进入睡眠态，等待信号唤醒。

* 线程的调度通过信号唤醒休眠进程的开销很大。



### 使用

为了使用保证线程安全，因此使用gil保证多处理器下同时只有一个线程运行任务。对于gil的使用类似于

```
Save the thread state in a local variable.
Release the global interpreter lock.
... Do some blocking I/O operation ...
Reacquire the global interpreter lock.
Restore the thread state from the local variable.
```

这个过程类似

```
PyThreadState *_save;

_save = PyEval_SaveThread();
... Do some blocking I/O operation ...
PyEval_RestoreThread(_save);
```

都是首先将当前线程状态保存，并释放锁，此时线程可以继续执行一些阻塞的IO操作，随后当阻塞操作完成后会重新申请gil并获得原先保存的线程状态。这个过程中gil用于保护线程状态指针。



### 多核性能

```python
def count(n):
    while n > 0:
        n-=1
        

def sequetial():
    count(10000000000)
    count(10000000000)
    
def thread():
    t1 = Thread(target=count, args=(10000000000,))
    t2 = Thread(target=count, args=(10000000000,))
    t1.start()
    t2.start()
    t1.join()
    t2.join()
    
```

两核的cpu下`sequetial`耗时24.6s，`thread`耗时45.5s。单核cpu下`thread`耗时38.0s。

这个原因是python线程的执行中，每100`ticks`都需要gil进行申请和释放。gil获取需要加锁，触发信号，并且需要额外的系统调用和线程处理用于分发信号。因此在多核的情况下，需要更多的系统调用，从而增加处理时间。



### 优先级问题

低优先级cpu密集型阻塞高优先级io密集型线程，cpu密集型更容易获得gil锁。这是因为基于gil作为线程执行的依据，当io密集型线程休眠后，即使被信号量唤醒，但是无法保证其唤醒速度快于其他cpu密集型线程。