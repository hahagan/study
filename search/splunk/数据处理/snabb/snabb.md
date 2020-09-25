## 概述
snabb app从input端口接收数据包，经过处理在output端口传输数据包。每个app可以拥有0或多个input和output端口

### app重要方法与参数
1. `new`: 创建app实例
* input，output: app的输入和输出
* appname: app名字
* shm: 共享内存大小
2. `link`: 在app的links改变时，被调用，确保在新link调用`pull和push`前执行
3. `pull`: 将数据包pull到网络中，例如将从网络适配器中的数据包通过`trnasmit`传递给output进入app的网络
4. `push`: 将数据包推入系统中，例如将数据包从input移动到output或网络适配器中
5. `reconfig`: 重配置app
6. `report`: 打印app状态
7. `stop`: 停止app，并释放资源
8. `zone`: luajit的profiling zone，默认为模块名
```lua
-- Function config.app config, name, class, arg
-- Adds an app of class with arg to the config where it will be assigned to name.
config.app(c, "nic", Intel82599, {pciaddr = "0000:00:00.0"})
```

#### app数据结构
类似于
* app.output: table 类型，属性名为link

### 配置(core.config)
可以用于描述数据包处理网络。网络可用有向图描述，图中的节点为每个数据包的处理app，边单向线，从output指向input
配置需要添加节点app，随后通过添加app间output和input间的连接将数据包连接起来
```lua
local config = require("core.config")

local c = config.new()
...
```

### 引擎(core.app)

#### 配置更新
通过配置初始化app，创建连接，并处理数据流。同时负责profiling和repoter。
支持热更新，热更新策略如下：
* app在旧配置不存在则启动
* app在新配置不存在则停止
* app没有改变则存留
* app配置改变则调用app的`reconfig`，如果`reconfig`未实现则停止旧实例创建新的实例并运行
* 连接未改变的app节点

#### 配置处理的过程
1. 将不存在于新配置中的link进行unlink操作,并最终完成free_link释放link
    * 其中主要包含link两端app的output和input中对应link移除，并释放link。(unlink代码有点令人疑惑，为何需要`remove_link_from_array`这一函数来进行移除，遍历的性能低于直接通过名字查找?或者是存在一个output多个相同的link，通过这样的方式避免后续的删除操作带来的损耗?)
2. 启动已有配置中不存在的app
3. 对于app同名但不同类按先停止后启动方式更新app
    * 停止app，首先调用app方法，停止app运行，同时处理app的shm，最后将全局变量app_tables和configuration中的app相关信息去除
    * 启动app，创建app实例，创建shm，并在全局变量中添加app相关信息
4. 对于app已存在，但配置不同，如果app提供了reconfig方法进行热更新则调用进行热更新，否则停止再启动进行更新
5. 重建link，创建不存在的link，并为新的link或被重建的app进行link挂载。由于link唯一标识由其两端决定，所以不存在与旧配置相同标识但参数不同的link。
6. 如果app具有pull方法则，将其加入breathe_pull_order，并排序。
7. 将app.input排序并插入breathe_push_order中

##### 引擎识别配置相关源码
```lua
-- https://github.com/snabbco/snabb/blob/master/src/core/app.lua
-- function apply_config_actions 
-- compute_config_actions 
-- compute_breathe_order 
    local function remove_link_from_array(array, link)
      for i=1,#array do
         if array[i] == link then
            table.remove(array, i)
            return
         end
      end
    end

    function ops.unlink_output (appname, linkname)
        local app = app_table[appname]
        local link = app.output[linkname]
        app.output[linkname] = nil
        remove_link_from_array(app.output, link)
        if app.link then app:link() end
    end

    function ops.free_link (linkspec)
      link.free(link_table[linkspec], linkspec)
      link_table[linkspec] = nil
      configuration.links[linkspec] = nil
   end

   function ops.link_output (appname, linkname, linkspec)
      local app = app_table[appname]
      local link = assert(link_table[linkspec])
      app.output[linkname] = link
      table.insert(app.output, link)
      if app.link then app:link() end
   end
```

引擎的执行：支持间隔执行或者循环执行，此过程中支持repoter的行为选择

### 链接(core.link)
链接在apps间存储数据包的`ring buffer`,可以如同数组或者数据流对待link。

### packet
packet用于存储当前正在处理的数据。
1. 每个packet必须有明确的生命周期。
2. 数据包通过两个接口明确的分配和释放。
3. 通过`link.receive`接口可以获取数据包的所有权。
4. app必须确保通过`link.trnasmit`将数据包将数据所有权传递给其他app或通过`free`释放数据包
5. app仅能在数据包没有被`transmit`或`free`前使用
6. 数据包分配，从一个数据包池中进行分配
数据结构
```c++
struct packet {
    uint16_t length;
    uint8_t  data[packet.max_payload];
};
```

### Memory
snabb会分配网络驱动能够直接访问的DMA内存。该内存会以一个稳定的连续的物理内存地址开始。
对应接口支持申请DMA大小，查看物理内存地址和单个huge page大小。
packet数据会保存在DMA中

### 共享内存
packet数据保存在DMA中，但是shm中会保存packet的指针信息，以及可用的packet指针队列
问题：snabb什么情况下会出现多个进程，并且需要在进程间进行内存共享

### counter
双精度的共享内存计数器，与共享内存挂接，类型counter

### Histogram 
共享内存直方图，与共享内存挂接，类型未histogram

### lib库
一些工具

### 多进程操作
通过调用core.worker模块(仅主线程可用)可以创建子进程，用于多进程写作行为。每一个worker都是一个完整的snabb进程，可以定义网络，运行引擎以及其他snabb行为。每个worker的具体行为由创建时提供的lua表达式决定。
并且snabb进程组有以下特性：
* 组级终止: 主进程停止时，所有worker会自动终止，包括`kill -9`级别
* 共享DMA内存: 可以通过shm共享内存对象，该内存自动映射
* PCI设备终止: 对于组内每个进程打开的PCI设备，总线控制(DMA)在终止时会被禁用直到DMA内存返回到内核中。这个特性可以保证已释放且重复使用的内存在中断时引起DMA悬挂
worker模块通过`fork()+execve())`执行lua代码



## snabbnfv配置

### 配置加载nfvconfig.lua
代码内容大致如下
```lua
-- 首先读取配置文件，添加一些port配置网络设备对应的app，例如RawSocket，Synth等，并返回app间的link配置信息
-- io_links 结构为io_links[i] = {input = NIC.."."..device.rx, output = NIC.."."..device.tx}
io_links = virtual_ether_mux.configure(c, ports, {pci = pciaddr})

-- 随后为每个port创建虚拟IO的app，随后变量t代表一个port配置
config.app(c, Virtio, VhostUser,
                 {socket_path=sockpath:format(t.port_id),
                  disable_mrg_rxbuf=t.disable_mrg_rxbuf,
                  disable_indirect_desc=t.disable_indirect_desc})

-- 此时如果port的配置中限制了输出(tx_police)的限速，则通过RateLimiter机械能限制
config.app(c, RxLimit, RateLimiter, {rate = rate, bucket_capacity = rate})
config.link(c, RxLimit..".output -> "..VM_rx)

-- 随后按顺序增加port的ingress_filter,egress_filter,tunnel,crypto,rx_police的app配置和link配置
```

一个snabbnvc的配置中的一个`port`配置解析后生成的各个app协作图如下，其中虚拟以太往设备有多种，在通常指向一个网络适配器。并且多个port的虚拟一台网设备本质上指向同一个网络适配器。
![snabb-traific网络图](images/snabb-traific网络图.png)


## 可借鉴思路
1. 在link中使用`ring buffer`(如同循环数组)作为app间的数据缓存，有利于数据流的缓冲，可以减少数据的移动
2. 数据包缓存的内存需要从全局内存池中分配，有利于在全局上进行内存控制
