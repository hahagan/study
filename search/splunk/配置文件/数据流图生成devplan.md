## 计划
### 层次
input，props，output，transform层


### 解析顺序
首先解析output和transform，生成输出对象和字段提取对象
其次解析props配置，生成数据处理对象。数据处理对象中包含transform同名对象，同时将其与transform层相连
其次解析input配置，生成数据接收对象


### output
暂时仅有名字属性

### transform
一阶段，仅有名字属性
二阶段，识别提取的字段，并作为属性的一部分
三阶段，增加描述信息

### props
生成数据处理对象，每个transform作为一个属性，并且将其与transform层对象相关联。
将数据处理对象间按执顺序串联，一阶段不搞，需要二阶段字段信息，如果无法稳定识别出提取的source，sourcetype，host，rulename几个字段值，则不串联

### input
input识别是否明确设定了source，sourcetype，host，rulename几个属性。如果可以明确source，sourcetype，host，rulename几个字段值，则指向满足的数据处理对象
