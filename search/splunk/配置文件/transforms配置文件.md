## 用途
配置文件transforms.cof主要用于对数据转换的配置，常见的使用有：
* 基于正则表达式配置host和sourcetype的重写
* 处理敏感信息
* 事件到索引的路由设置
* index-time的字段提取(不建议，除非性能上的需要)
* 创建search-time的字段提取，优化以下情况：
    * 在search-time有多个相同的表达式谓语source，source type或hosts
    * 相同的source，source type或host应用不同的正则表达式
    * 使用一个正则从其他字段值，提取一个或多个值
    * 基于分隔符的字段提取
    * 同一整段的多值提取
    * 提取名字以数字或下划线开头的字段
* 设置查找外来源的查找表

以上的配置生效，需要与props.conf配置文件相对应

## 配置项

### 全局配置项
1. 通过`[default]`语句块定义全局变量或在文件头定义全局
2. 默认配置项，全局唯一，如果存在多个则合并，并取最后一个配置项的值
3. 有限取特定语句块的配置值


#### 定义
```conf
[<unique_transform_stanza_name>]
REGEX = <regular expression>
FORMAT = <string>
MATCH_LIMIT = <integer>
DEPTH_LIMIT = <integer>
CLONE_SOURCETYPE = <string>
WRITE_META = [true|false]
LOOKAHEAD = <integer>
DEST_KEY = <KEY>
DEFAULT_VALUE = <string>
......
......
```

##### 配置项详解

##### `[<unique_transform_stanza_name>]`
语句块名字，以`TRANSFORMS-`作为前缀,用于在props.conf文件中与本配置形成对应关系
* 语句块名后跟随多个配置与配置值，如果为空则使用默认值

##### `REGEX = <regular expression>`
在数据上进行的正则表达式
* index-time的转换中必填，search-time中除非使用了分隔符抽取，否则必填
* search-time中`FORMAT`为可选项，在index-time中为必填项
* 正则中的含有名字的匹配组会直接将名字作为字段名
* 如果正则提取配置没有匹配在`FORMAT`中相关定义，使用该配置的搜索将不会生成event
* 正则表达式必须有一个组，即使`FORMAT`中没有相关使用
* 如果不需要再`FORMAT`中定义格式，可以通过`_KEY_<string>, _VAL_<string>`对键和值进行配置，但仅在search-time可用
* 默认为空

##### `FORMAT = <string>`
定义结果输出格式
* 在index-time 进行转换时，可以使用`$n`指定正则匹配的输出，在正则输出结果不足n会失败
* `$0`代表在未进行正则前，`DEST_KEY`的值
* `FORMAT`的字段与字段值表示使用`::`分隔，例如`FORMAT=$1::$2,IP::$3.$4`
* index-time中默认为\<stanza-name>::$1
* 如果正则提取配置没有匹配在`FORMAT`中相关定义，使用该配置的搜索将不会生成event
* 如果`FORMAT`中定义的字段式不可索引字段，应再fields.conf中将INDEXED_VALUE设置为false，以免在搜索时发生冲突
* 不可在search-time定义串联的字段
* search-time中默认为空字符串

##### `MATCH_LIMIT = <integer>`
可选，限制PCRE正则正则不匹配时所花费的资源量，用于限制PCRE调用内部函数match()的次数上限
* 仅在REPORT和TRANSFORMS字段提取中可设置，对于EXTRACT类型的字段提取，在props.conf文件中配置
* 默认值为100000

##### `DEPTH_LIMIT = <integer>`
可选，限制PCRE正则正则不匹配时所花费的资源量，使用此选项可限制内部PCRE函数match()中backtracking的深度。
* 仅在REPORT和TRANSFORMS字段提取中可设置，对于EXTRACT类型的字段提取，在props.conf文件中配置
* 默认1000

##### `CLONE_SOURCETYPE = <string>`
使用该配置，转换会将event进行复制，并分配指定的sourcetype
* 新事件会交给props.conf里配置的规则进行处理
* `<string>` 将会设置为新事件的sourcetype

##### `LOOKAHEAD`
可选，用于指定在一个事件中查找多少个字符,默认4096
* 复杂正则在处理较大的文本时，性能可能会下降。当使用多个贪婪分支或lookaheads/lookbehind时，速度可能会以二次方下降或更糟。

##### `WRITE_META = true`
将提取字段额外写入`_meta`,除了在`DEST = _meta`情况下，该设置必填。
* 仅在index-time可用

##### `DEST_KEY`
不设置或当`WRITE_META = false`时必填,用于表示正则提取的`FORMAT`结果存放位置
* 仅在index-time可用
* 当`DEST_KEY = _meta`时，添加`$0`作为`FORMAT`的首部
* \<key>可为大小写敏感，并且必须与的`Keys`列表中的定义相同

##### `DEFAULT_VALUE = <string>`
正则失败时，写入DEST_KEY的值，默认为空，仅index-time可用

##### `source_key = <string>`
可选，设置正则的字段来源,默认为_raw
* 如果以"field:"或"fields:"开头，则将会使用已索引的字段而不是一个`KEY`

##### `REPEAT_MATCH`
可选，当`regex`多次匹配`SOURCE_KEY`时，设置为`true`。默认为`false`
* `REPEAT_MATCH`: 从上一次匹配停止开始，不断匹配，直到不不在匹配停止。有利于已知连续原始事件连续匹配的场景

##### `INGEST_EVAL = <comma-separated list of evaluator expressions>`
* index-time可用
* 可选，该配置会重写其他index-time配置，例如REGEX，并用该配置进行执行
* 该配置类似于searchtime的"|eval"命令
* 暂略

##### `DELIMS = <quoted string list>`
* search-time可用
* 暂略

##### `FIELDS = <quoted string list>`
* search-time可用
* 暂略

##### `MV_ADD = [true|false]`
* search-time可用
* 暂略

##### `CLEAN_KEYS = [true|false]`
* search-time可用
* 暂略

##### `KEEP_EMPTY_VALS = [true|false]`
* search-time可用
* 暂略

##### `CAN_OPTIMIZE = [true|false]`
* search-time可用
* 暂略


### lookup
search-time可用，暂略

### Metrics
#### statsd定义
```conf
[statsd-dims:<unique_transforms_stanza_name>]
REGEX = <regular expression>
REMOVE_DIMS_FROM_METRIC_NAME = <boolean>
```

#### statsd配置项详解
* `[statsd-dims:<unique_transforms_stanza_name>]`: 
    * 'statsd-dims' prefix indicates this stanza is applicable only to statsd metric
    type input data.
    * This stanza is used to define regular expression to match and extract
    dimensions out of statsd dotted name segments.
    * By default, only the unmatched segments of the statsd dotted name segment
    become the metric_name.
* `REGEX = <regular expression>`: 通过正则和正则组名提取维度名，例如`(?<dim1>group)(?<dim2>group)..`
* `REMOVE_DIMS_FROM_METRIC_NAME = <boolean>`: 
    * 为false时，正则匹配的值同样作为metric的名字的一部分呢
    * 为true时，正则匹配的值不作为作为metric的名字的一部分
    * 默认为true

#### metric-schema定义
```conf
[metric-schema:<unique_transforms_stanza_name>]
METRIC-SCHEMA-MEASURES-<unique_metric_name_prefix> = (_ALLNUMS_ | (_NUMS_EXCEPT_ )? <field1>, <field2>,... )
METRIC-SCHEMA-BLACKLIST-DIMS-<unique_metric_name_prefix> = <dimension_field1>,
<dimension_field2>,...
METRIC-SCHEMA-WHITELIST-DIMS-<unique_metric_name_prefix> = <dimension_field1>,
<dimension_field2>,...
METRIC-SCHEMA-MEASURES = (_ALLNUMS_ | (_NUMS_EXCEPT_ )? <field1>, <field2>,... )
METRIC-SCHEMA-BLACKLIST-DIMS = <dimension_field1>, <dimension_field2>,...
METRIC-SCHEMA-WHITELIST-DIMS = <dimension_field1>, <dimension_field2>,...
```

#### metric-schema配置详解
暂略

## KEYS列表
```
* NOTE: Keys are case-sensitive. Use the following keys exactly as they
        appear.

queue : Specify which queue to send the event to (can be nullQueue, indexQueue).
        * indexQueue is the usual destination for events going through the
          transform-handling processor.
        * nullQueue is a destination which causes the events to be
          dropped entirely.
_raw  : The raw text of the event.
_meta : A space-separated list of metadata for an event.
_time : The timestamp of the event, in seconds since 1/1/1970 UTC.

MetaData:Host       : The host associated with the event.
                      The value must be prefixed by "host::"

_MetaData:Index     : The index where the event should be stored.

MetaData:Source     : The source associated with the event.
                      The value must be prefixed by "source::"

MetaData:Sourcetype : The source type of the event.
                      The value must be prefixed by "sourcetype::"

_TCP_ROUTING        : Comma separated list of tcpout group names (from
                      outputs.conf)
					  Defaults to groups present in 'defaultGroup' for [tcpout].

_SYSLOG_ROUTING     : Comma separated list of syslog-stanza  names (from
                      outputs.conf)
					  Defaults to groups present in 'defaultGroup' for [syslog].

* NOTE: Any KEY (field name) prefixed by '_' is not indexed by Splunk software,   in general.

[accepted_keys]

<name> = <key>

* Modifies the list of valid SOURCE_KEY and DEST_KEY values. Splunk software
  checks the SOURCE_KEY and DEST_KEY values in your transforms against this
  list when it performs index-time field transformations.
* Add entries to [accepted_keys] to provide valid keys for specific
  environments, apps, or similar domains.
* The 'name' element disambiguates entries, similar to -class entries in
  props.conf.
* The 'name' element can be anything you choose, including a description of
  the purpose of the key.
* The entire stanza defaults to not being present, causing all keys not
  documented just above to be flagged.
```
    
