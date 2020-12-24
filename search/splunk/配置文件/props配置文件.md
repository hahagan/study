## 官网参考
https://docs.splunk.com/Documentation/Splunk/8.0.5/admin/Propsconf

## overiew
props.conf可配置能力：
* 字节编码character
* 事件多行拆分
* 事件提取
* 结构化数据头部定义于配置
* 字段提取(结合transforms.conf配置文件)
* 二进制文件配置
* 段文件配置
* 文件校验配置
* 小文件设置
* sourcetype配置
* Annotation处理器配置
* Header处理器配置
* 内部配置
* sourcetype策略和描述

## 全局配置
通过`[default]`语句块定义全局默认配置或在语句块外的文件头部定义。默认配置全局唯一，存在多个默认配置则将其合并，如默认配置冲突则取最后一项

### 定义格式
```
[<spec>]

```

### 配置说明
#### [\<spec>]
`[<spec>]`
语句块起始，`<spec>`用于为匹配该属性的事件启动某些特性，一个props中可能会有很多不同的语句块，每个语句块头部后跟随这若干配置对。如果未设置语句块头，默认为default
* `<spec>`可选值有：
    * `<sourcetype>`，通过sourcetype选择需要处理的event
        * 大小写敏感
    * `host::<host>`, 通过host字段选择event
        * 大小写不敏感，有利于dns域名识别，可通过在开头添加"(?-i)"正则使其大小写敏感
    * `source::<source>`, 通过source选择event
        * ... recurses through directories until the match is met or equivalently, matches any number of characters.
        * \*   matches anything but the path separator 0 or more times.The path separator is '/' on unix, or '\' on Windows.Intended to match a partial or complete directory or filename.
        * |   is equivalent to 'or'
        * ( ) are used to limit scope of |.
        * \\\ = matches a literal backslash '\'.
        * Example: [source::....(?<!tar.)(gz|bz2)]
    * `rule::<rulename>`, 通过rulename选择event
    * `delayedrule::<rulename>`, todo
    * todo,一个event同时设置了以上值后数据处理流程
* 语句块冲突处理
    * 注意如果存在多个语句块冲突，例如`[host::<host>]`和`[<sourcetype>]`都指向同一个event，那么它们之间会存在覆盖或者合并。覆盖规则为：
        * `[host::<host>]`配置会覆盖`[<sourcetype>]`配置
        * `[source::<source>]`会覆盖`[host::<host>]`,`[host::<host>]`
    * 对于`[source::<source>] and [host::<host>]`支持PCRE正则，但其中的"..."，"*"和"."会被转换。"."匹配句点，"*"匹配非目录分隔符,"…"匹配任意数量的任意字符。
    * 支持给定的输入匹配多个[source::<source]，并且将这些配置合并应用到该输入数据中，但如果配置项出现冲突，例如都同时配置`sourcetype`时，会根据ASCII属性呢选择配置的值，例如同时设置`sourcetype`为'a'或'z'时，此时会设置为'a'。另一种合并策略时通过设置配置项`priority`，优先级最高者为最终设置值，
    * 最终`[<spec>]`的语句块的构造是通过综合合并**字符匹配**和**模式匹配**的语句块得到，如果每个语句块中定义了优先级，则根据优先级进行合并，**字符匹配**默认优先级为100，**字符匹配**默认优先级为0
    * 语句块的优先级设置可以处理`[<sourcetype>]`或`[host::<host>]`语句块间的冲突,决定谁覆盖谁。然而不能改变不同种类的\<spec>间的覆盖行为，例如无论`[<sourcetype>]`的优先级有多高，都不会改变`[host::<host>]`配置会覆盖`[<sourcetype>]`配置的行为

#### priority
`priority = <number>`
通过内部的优先级定义，覆盖默认的语句块名的ASCII顺序，按优先级进行语句块合并

#### CHARSET
`CHARSET = <string>`
用于设置输入的字符编码，对特定输入进行指定的字符编码行为
* 语句块`[<spec>]`仅能够使用`[<sourcetype>] or [source::<spec>]`
* 在*nix类型系统中通过命令`iconv -l`能够列出支持的字符编码集合
* 无效的字符集被配置时，在初始化过程中会报出警告日志，对于输入会被丢弃
* 字符集有效，但实际编码中对一些字符无效，则这些字符将被表示为十六进制
    * 个人觉得应该将原始数据保留，在_meta中进行额外标记与特殊处理，不进行后续的数据的相同处理，也许默认不可进行搜索，但是在查找或者管理处有相关警告。如果不做额外标记难免会对数据产生污染，进行标记方便后续查看与处理
* 设置为"AUTO"时，会根据其内部算法进行识别，将会自动尝试字符编码，并转换为UTF-8
    * splunk支持添加额外的字符编码，方法为通过将样例文件存入指定位置，并重启splunk进行数据训练，存放路径为`SPLUNK_HOME/etc/ngram-models/_<language>-<encoding>.txt`
* 在input阶段生效，在读取数据时进行
    * 疑问：input阶段生效，在props.conf中进行配置？如何与input绑定，且如果多个语句块间产生冲突如何处理？合并为一个尝试列表，这岂不是会导致性能下降
        * 如果在props.conf文件中定义了该配置，是否应在props阶段对同一二进制数据进行字符编码，形成可分行分事件的内容，而不是对input阶段的字符编码进行设置。又或者根据在这里的`[<spec>]`配置生成input阶段的多种字符集编码条件，在input拿到数据后根据`<spec>`进行选择字符集再提供给后续的数据处理管道。
        * 能想象到的在input阶段确定的字段值有host，port，source，sourcetype等，而input的字符集应该是根据这些属性进行确定
* 在windows中默认为AUTO，其他系统中为UTF-8


## 事件与多行拆分
两种划分行为，一种可以通过分行符对数据流进行多行合并；另一种通过LINE_BREAKER配置进行分行分界符限定，将一块数据划分为一个事件
配置项：
* `SHOULD_LINEMERGE`: 开启多行合并功能，将划分的多行合并为一个事件，默认`true`
* `TRUNCATE=<non-negative integer>`: 默认10000字节，行最大长度，设置为0，无限制
* `LINE_BREAKER = <regular expression>`: 默认为`([\r\n])`，对原始数据的分行正则匹配，该表达式中**必须包含一个捕获组表达式**，即一对括号`()`,splunk将第一个捕获组的开始作为上一个事件的结束，并将第一个捕获组的结束作为下一个事件的开始
    * `LINE_BREAK`的正则如果由多个管道符"|"组合时，产生了多个分支，会有特殊的处理规则
        * 如果第一个匹配组为匹配的一部分，和正常的分行处理相同
        * 如果第一个匹配组不是匹配的一部分，但最左边的匹配组是正则的一部分，那么这个捕获组将会被认为是分行符
        * 如果匹配中不包含匹配组，那么会匹配位置前有一个零长度的分行符，并对行进行分割
    * 例如`LINE_BREAKER = end(\n)begin|end2(\n)begin2|begin3`
        * 如果匹配"end(\n)begin",符合第一条规则，那么会以"end"为上一行内容，"begin"为下一行内容
        * 如果匹配"end2(\n)begin2"，第一个匹配组不存在，为匹配内容一部分的匹配组，且位于最左边的匹配组为第二个匹配组，满足第二条规则，因此第二个匹配组作为分隔符，以"end2"为上一行，"beging2"为下一行
        * 如果匹配"begin3"，此时没有匹配组，因此认为"begin3"前存在一个零长度的分割符，上一行有以一个零长度的字符结尾，新行以"begin3"开始
* `LINE_BREAKER_LOOKBEHIND = <integer>`: 默认100字符，当原始数据块分行后仍有剩余数据，将会将剩余的数据连接到下一个数据块中，这里的配置指定了从原始数据块末端往前一共会将多少个字符连接到下一个数据块
* `EVENT_BREAKER_ENABLE = <boolean>`: 决定universal forwarder是否使用'ChunkedLBProcessor'数据处理程序改善指定sourcetype的分布式事件到索引器的分发
    * When set to true, a UF splits incoming data with a light-weight chunked line breaking processor ('ChunkedLBProcessor') so that data is distributed fairly evenly amongst multiple indexers.
    * When set to false, a UF uses standard load-balancing methods to send events to indexers.
    * Use this setting on a UF to indicate that data should be split on event boundaries across indexers, especially for large files.
    * This setting is only valid on universal forwarder instances.
    * Default: false
* `EVENT_BREAKER = <regular expression>`: 决定UF可以将事件传递给索引器的事件分界
    * The regular expression must contain a capturing group
    (a pair of parentheses that defines an identified sub-component
    of the match.)
    * When the UF finds a match, it considers the first capturing group
    to be the end of the previous event, and the end of the capturing group
    to be the beginning of the next event.
    * At this point, the forwarder can then change the receiving indexer
    based on these event boundaries.
    * This setting is only active if you set 'EVENT_BREAKER_ENABLE' to
    "true", only works on universal forwarders, and
    works best with multiline events.
    * Default: "\r\n"

### 事件拆分
通过`LINE_BREAK`设置事件边界，同时设置`SHOULD_LINEMERGE`为false。这种方法的好处在于可以提高存在大量多行事件的数据索引的速度。

### 分行符分行
通过`LINE_BREAK`对数据流分行，将`SHOULD_LINEMERGE`设置为true，通过`BREAK_ONLY_BEFORE, BREAK_ONLY_BEFORE_DATE, MUST_BREAK_AFTER`配置其多行合并行为

#### 多行合并的有效配置
* `BREAK_ONLY_BEFORE_DATE = [true|false]`: 默认为true，只有当发现新行中存在时间时，本行之前的行形成新事件
* `BREAK_ONLY_BEFORE = <regular expression>`: 默认为空，配置后，只有发现了新行中匹配该正则，本行之前的行生成新事件
* `MUST_BREAK_AFTER = <regular expression>`: 当配置匹配时，下一行将作为新事件，本行仍会继续匹配其他规则并进行处理
* `MUST_NOT_BREAK_AFTER = <regular expression>`:  当配置时，如果本行匹配了该正则，那么在行匹配`MUST_BREAK_AFTER`前都不会后续行不会中断当前事件。
* `MUST_NOT_BREAK_BEFORE = <regular expression>`: 当配置，并在本行匹配时，在本行结束前不会中断当前事件。
* `MAX_EVENTS = <integer>`: 默认256行，指定一个事件的最大行数，当达到这个值时将会中断事件

### ref
https://docs.splunk.com/Documentation/Splunk/8.0.5/Data/Configureeventlinebreaking

#### todo: 
splunk页面有BUG，无法正确识别配置文件，不要在页面上进行配置或查看配置
* 确定这些配置的优先级，顺序
  * `MUST_BREAK_AFTER` > `MUST_NOT_BREAK_BEFORE`
  * `MUST_NOT_BREAK_AFTER` > `BREAK_ONLY_BEFORE`
  * `MUST_NOT_BREAK_BEFORE`没发现有卵用啊
* 确定这些配置覆盖场景
  * `MUST_BREAK_AFTER`当前行匹配后，下一行作为新事件处理，适用于知道事件的结束匹配模式的场景，往往于`MUST_NOT_BREAK_AFTER`结合。
  * `MUST_NOT_BREAK_AFTER`用在当前行匹配模式后，不会形成事件，适用于已知事件的开始行和结束行模式匹配的场景。
  * `BREAK_ONLY_BEFORE`适用于知道事件的结束匹配模式的场景，将本行之前的数据形成事件。与`MUST_BREAK_AFTER`区别在于，本行数据属于新事件。并且优先级低于`MUST_NOT_BREAK_AFTER`
  * `BREAK_ONLY_BEFORE_DATE`适用于通过时间进行事件划分的模式

## 时间提取
### 定义
```
[<spec>]
DATETIME_CONFIG = <filename relative to $SPLUNK_HOME>
TIME_PREFIX = <regular expression>
MAX_TIMESTAMP_LOOKAHEAD = <integer>
TIME_FORMAT = <strptime-style format>
TZ = <POSIX time zone string>
MAX_DAYS_AGO = <integer>
MAX_DAYS_HENCE = <integer>
MAX_DIFF_SECS_AGO = <integer>
MAX_DIFF_SECS_HENCE = <integer>
```

### 配置项说明

#### DATETIME_CONFIG
`DATETIME_CONFIG = [<filename relative to $SPLUNK_HOME> | CURRENT | NONE]`
配置时间戳提取器的文件位置
* 值为NONE时保持已有的时间戳
    * 保持input阶段读取select到数据的时间
* 值为CURRENT时，将当前系统时间分配给每个事件
    * 多行合并时的时间，或者说提交给聚合程序的时间
* "NONE"和"CURRENT"都显示的禁用了每个文本的时间戳识别，因此默认的事件边界配置`BREAK_ONLY_BEFORE_DATE = true`会显得没有生效。当使用这些配置时，应该结合`SHOULD_LINEMERGE` 和/或`BREAK_ONLY_*` ,`MUST_BREAK_*`这些配置对事件合并进行控制.
* 默认为/etc/datetime.xml
* 该配置之所以被称为时间提取处理器，是因为该配置指向的文件是一组正则表达式模板，这些正则表达式指明了如何从数据中提取出各个时间戳字段。
  * 共7种时间提取模板，11种日期提取模板。并根据模板不同提取出来的字段各有差异
  * 正是基于这些默认的时间模板，splunk在时间识别一文说，如果没有特殊情况，不需要额外的配置splunk就能够识别出数据中的事件戳。

#### TIME_PREFIX
`TIME_PREFIX = <regular expression>`
默认为空，非空时，只有以匹配该正则的事件文本会的后续内容才会发生时间戳提取，如果没有匹配则不会发生时间提取

#### MAX_TIMESTAMP_LOOKAHEAD
`MAX_TIMESTAMP_LOOKAHEAD = <integer>`
生成事件前应该查找的时间戳的最大字节数量
* 该时间戳提取约束在`TIME_PREFIX`匹配的位置开始生效
* 设置为0或-1可以取消限制，但会降低性能
* 默认128

#### TIME_FORMAT
`TIME_FORMAT = <strptime-style format>`
提取时间戳的时间描述格式
* "strptime"指行业标准指定的时间戳格式(详细信息位于"Configure timestamp recognition)
* 该提取在`TIME_PREFIX`匹配后开始
* `<strptime-style format>`应能描述具体的日期与在当天的时间
* 当设置了该值，并且没有设置`TIME_PREFIX`,那么在event的头部则必须匹配该设置，否则将会引起splunk的使用无效strptime的告警
  * 这个说明意味着，如果配置了该值将会导致`DATETIME_CONFIG`配置无效
* 默认为空

#### TZ_ALIAS
`TZ_ALIAS = <key=value>[,<key=value>]...`
admin级别的事件时区解释
* 可自动根据时间形成映射
* 仅影响配置了`TIME_FORMAT`后，在其中的时区表示
* 默认不设置

#### MAX_DAYS_AGO
`MAX_DAYS_AGO = <integer>`
距当前有效的，最早事件的天数
* 当事件中的时间戳早于当前时间减去配置的天数时，那么该事件的时间戳会被设置为最近的可接收的事件的事件戳，如果没有最近可接收的事件，则设置为当前时间戳
    * 疑问： 为什么要这么设计
* 默认为2000天

#### MAX_DAYS_HENCE
`MAX_DAYS_HENCE = <integer>`
距现在有效的，未来事件的天数
* 当事件中的时间戳晚于当前时间减去配置的天数时，那么该事件的时间戳会被设置为最近的可接收的事件的事件戳，如果没有最近可接收的事件，则设置为当前时间戳
* 默认为2

#### MAX_DIFF_SECS_AGO
`MAX_DIFF_SECS_AGO = <integer>`
此设置可防止拒绝具有无序时间戳的事件
* 如果一个事件的时间戳比上一个时间戳早出该设置的值并且没有其他相同格式的时间戳文本被识别，那么将会发生告警
* 在发出告警后，如果无法给事件打上时间戳，那么该事件会被拒绝
* 默认为3600秒

#### MAX_DIFF_SECS_HENCE
`MAX_DIFF_SECS_HENCE = <integer>`
此设置可防止拒绝具有无序时间戳的事件
* 如果一个事件的时间戳比上一个时间戳晚于该设置的值并且没有其他相同格式的时间戳文本被识别，那么将会发生告警
* 在发出告警后，如果无法给事件打上时间戳，那么该事件会被拒绝
* 默认604800(一周)

#### ADD_EXTRA_TIME_FIELDS
`ADD_EXTRA_TIME_FIELDS = [none | subseconds | all | <boolean>]`
决定为事件生成和索引以下键.`date_hour, date_mday, date_minute, date_month, date_second, date_wday,date_year, date_zone, timestartpos, timeendpos, timestamp`
* none/false值，所有有关时间戳的indextime数据将会被出去。除了上面列出的字段会被删除外，也会删除subs-second级别的字段。当事件被搜索时，仅在"_time"中返回秒级时间戳
* all/true值，以上所有的时间字段会被包含
* 默认为true

## 字段提取配置
主要有三种字段提取方式"TRANSFORMS","REPORT","EXTRACT"。前两种提取在"transform.conf"文件中单独定义，在"props.conf"中创建使用连接。最后一种提取在"props.conf"文件中定义,在search time使用。
"TRANSFORMS"为index time数据处理，而"REPORT"和"EXTRACT"为search time字段提取。

### TRANSFORMS
`TRANSFORMS-<class> = <transform_stanza_name>, <transform_stanza_name2>,...`
* `<class>`为字段提取的命名空间
* `<transform_stanza_name>`对于transform.conf文件中的语句块名称
* 支持用多个逗号分割的transform名，并按顺序处理

### REPORT
`REPORT-<class> = <transform_stanza_name>, <transform_stanza_name2>,...`
* `<class>`为字段提取的命名空间
* `<transform_stanza_name>`对于transform.conf文件中的语句块名称
* 支持用多个逗号分割的transform名，并按顺序处理

### EXTRACT
`EXTRACT-<class> = [<regex>|<regex> in <src_field>]`
* `<class>`为字段提取的命名空间
* `<regex>`必须有一个命名的匹配组，当正则匹配时，匹配组名和值会被加入到事件中
    * dotall (?s) and multi-line (?m) modifiers are added in front of the regex.(?ms)<regex>
* `<regex> in <src_field>`对指定字段进行正则匹配，提取字段，否则只会匹配"_raw"字段
    * `<src_field>`只支持a-z,A-Z,0-9和下划线_
    * src_field字段必须已存在或已被其他`EXTRACT-<class>`提取，这里的`\<class>`值在ACSII排序上必须优先于本提取配置(或者说，EXTRACT的执行顺序由class值得ASCII顺序确定)
    * `<src_field>`不可以是"REPORT-\<class>"中提取得字段、自动的KV提取生成字段、字段alias、一个被计算字段或一个lookup，因为这些行为会在inline 字段提取之后进行

## 其他字段提取配置

### KV_MODE
`KV_MODE = [none|auto|auto_escaped|multi|json|xml]`
仅在search time可用，用于字段和字段值的提取
* auto: 按等号划分
* auto_escapeed: 按等号划分且将双引号内的反斜杠和双引号进行转义
* multi: 调用 "multikv"搜索命令将表格事件转换为多个事件
* json: json解析
* xml: xml解析
* 默认为auto

### MATCH_LIMIT
`MATCH_LIMIT = <integer>`
仅为EXREACT类型的字段提取配置，用于限制在PCRE正则最多调用多少次内部函数`match()`
* 可选，限制PCRE正则不匹配时的资源限制
* 默认，100000

### DEPTH_LIMIT
`DEPTH_LIMIT = <intege>`
仅为EXREACT类型的字段提取配置，用于限制在PCRE正则内部函数`match()`的嵌套栈
* 默认1000

### AUTO_KV_JSON
`AUTO_KV_JSON = <boolean>`
是否自动进行json解析提取，默认true，仅search-time可用

### KV_TRIM_SPACES
`KV_TRIM_SPACES = <boolean>`
改变当`KV_MODE`设为auto或auto_escaped时的行为
* 当值为false时，`KV_MODE`提取的字段不会去掉字段值前后的空格
* 该配置仅针对空格符，不针对tab或其他空白符
* 默认为true

### CHECK_FOR_HEADER
`CHECK_FOR_HEADER = <boolean>`
仅在index-time可用，值为true时，为文件启动基于文件头部的字段提取
* 仅个对`[<sourcetype>]`或`[source::<spec>]`有效，对`[host::<spec>]`无效
* 当`LEARN_SOURCETYPE = false`时，禁用
* 在input time应用，如果文件拥有列名列表，并且每个事件都包含一个字段值，那么会使用一个合适的文件头部行作为提取的字段名
* 默认为 false

### SEDCMD
`SEDCMD-<class> = <sed script>`
仅在index-time使用，用于处理敏感信息，应用于_raw字段
* `<sed script>`空格分隔的sed命令列表组成的sed脚本，目前仅支持替换操作(s)和**字符**替换(y)
* 一条命令的格式 `s(y)/regex/replacement/flags`，flags支持g表示全匹配替换，或者使用数字指定特定匹配

### FIELDALIAS
`FIELDALIAS-<class> = (<orig_field_name> AS|ASNEW <new_field_name>)+`
为字段设置别名，允许在搜索时使用一个或多个别名
* `<orig_field_name>`原始字段名
* `<new_field_name>`新的字段别名
* 使用`AS`时，如果新别名已存在，那么会将新别名的值使用当前的，原始字段值替换。如果原始字段无值或不存在，新别名字段将会被移除
* 使用`ASNEW`时，如果新字段别名已存在，那么不会改变。如果原始字段不存在或无值，新字段别名也不会移除
* 字段别名在search-time产生，在字段提取之后，字段执行和lookup之前

### EVAL
`EVAL-<fieldname> = <eval statement>`
通过执行`<eval statement>`生成输出结果，并将输出结果作为`<fieldname>`的值
* 多条EVAL-*语句执行，行为如同并行执行
* 在字段提取之后，lookup之前

### LOOKUP
`LOOKUP-<class> = $TRANSFORM (<match_field> (AS <match_field_in_event>)?)+ (OUTPUT|OUTPUTNEW (<output_field> (AS <output_field_in_event>)? )+ )?`
在search-time，声明一个指定的查找表，并描述如何将查找表应用到事件上
* `<match_field>` 指定查找表查找匹配的字段
    * 默认情况下，如果`<match_field_in_event>`没指定，查找与事件中同名的字段
    * 至少提供一个match_field，允许定义多个
* `<output_field>` 指定lookup项中的字段，并将其拷贝到匹配事件的`<output_field_in_event>`字段中
    * 如果未指定`<output_field_in_event>`则默认使用`<output_field>`
* 如果未指定，则对于匹配的事件，除了匹配的字段外查找表中所有字段都会被添加到事件中
* 如果使用了`OUTPUTNEW`,那么只会在事件中不存在`<output_field_in_event>`字段时，才会添加。4.1.x版本后该行为的描述是，如果已存在任意个字段，则**none**个输出字段被写入
* 如果使用了`OUTPUT`,每个输出字段都会被复写
* 执行顺序在`EVAL`之后


## 二进制文件配置
略
```conf
NO_BINARY_CHECK = <boolean>
* When set to true, Splunk software processes binary files.
* Can only be used on the basis of [<sourcetype>], or [source::<source>],
  not [host::<host>].
* Default: false (binary files are ignored).
* This setting applies at input time, when data is first read by Splunk
  software, such as on a forwarder that has configured inputs acquiring the
  data.

detect_trailing_nulls = [auto|true|false]
* When enabled, Splunk software tries to avoid reading in null bytes at
  the end of a file.
* When false, Splunk software assumes that all the bytes in the file should
  be read and indexed.
* Set this value to false for UTF-16 and other encodings (CHARSET) values
  that can have null bytes as part of the character text.
* Subtleties of 'true' vs 'auto':
  * 'true' is the historical behavior of trimming all null
           bytes when Splunk software runs on Windows.
  * 'auto' is currently a synonym for true but may be extended to be
           sensitive to the charset selected (i.e. quantized for multi-byte
           encodings, and disabled for unsafe variable-width encodings)
* This feature was introduced to work around programs which foolishly
  preallocate their log files with nulls and fill in data later.  The
  well-known case is Internet Information Server.
* This setting applies at input time, when data is first read by Splunk
  software, such as on a forwarder that has configured inputs acquiring the
  data.
* Default (on *nix machines): false
* Default (on Windows machines): true

```

## Segmentation configuration
```conf
SEGMENTATION = <segmenter>
* Specifies the segmenter from segmenters.conf to use at index time for the
  host, source, or sourcetype specified by <spec> in the stanza heading.
* Default: indexing

SEGMENTATION-<segment selection> = <segmenter>
* Specifies that Splunk Web should use the specific segmenter (from
  segmenters.conf) for the given <segment selection> choice.
* Default <segment selection> choices are: all, inner, outer, raw. For more
  information see the Admin Manual.
* Do not change the set of default <segment selection> choices, unless you
  have some overriding reason for doing so. In order for a changed set of
  <segment selection> choices to appear in Splunk Web, you need to edit
  the Splunk Web UI.
```

## Sourcetype configuration
### sourcetype
`sourcetype = <string>`
指定数据的sourcetype，并且只能够在[source::...]中配置
* 对文件类型的INPUT使用，在input-time生效
* 默认未空

### rename
search-time可用，仅对`[<sourcetype>]`语句块可设置，用于重命名sourcetype
* 如果希望搜索原始未重命名sourcetype使用字段_sourcetype
* 对重命名的sourcetype数据，仅为目标sourcetype使用search-time配置，为这个语句块的sourcetype进行的字段提取将会被忽略

### invalid_cause
`invalid_cause = <string>`
只能够在`[<sourcetype>]`中配置
* If invalid_cause is set, the Tailing code (which handles uncompressed logfiles) does not read the data, but hands it off to other components or
throws an error.
* Set <string> to "archive" to send the file to the archive processor
  (specified in unarchive_cmd).
* When set to "winevt", this causes the file to be handed off to the
  Event Log input processor.
* Set to any other string to throw an error in the splunkd.log if you are
  running Splunklogger in debug mode.
* This setting applies at input time, when data is first read by Splunk
  software, such as on a forwarder that has configured inputs acquiring the
  data.
* Default: empty string 

### force_local_processing
`force_local_processing = <boolean>`
* Forces a universal forwarder to process all data tagged with this sourcetype
  locally before forwarding it to the indexers.
* Data with this sourcetype is processed by the linebreaker,
  aggerator, and the regexreplacement processors in addition to the existing
  utf8 processor.
* Note that switching this property potentially increases the cpu
  and memory consumption of the forwarder.
* Applicable only on a universal forwarder.
* Default: false

### unarchive_cmd
`unarchive_cmd = <string>`
* Only called if invalid_cause is set to "archive".
* This field is only valid on [source::<source>] stanzas.
* <string> specifies the shell command to run to extract an archived source.
* Must be a shell command that takes input on stdin and produces output on
  stdout.
* Use _auto for Splunk software's automatic handling of archive files (tar,
  tar.gz, tgz, tbz, tbz2, zip)
* This setting applies at input time, when data is first read by Splunk
  software, such as on a forwarder that has configured inputs acquiring the
  data.
* Default: empty string

### unarchive_sourcetype
`unarchive_sourcetype = <string>`
* Sets the source type of the contents of the matching archive file. Use
  this field instead of the sourcetype field to set the source type of
  archive files that have the following extensions: gz, bz, bz2, Z.
* If this field is empty (for a matching archive file props lookup) Splunk
  software strips off the archive file's extension (.gz, bz etc) and lookup
  another stanza to attempt to determine the sourcetype.
* This setting applies at input time, when data is first read by Splunk
  software, such as on a forwarder that has configured inputs acquiring the
  data.
* Default: empty string

### LEARN_SOURCETYPE
`LEARN_SOURCETYPE = <boolean>`
* Determines whether learning of known or unknown sourcetypes is enabled.
  * For known sourcetypes, refer to LEARN_MODEL.
  * For unknown sourcetypes, refer to the rule:: and delayedrule::
    configuration (see below).
* Setting this field to false disables CHECK_FOR_HEADER as well (see above).
* This setting applies at input time, when data is first read by Splunk
  software, such as on a forwarder that has configured inputs acquiring the
  data.
* Default: true

### maxDist 
`maxDist = <integer>`
* Determines how different a source type model may be from the current file.
* The larger the 'maxDist' value, the more forgiving Splunk software is
  with differences.
  * For example, if the value is very small (for example, 10), then files
    of the specified sourcetype should not vary much.
  * A larger value indicates that files of the given source type can vary
    quite a bit.
* If you're finding that a source type model is matching too broadly, reduce
  its 'maxDist' value by about 100 and try again. If you're finding that a
  source type model is being too restrictive, increase its 'maxDist 'value by
  about 100 and try again.
* This setting applies at input time, when data is first read by Splunk
  software, such as on a forwarder that has configured inputs acquiring the
  data.
* Default: 300