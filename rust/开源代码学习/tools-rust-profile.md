## 火焰图工具

使用`cargo-flamegraph`，平台依赖`perf`或者`dtrace`

1. 安装

   ```sh
   sudo apt-get install linux-tools-common linux-tools-generic linux-tools-`uname -r`
   cargo install flamegraph
   ```

2. 使用

   1. 火焰图中的一些函数名字可能被release优化掉，所以要补充一些debug信息，使用环境变量 `RUSTFLAGS='-g'`，或者`Cargo.toml`中添加

      ```toml
      [profile.release]
      debug = true
      ```

   2. 为非root用户启用`pref`权限，编辑`/etc/sysctl.conf`

      ```sh
      kernel.perf_event_paranoid = -1
      ```

   3. `sudo sh -c " echo 0 > /proc/sys/kernel/kptr_restrict"` 可显示系统调用的函数名，否则全是`unkown`，这个也可以写入`/etc/sysctl.conf`

   4. 启动命令

      ```sh
      # profilecarg an arbitrary executable:
      flamegraph [-o my_flamegraph.svg] /path/to/my/binary --my-arg 5
      
      # cargo support provided through the cargo-flamegraph binary!
      # defaults to profiling cargo run --release
      cargo flamegraph
      
      # by default, `--release` profile is used,
      # but you can override this:
      cargo flamegraph --dev
      
      # if you'd like to profile a specific binary:
      cargo flamegraph --bin=stress2
      
      # if you want to pass arguments as you would with cargo run:
      cargo flamegraph -- my-command --my-arg my-value -m -f
      
      # if you want to use interesting perf or dtrace options, use `-c`
      # this is handy for correlating things like branch-misses, cache-misses,
      # or anything else available via `perf list` or dtrace for your system
      cargo flamegraph -c "record -e branch-misses -c 100 --call-graph lbr -g"
      ```

3. 分析实例

   因为`perf`的采样率是`99/s`，所以每个函数的samples基本上就代表了运行的时间。



## Heaptrack

1. 安装

   ```shell
   sudo apt install heaptrack heaptrack-gui
   ```

2. 使用

   - 和火焰图一样，需要重现被优化的函数名，在cargo中写入

     ```toml
     [profile.release]
     debug = true
     ```

   - 命令`heaptrack [binary args]`，例如

     ```shell
     $ heaptrack ./rindex -D index_dir [files]
     ```

     生成`heaptrack.$binary.$PID.gz`数据，可以通过`heaptrack_gui`解析查看[强烈推荐]，也可使用`heaptrack_print`查看，或者转换为`flamegraph.pl`可识别的数据类型

   - 命令`heaptrack_gui heaptrack.$binary.$PID.gz`打开图形窗口，查看相关内存占用情况。

3. 分析实例

   下图是`rindex -c 1024`的例子，也就是处理1024MB日志数据后落盘索引文件。Consumed标签中可以看到一个断崖式下跌，就是开始写索引了。并且主要内存由三个部分组成，深橙色，是`HashMap`中的倒排表，从无到有。中橙色，从一开始就有，是`SparseSet`的预留，最后浅橙色是一些额外的短生命周期的集合类型的内存使用，占用不多，申请和释放比较频繁。鼠标放到颜色上有标签说明，到Bottom-Up中可以通过调用栈基本确定代码位置。

   这样可以得到一个结论，倒排表的内存占用的确为处理日志的10%~15%，而SparseSet的极限是128MB，在1G落盘的情况下还没有使用到一半。

   ![image-20200817135955878](https://i.loli.net/2020/08/17/D8uTxncmPAI6NFZ.png)

   ![image-20200817140029632](https://i.loli.net/2020/08/17/F1l8LJzCSZENfja.png)
   
   > 多提一句，`SparseSet`中sparse的初始化是`vec![0; 1<<24]`，源码中使用了`vec::from_elem`函数，因为0初始化，还走了快捷方式，直接申请0初始化内存，否则还要写入。这样看来heaptrack跟踪的的确是物理内存，`SparseSet`经过page-fault处理后实际占用的内存是134M左右，不然应该是256M
>
   > ```rust
> impl SpecFromElem for u8 {
   >     #[inline]
	>     fn from_elem(elem: u8, n: usize) -> Vec<u8> {
   >         if elem == 0 {
   >          return Vec { buf: RawVec::with_capacity_zeroed(n), len: n };
   >         }
   >         unsafe {
   >             let mut v = Vec::with_capacity(n);
   >             ptr::write_bytes(v.as_mut_ptr(), elem, n);
   >             v.set_len(n);
   >             v
   >         }
   >     }
   > }
   > ```




## IO tools

1. 安装，包含常用的`iostat`和`pidstat`

    ```shell
    sudo apt install sysstat
    ```

2. `iostat`常见tldr，含义忘了就`man iostat`往后翻

   ```shell
   iostat 1(秒一次) 5(次) [-d DEV 显示IO设备] [-c显示CPU信息] [-k单位kb] [-m单位mb] [-x详细信息]
   ```

   然后使用dd测单纯的读，可以到600+MB/s


