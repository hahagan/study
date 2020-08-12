## 全局配置
* 通过`[default]`定义全局配置
* 在文件头部定义全局配置
* 全局最多只有一个`[default]`语句，如果有多条，则合并其定义，有重复定义取最后一条
* 全局定义于特定对象定义冲突，取特定对象

### 定义格式
在配置文件fields.conf中添加如下格式
```conf
[<field name>]
TOKENIZER = <regula expression>
INDEXED = <boolean>
INDEXED_VALUE = [true|false|<sed-cmd>|<simple-substitution-string>]
```

* \<field name>: 字段名称
    * 疑问：如果多个数据处理或者不同host中出现同一字段名，在fields.conf中的配置岂不是会发生冲突
* TOKENIZER: 用于将字段值设置为多值，将字段的值划分为原始值的子集。例如原始字段值为"abc123"，可以划分多值"abc"和"123"做为字段值
    * 为空值时表示只有一个值
    * This setting is used by the "search" and "where" commands, the summary and XML outputs of the asynchronous search API, and by the "top", "timeline", and "stats" commands.
    * 不支持`INDEXED = true`的字段，如果存在该设置，则`TOKENIZER`配置将会被忽略
    * 无默认值
* INDEXED: 决定字段是否可以被索引，默认为false
* INDEXED_VALUE: 
    * 为 true时，值为事件的原始文本，并且将搜索时的"search=value"语句改写为"AND key=value"
    * 为false时，值不在事件的原始文本中
    * 支持替换命令，在搜索时允许在字段值上执行值替换操作，例如`INDEXED_VALUE=s/foo/bar/g`可以将字段值中的foo替换为bar
    * 支持查找替换，查找值以符号`<>`包裹，例如在"myfield=myvalue"上设置`INDEXED_VALUE=source::*<VALUE>*`那么将会查找`source::*myvalue*`作为一个term
    * For both substitution constructs, if the resulting string starts with a [', Splunk interprets the string as a Splunk LISPY expression.  For example,'INDEXED_VALUE=[OR \<VALUE> source::*\<VALUE>]' would turn 'myfield=myvalue' into applying the LISPY expression '[OR myvalue source::*myvalue]' (meaning it matches either 'myvalue' or 'source::*myvalue' terms).
    * 默认为true

## 官网原文
https://docs.splunk.com/Documentation/Splunk/8.0.5/Admin/Fieldsconf