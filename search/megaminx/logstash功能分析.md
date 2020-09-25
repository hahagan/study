## 名词解释
* 数据处理插件: 对数据进行处理的数据处理对象，负载对数据进行清洗或某种处理。例如logstash的filter或splunk里的transform，lookup等
* event(事件): 在程序内部的数据存储的基本单元数据结构，往往代表的是一次日志事件，而不是一行日志。
* 数据源(数据接收插件): 用于接收数据或读取数据生成事件的数据处理对象，例如logstah的文件读取，tcp接收或splunk的input
* 输出源(数据输出插件): 用于向特定接收程序发送数据的数据处理对象，例如 logstash的elasticsearch输出，splunk中的output配置文件支持的几种输出定义等
* megaminx: 目前开发的字段提取，数据处理项目

## logstash整体结构
logstash主要有四种插件组成，分别为input，filter，output，codec。
其中input为数据源，用于接收客户端数据或从客户端采集数据，生成事件。
filter为数据清洗插件，用于对event进行数据处理。
output为数据输出插件，支持多种输出源，用于向第三方或es传输数据。
codec为数据解释插件，常常用于数据接收阶段的数据解释生成event和output插件的输出格式化，例如在INPUT使用codec用于结构化数据json的解析，或者多行事件的分行。codec也可以在output阶段用于将event生成特定格式的输出，例如将event编码为json文本后输出到kafka。

每个input为单独的线程运行，filter和output在同一线程内，filter_output线程数量由配置`pipline.workers`控制，官方推荐数量设置于cpu相同。input通过一个队列与filter_output线程传递数据，形成消费者-生产者模型。其数据流如下
```
 input ---push---> queue <---pull--- filter_func ---> output_func
```

## 优点与缺点
### 优点
* 通过消息队列将input和filter分离，易于横向扩展filter线程，增加数据处理的能力

### 缺点
* 横向扩展filter线程后，数据顺序无法保证
    * 解决办法是filter线程与消息队列一一对应，input线程将具有连续性的数分发给同一消息队列
* filter线程中每个filter对象，本质上是顺序执行，并且event需要经过条件判断是否需要被filter处理，多个filter对象间为顺序执行，且由于不存在隔离，对象间容易混淆，造成脏数据。

## AR数据流结构
```
                                                         + --pull-- logstash_archive
                                                         V
agent ---->(lvm option) ----> logstash-input --push--> kafka <--pull-- logstash_filter ----> elasticsearch
                                                        ^
                                                        +
                                                        +
                                  logstash_jdbc --push--+---pull--- logstah_shipper ---> other
```
* agent: 部署在客户环境，作为采集端采集日志数据发送给logstash-input
* lvm: 负载均衡，proton提供，部署在每个logstash_input所在机器，负责将接收到的数据进行负载到多个logstash_input实例上。这是因为logstash的预处理性能力在某些场景下单点无法满足，通过横向扩展logstash_input增加预处理能力。
* logstash_input: 部署在AR服务器环境，负责接收来自agent的数据，并进行预处理，随后存放到kafka中。可能存在多个。
* kafka: 消息队列，缓存logstash_input传递的数据，将数据接收和数据处理分离，主要是由于数据处理能力远低于数据接收能力，因此使用kafka进行削峰。
* elasticsearch: AR日志数据存储库
* logstah_shipper: 归档分支，用于读取kafka中的原始日志，转发到第三方，由于曾经由于logstash版本问题没有将该功能放入logstash-filter中
* logstash_archive: 从kafka读取数据后进行归档
* logstash_jdbc: 使用jdbc驱动，采集数据库数据。由于logstash版本问题没有与logstash-input合并(名字不记得叫什么了，现在是否存在未知)



## INPUT插件使用情况和能力支持
### 已用插件
logstash-input有54种
但目前AR应该经常使用的应该TCP,UDP,HTTP，用于接收agent采集的日志或filebeat。
AR中还使用了input插件中的jdbc插件用于采集数据库数据(该插件应该被修改过)，使用频率未知。
最后AR会使用input插件中的kafka插件从kafka中读取数据进行解析。

### 插件支持分析
1. 如果需要替换logstash_input，那么TCP，UDP和HTTP插件主要用于接收数据，客户方有多种数据采集方式，有些使用自研agent，有些是客户方部署的filebeat或其他工具。megaminx只支持kafka抽取，进行字段提取的场景可能过于理想化，无法支持覆盖各个客户场景。因此可能需要支持TCP，UDP和HTTP接收源。甚至如果有需要，可能会为agent专门设计一套基于网络的发送-接收方式。
2. kafka读取类型的数据源是需要支持的。
3. jdbc_static插件，需要考虑AR在什么场景下需要采集数据库，并且数据库数据作为日志保存，还需要考虑器保存策略。需要结合场景讨论是否支持

## OUTPUT插件使用情况和能力支持
### 已用插件
AR主要使用的output插件为kafka和elasticsearch。
kafka插件主要在logstash_input向kafka输出数据，elasticsearch主要用于在logstash_filter将解析的日志存储到elasticsearch。
其他使用的插件可能有TCP和HTTP，这个是因为归档支持向第三方平台转发数据，使用了这些协议。

如果使用megaminx替换logstash_input则需要考虑是否使用kafka削峰，决定是否支持kafka输出源
output的各个插件与kafka插件类似，需要结合后续需求。


## codec插件使用情况和能力支持
### 已用插件
目前查用的codec插件应该有plain，json，json_lines，mutiline
其中对于非跨行数据常用plain解释文本内容，跨行数据使用mutiline解释并生成event。
json和json_lines往往用从kafka中读取数据或向kafka传递数据，以及AR的本地上传功能。
megaminx考虑数据交换效率可能会首先支持arrow和messagepack等。

### 能力支持
对于多行和单行事件，megaminx会提供专门的类似filter模块提供多行合并能力。
对json和json_lines则视需求决定。
megaminx主要能力是基于正则的字段提取，多行事件生成数据。在满足跨行事件划分和正则提取字段的能力下，大部分文本都可以通过正则的方式解释处理，并不需要像logstash一样太多的codec插件，如果需要支持某种插件，大概率是因为对特定数据存在比正则提取更高效的解析算法或加密需求。

## filter插件使用情况和能力支持
AR在用的filter插件应该有date，grok，geoip，ruby，split，uuid，jdbc_static,jdbc_streaming(其他不清楚，需要AR补充)
* date用于对事件时间进行提取
* grok用于对字段进行正则提取
* geoip用于对地理位置进行lookup解析
* ruby应很少使用，主要是在已有插件无法满足的情况下编写ruby代码处理数据
* split应该也很少使用，用于对某个字段进行二次拆分为多个事件
* uuid应也很少使用，用于对某个字段进行id计算，往往用于指定事件ID或用于后续数据路由选择的场景
* jdbc系列插件，在数据数据接收后通过与数据库数据比对进行数据字段替换，添加等行为

### 能力支持
* date主要功能为时间提取，megaminx时间提取会和跨行事件生成一同进行
* grok，megaminx含有基于正则的字段提取
* geoip，megaminx在后续迭代中会支持splunk的lookup能力，这个能力也可以支持从某种意义上geoip的地址解析
* ruby，AR中很少使用甚至可能没有使用，如果megaminx需要支持，可能会排期到后面，并且可能支持的语言为lua或者其他高性能语言
* split，其功能其实与megminx的多行事件划分有重叠，可以考虑提取相应功能作为一个插件
* uuid，看需要，后续可支持
* jdbc系列插件，其能力类似于splunk的lookup，不过lookup table为远程的sql数据库，在logstash的实现中通过将远程sql定期或实时触发到本地local table，日志数据实际的lookup的是local table。如果需要可以以后考虑在megaminx支持lookup后支持远程数据库作为lookup table

## 总结
针对logstash的四种插件和AR常用的几种。
对于INPUT类型的数据源，megaminx可能会默认支持从kafka读取，其他类型的INPUT会是具体需求决定。对于CODEC和FILTER类型的数据解释插件，megaminx主要使用基于正则的字段提取、多行事件划分满足大部分场景。对于特定类型数据如果存在高效算法，如果有需要可以提供额外插件支持。
对于OUTPUT类型的插件则根据具体需求决定支持哪些种类。
对于CODEC的编码行为megaminx考虑数据交换效率可能会首先支持arrow和messagepack。


