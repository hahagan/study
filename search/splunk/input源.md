## inputs配置方法
* 各种splunk支持的app
* splunk web页面
    * 使用splunk web添加的input会在**$SPLUNK_HOME/etc/apps/search/local/inputs.conf**中添加input定义
* splunk命令行工具
* 配置文件中的inputs.conf
    * 可配置目录 $SPLUNK_HOME/etc/system/local/
    * 自定义应用配置目录 $SPLUNK_HOME/etc/apps/<app name>/local

## input内容

### 定义全局配置
可以使用[**default**]作为全局默认配置，也可以在文件头部进行全局配置，但是全局配置只能有一个

### 通用配置项
* input类型，定义方法为`[inputType]`,例如`[tcp://5140]`
* host = <string>, 可设置为静态变量，也可以设置为特殊值
    * 特殊值为 $decideOnStartup, 将会设置host字段为执行机器的主机名
    * 如果去除了host配置项，对应host配置项默认被设置为 $decideOnStartup
* index = <string>, 设置用于存储该input的索引，默认为main
* source = <string>, 设置event的source字段值
    * source字段将会用于在parsing/indexing阶段
    * 默认为input文件的路径
* sourcetype = <string>, 设置sourcetype字段
    * 这个字段在parsing和indexing阶段被用于设置 sources type 字段
    * 没有默认值，如果没有设置，由索引解析数据并选择类型
* queue = [parsingQueue|indexQueue], 选择处理队列
    * parsingQueue 将会被props.conf文件定义的方法或其他解析规则处理
    * indexQueue 直接进入索引
    * 默认为parsingQueue
* _TCP_ROUTING = <tcpout_group_name>,<tcpout_group_name>,<tcpout_group_name>, ...
    * This setting lets you selectively forward data to specific indexer(s).
    * To forward data from the "_internal" index, you must explicitly set '_TCP_ROUTING' to either "*" or a specific splunktcp target group.
    * To forward data to all tcpout group names that have been defined in outputs.conf, set to '*' (asterisk).
* _SYSLOG_ROUTING = <syslog_group_name>,<syslog_group_name>,<syslog_group_name>, ...
* _INDEX_AND_FORWARD_ROUTING = <string>
    * Only has effect if you use the 'selectiveIndexing' feature in outputs.conf.


## 各种inputType定义



