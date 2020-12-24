[TOC]

## vector替换logstash探讨内容

0. 目的，场景
    * vector 介绍
    * logstash和vector优缺点分析
1. logstash与vector在当前AR的架构下的组件、功能替换范围，客户使用场景覆盖
    *  [logstash功能分析](http://confluence.aishu.cn/pages/viewpage.action?pageId=82073819)
    *  其他参考资料
    *  使用的logstash插件、功能或接口
        *  output
            *  kafka
                *  按分区传递
                *  性能
            *  es
                *  索引管理
                *  年月日(与logstash行为不同，存在风险)
        *  filter
            *  grok -> regex
            *  grok -> grok
            *  json 
            *  xml(缺)
            *  kv -> regex
            *  时间戳解析(type)
            *  脱敏(sed,缺)
            *  geoip(待研究适配es搜索)
            *  syslog-pi(计算公式，lua)
            *  http信息(正则)
            *  sql数据()
        *  input:
            *  流量管控插件....
            *  kafka
            *  tcp
            *  udp
            *  http
                *  http 1或2？
            *  rsyslog
            *  smptrap(待研究)
                * 特殊是数据内容
                * 特殊是数据协议
            *  filebeat(内定数据协议，还是仅仅使用socket)
                * elk全家桶，待研究
    * 改造风险点
        * 涉及的logstash组件是否能够覆盖
          * kafka读取性能，es输出性能
          * 几个插件性能和功能差异
        * 两者的稳定性，和极端情况下的表现
          * 资源不足
          * 资源抢占
          * 和es的同步行为
          * 解析规则复杂
            * 正确性
            * 性能
        * 已有规则的转换
          * logstash -> vector
    * **改造设计**内容，工作量
        * 前端页面请求的规范化
            * 重新设计与工作量评估
                * 相应功能改变
        * 后端的规则对象解析模块改造
            * 前端解析
            * **后端生成**
            * 规则对象的存储规范化
        * **规则对象的存储和分发调整策略**
            * 重新调整规范化，设计功能太多了
                * etlserver
        * **已有解析规则的转换**
        * 测试工作
            * 测试数据集准备
            * 测试用例设计
            * 测试结果分析
            * 功能性正确性测试与性能测试
2. 服务监控指标
    * 服务可观测性问题
    * [vector监控指标](###vector监控项)
    * [logstash监控指标]
    * [其他需要的指标]
3. 服务管理管理方案与接口
    * 服务管理问题
    * [ar当前控制方案](##AR\ logstash服务控制)
    * [新管理方案]
    * [风险]
4. [app概念引入(服务管理衍生)](##APP)
    * app概念
    * app涉及功能范围
    * 如何设计
    * 实现工作量
    * 使用
    * 测试
5. 测试方案
    * 开发进行前的性能和稳定性对比测试
        * 测试数据准备
        * 规则对象转换
        * 规则对象复杂场景测试
        * 相同规则对象下性能对比(极限场景与生产环境)
    * 功能验证
        * 规则对象转换
        * 数据正确性
    * 性能测试
        * 针对特定组件的性能极限测试
        * 针对组件组合的正常性能测试
        * 测试数据分析与测试用例再设计



### vector监控项
* 服务健康接口
    * 仅的返回`health: ok`，可作为服务存活探针
* 服务启动接口
* 已处理event数量
    * 含时间戳字段，已处理event数量
* 同时获取以上所有监控项
* 心跳接口
    * 上一次服务器发送负载的时间



## APP

[文档链接](app概念.md)

## logstash 监控

* logtash监控检查，端口检测
* 线程数，pipline存活，cpu、内存
* 接收速率，处理速率，输出速率(整个logstash)
* 待补充


## AR logstash服务控制
### reference

* [logstash功能分析](http://confluence.aishu.cn/pages/viewpage.action?pageId=82073819)
* [ar数据流](http://confluence.aishu.cn/pages/viewpage.action?pageId=84218780)

### 配置文件的生成

目前配置文件的生成基本上是基于etl-server(下文简称etl)接收请求参数，将参数解析为特定的logstsh配置文件，并存储到特定的etcd分组下。在etl中有三大类解析器，分别为input，filter，output。filter又分为普通filter与alterfilter解析器。

etl覆盖范围为logstash_input的input插件，logstash_filter的filter和output插件。以及logstash_archive的input，filter和output管理，其中logstash_archive的配置是固定非动态生成的。

### 配置文件的同步

llogstash_archive，logstash_input，logstash_filter的配置文件全部存放在etcd文件中。这些logstash服务通过定时监听本地配置文件的变化实现配置的变更。而本地配置文件的变更由pod中的confd服务以sidecar的方式定时拉去etcd某个分组中的内容。

### 配置文件的管理

* input插件：由前端或解析规则管理服务向etl接口发送input解析规则，触发etl中的inputworker进行解析后存储到etcd中。在获取数据源列表时，则逆解析etcd中的配置文件形成数据源列表。但是目前只有logstash_input的input得到了管理(即只发往logstash_input监听的etcd路径)，其他logstash服务实例皆为默认配置
* filter插件：由前端发送filter规则请求到解析规则管理服务，解析规则管理服务则将规则存入sql数据库中，再由解析规则管理服务将规则发给etl服务，触发etl服务生成filter解析规则，并同步到logstash_filter监听的路径(没有管理其他服务实例)。在获取解析规则列表时从数据库中获取到对应的数据对象并返回。
  * alertfilter：etl额外拆分了filter插件，专门提供了alter规则的生成，但是管理方式与filter相同。
* output插件：主要用于索引管理，管理方式同filter

### 当前已有分组

logstash已有三种分组"logstash_input","logstash_filter"和"logstash_archive"(实际不止三组，但都属于这三个范围，为了说明简单这里说是三个)，本质上就是工种的概念，但是没有有效的管理，每个工种的组成部分也没有完全进行管理，而是只管理了input,filter和output中的某些。每个分组对应了etcd下的不同路径，confd在服务启动时根据当前logstash服务实例所在分组监听对应的etcd路径文件。



## vector替换logstash方案

### vector调整

#### 监控

当前vector监控是基于一个名为`metrics_runtime`的全局、高速监控库。但可能在vector中并不适合使用这种全局的监控行为，在实际的测试中发现该目前的全局监控行为严重制约了vector的性能，并且获取到的监控指标仅有一些简单的[监控项](####vector监控项)。

目前vector的监控服务目前是通过`graphql`提供，通过目前浅显的了解，可以考虑基于`graphql`对vector的监控进行改造。将全局的监控行为改为接口获的数据处理对象组件，针对全局性的监控指标，通过聚合局部各个组件的监控指标得到。

##### 方案

原vector监控为在`metrics_runtime`的仓库中注册多个监控指标，每个组件在完成数据对象的处理后对仓库中的指标进行更新，因此多个组件间在更新监控指标时需要通过查找和获取仓库内设置的指标，在`metrics_runtime`更新指标时还需要考虑多个线程的线程安全问题。

而实际上在vector中每个组件以task的方式运行，多个task分布在不同的线程上运行，如果使用这种全局监控的指标，各个task确实需要`metrics_runtime`这样的方式进行监控的更新。但由于vector的运行框架，同一时间内不会出现相同的task在同时运行，因此目前考虑的方案为每个组件的task设置相互独立的独占的性能指标，各个task在需要更新时可以不需要考虑同步、线程安全等直接更新对应指标。

当用户需要查询某个组件的监控指标时，`graphql`根据接口直接获取对应task绑定的监控指标并返回。

当用户需要查询多个组件的监控指标或整体的监控指标时，可能会造成监控指标存在一定的误差，这个误差来源于聚合多个监控指标需要一定的时间延时，在这个过程中每个组件的监控指标会不断更新，但是可以内部管理的组件规模决定其这个延时不会很大，误差不会很大，并且通过在返回时返回每个指标的得到的时间点可以进一步降低误差。



##### vector监控项

* 服务健康接口
  * 仅的返回`health: ok`，可作为服务存活探针
* 服务启动接口
* 已处理event数量
  * 含时间戳字段，已处理event数量
* 同时获取以上所有监控项
* 心跳接口
  * 上一次服务器发送负载的时间

#### 启动与更新

当前vector启动时会自动根据配置文件启动数据流，如果启动配置项设置了定时更新配置，则vector会定期对配置文件进行读取，并与当前数据流进行比较，实现热更新。

目前的热更新方式需要定时插叙，需要额外的耗费计算资源，且存在一定延时，监控间隔小低越需要的计算量越大，并且难以对当前热更新状态进行查看，仅能通过服务日志查看。因此考虑增加vector的热更新策略，基于接口触发式，由特定控制服务进行控制，并提供接口查看状态。

#### 数据流查看与解析

* 提供接口查看当前vector实例的数据流图。
* 提供接口对配置文件进行检测，该接口可以用于对配置内容的检测
* 提供接口对配置文件进行解析，提供配置文件数据流图。

#### 新增插件

* 多行合并能力，vector中仅在文件输入中支持多行合并，为了替换logstash需要提供基于数据包的多行合并插件
* splunk UF接收插件，未来接收splunk产品转发数据
* 时间提取插件: 目前vector从某种意义上来说是提供了类型转换器，其中可以将某个字段转换为时间戳。可能需要提供类似splunk方式的时间提取方式
* 多行合并于时间提取插件：在一个插件中同时完成对时间戳的提取和多行事件划分，因为有些数据的多行划分是基于时间进行
* 脱敏插件，对敏感数据进行脱敏
* lookup，字段查表与替换，也许可以替代脱敏插件
* 字符集解码插件

#### 需要新增接口与服务调整

* 基于网络api的服务reload接口，用于支持对配置的变更
* reload状态查询接口，用于在变更配置后查看服务的重载状态
* 当前实例数据流图查看接口，用于查看实例内部的数据流图
* 配置文件检测接口，用于检查用户提交的配置是否正确，也可以用于对app安装时与当前数据流状态的冲突检测
* 解析配置文件生成数据流图接口，用于展示一套配置所代表的数据流图。
* 监控查询接口，监控查询接口分为总体服务查询接口局部组件状态查询接口
* 调整vector实例启动，服务启动时支持首先从数据库/配置中心/etl读取vector所在工种的配置文件，根据配置文件进行启动

### etl调整

拆分所有logstash包含的数据流，保留内涵数据库操作的logstash实例。对其他logstash分别以vector替换，并且赋予vector角色概念。例如原转发logstash为转发工种，原接收logstash接收为接收路由工种，原处理logstash为处理工种。同一工种共享相同配置，不同工种配置相互隔离。并且支持对各个工种下服务实例的管理，etl为各个vector和logstash实例的协调者。为了支持app概念，增加相应的资源预留接口，安装准备接口和安装执行接口和回滚。

#### logstash解析规则规则转换为vector规则在使用上的转换

##### 两种组件的拼接

###### logstash的使用方式

假设logstash的初始数据流图如下，包含对两个数据源接收不同的数据，同时有两个filter处理两个不同的数据，最终发往不同的输出方



![0fabi6.jpg](https://s1.ax1x.com/2020/10/13/0fabi6.jpg)

那么当前添加以input插件，那么input插件会默认将其下游设置问内部消息队列，新增的数据同样会经过不同的条件判断是否进入数据处理插件。里有两个核心概念，第一个filter和output线性执行，第二个每个插件的增加都会自动设定其上下游，是否会被filter插件对象处理，则基于在它之前的条件判定决定。最终数据流图如下。

那么添加一个filte插件进一步处理第一种数据，则会在上图中自动添加以一个条件判断和数据处理的组合体，其顺序会自动根据配置文件的位置自动的确定其上下游。新增filter对数据流的改变如下图所示

![0faqJK.jpg](https://s1.ax1x.com/2020/10/13/0faqJK.jpg)

###### vector的使用方式

vector和logstash相比，首先一个区别在于vector中每个插件都是独立的计算任务(在多个线程中以M:N模型允许，即M个线程允许N个计算任务)，可以认为他们是在多个线程间异步执行的。

在vector中每个组件间都会有一个消息队列来进行数据同步，而与logstash不同并非线性执行。

一个vector的初始数据流图假设如下图，其代表功能与logstash初始状态相同。

![0fa7Ix.jpg](https://s1.ax1x.com/2020/10/13/0fa7Ix.jpg)

对比logstash，我们可以发现整个数据流图更加清晰，数据间的隔离性更好，而且不需要条件判断作为数据隔离。当然如果有一种输入对应不同数据处理，那么仍然需要条件判断，这里不深入说明。

这里同样添加一个数据处理插件，用于进一步处理第一个接收源接收的数据，其数据流图代表的功能等价于logstash添加一个filter后的数据流图代表功能。

![0faTd1.jpg](https://s1.ax1x.com/2020/10/13/0faTd1.jpg)

但是需要注意的是由于logstash和vector使用方式的不同，在vector中需要人为的指定其新增插件的上下游。这也是logstash和vector在使用理念上的差异。vector将组件的上下游拼接权力交给了使用者，而logstash则内涵了一套默认的组件上下游拼接规则。

##### 使用上的转换方式

由于两者间使用上的差异，现在由logstash的解析规则转换为vector，处理基本的各个数据组件/插件的语法转换外，最大的差异在于两者在数据插件上的拼接方式差异。如果沿用logstash的规则，那么将会造成数据流仍然复杂混乱。如果使用vector的拼接方式，则需要将用户的使用习惯进行调整。

##### 当前logstash规则解析的转换

目前logstash解析分为三三种，filter、input和output

###### filter

filter解析规则有两个部分组成，第一个部分为数据是否进入filter语句块的判断条件。第二个部分为各个类型的filter插件生成。etl代码中filter插件的解析类型有15种，分别为

1. grok: 正则表达式插件
   * 在vector中也提供了grok，同时提供regex插件，后者性能较好
2. json: 指定字段的json解析
   * vector提供，需要改写规则
3. kv：kv类型的字符串内容解析
   * vector中使用正则替换
   * 有split插件，某种场景下可以支持
4. urldecode：
   * vector无，需要额外添加插件
5. str2num：字符串转数字(类型转换)
   * vector提供类型转换检查
6. geo：地理位置查询
   * vector提供相关插件
7. syslog_pri：
   * vector无
8. mutiline：多行合并插件
   * vector目前没有多行合并插件，仅在file类型的input中有，需要额外开发
9. useragent：解析浏览器请求数据的的useragent
   * vector无
10. mutate：字段重命名、重写、新增与删除以及类型转换
    * vector提供相关插件，重写使用add_fields可以完成
11. xml：xml文件解析
    * vector无
12. extractField：ruby写的字段内容替换
    * vector无，可以考虑通过lua重写
13. drop：日志丢弃
    * 通过甬道插件将drop日志传递到null输出
14. jdbc：数据库查表增加字段
    * vector不支持，仍由特定工种的logstash支持。
15. deesensitize：脱敏插件，非官方提供
    * 用extractField代替

###### input

ar在代码/默认实现中使用的input中含有：

1. AR-Agent：本质上为http和https、tcp
2. tcp/udp：socket套接字
3. http/https
4. syslog
5. smnp trap
   * vector无
6. beat
   * vector无
7. kafka
   * vector已提供
8. jdbc
   * vector目前不打算支持，仍由logstash完成jdbc数据的采集。

除了input插件的支持外，还需要提供额外的字符集设置模块，这个为vector没有的需要额外添加。

###### output

目前的代码仅涉及elasticsearch输出策略配置，涉及调整为日志库管理模块。

#### etl能力调整

etl能力为数据流各个组件控制、协调者。

##### 工种

logstash和vector新增工种概念对应到AR当前系统中的etcd分组，etl的管控以工种为单位，每个工种包含input,filter和output插件的管理。

当前ar中每个分组间分别管理不同的input,filter和output的插件，再由不同的真实分组组合为一个工种的概念。管理粒度不够完整且比较零散。

目前工种划分可以参考[logstash功能分析](http://confluence.aishu.cn/pages/viewpage.action?pageId=82073819)

* logstash保留logstash_jdb角色。其余工种考虑使用vector替换
* logstash_input使用vector替换，工种为router
* logstash_filter使用vector替换，工种为index-Extractor
* logstash_archive和logstash_shipper使用vector替换，工种为shipper，负责数据归档和数据转发

etl对工种的控制除配置文件的控制外，还包括对工种下各个实例的协调和状态管理。当数据流需要变化时由etl触发各个vector的reload接口。etl负责对各个工种的数据流图进行组装。为此定义类似如下资源接口

```
/api/etl/vector/<role>/input				## 管理特定工种的vector的input插件
/api/etl/vector/<role>/filter			## 管理特定工种的filter插件，上层应用可抽象化为告警规则，解析转发等功能
/api/etl/vector/<role>/output			## 管理特ing角色的output插件，上层应用可抽象化为日志库等功能
/api/etl/vector/<role>/overiew			## 对一个role的整体管理

/api/etl/vector<role>/input				## 管理特定工种的logstash的input插件
......

/api/etl/<modules>/overview				## 对一个模块的整体管理
/api/etl/overiew						## 对数据流的整体各个组件管理，包含logstash，vector，kafka各个工种和服务实例
```

在etcd中的分组划分按工种划分为

```
/<modulename>/<role>/input
/<modulename>/<role>/filter
......
```

因此基于当前etl的管理，首先要调整而增加**工种的管理能力**，前端如果不想进行大量调整，可以暂时在请求中加上**默认工种参数**完成。如果这也不想调整，那么在后端接口中设置为当前的工种。

##### vector替换logstash

如果使用vector替换logstash则需要前端以及其他依赖etl生成logstash配置文件的请求参数，使其满足vector的配置方式，与原先的数据流使用习惯会有较大的改变，改动涉及范围会很大，详情见[涉及功能调整](####涉及功能调整)。

最终的etl管理使用方式会变为，基于已有的数据流图，创建一个数据处理模块(input/filter/output)，并为它选择工种，以及在工种中的上下游组件，以及涉及组件的调整。

如果想基于当前logstash配置方式，通过解析器转换，无痕切换到vector，由于两种行为差异还是挺大的，还得深入考虑，甚至可能带来一些问题或留下后遗症。

##### kafka配置管理

目前kafka处于无管理状态，etl需要负责根据配置要求对kafka的topic等进行管理。例如创建topic，设置topic配置等。同样按工种进行划分。etl管理kafka实例的启动与停止暂不提供，仅提供状态查看。

##### 涉及功能调整

* 所有跟logstash生成相关的服务，其解析规则生成的重新调整，将原来基于logstash的操作变为基于vector。使用新的数据处理组件组合方式。

* 日志库：etl基于vector的elasticsearch的output插件，解析出当前日志库信息。如果需要对日志库进行调整，则改为调用vector的output管理接口生成对应的数据输出配置。日志库概念也许可以消除，反而是基于工种的输出管控。
* 数据开放：规则调整为使用vector进行，为其分配shipper角色。涉及对openlog服务并入etl服务，利用工种分配其vector配置。
* 索引库：目前主要是基于日志库的输出规则进行对elasticsearch索引的查找与分组，所以基于日志库的输出规则需要由原来的形式改为基于vector的输出，涉及获取日志库的接口调整。
* 数据归档：目前配置写死，可以直接设置静态配置，如果后续支持动态调整，则涉及如何利用etl的vecto操作配置出归档配置。
* 数据源(数据输入)：主要用于对数据的接收，和对数据的路由。可能调整为将数据接收后，按一定的策略分类并分散输出到不同的kafka实例，不同的kafka topic中进行。计划使用vector替换，工种为router，目前测试情况看来vector的socket接收能力大于logstash，需要进一步测试。主要涉及input和outpu插件的管理，也许会存在少量的filter。
* 数据流：主要涉及logstash-filter容器提供的服务，即数据解析模块的调整。但是这里vector的功能不一定能够覆盖logstash，除了需要考虑logstash规则生成调整为vector为，还需要考虑如何使用vector完成对应filter的能力，如果无法满足需要考虑扩展vector的插件。
* 本地上传：其改动类似数据流模块，需要为指定数据流进行数据组合管理或创建数据流。
* *远程采集：数据流组装习惯，也许代码不用改动，只需要在数据流管理模块设置好相应的规则后，控制远程采集往对应的kafka topic或端口发送*
* *Agent：类似远程采集*

在完成所有涉及模块的改造后，日志库、数据归档、数据开放、数据源和数据流这些功能从某种意义上变为基于工种的数据流管理，对应功能是基于使用手册的方式指导用户配置数据流。前端页面上的展示则会变为基于etl接口的前端/后端服务特定抽象。

