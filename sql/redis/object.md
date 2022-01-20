# 对象

* database：数据库，库名最长64字符	
* collection：文档集合，对应数据库的表。
  * 集合名以下划线或字母开头，不可为空，不可以"system."作为开头，不可含有"$"，不可含有空字符。
  * [capped collection](https://docs.mongodb.com/manual/core/capped-collections/)类似回环数组一样固定大小的collection
  * [time series coleection](https://docs.mongodb.com/manual/core/timeseries-collections/)
* document：文档，对应数据库的记录/行
  * mongodb中的document即使在同一collection中也可以有不同的schema
  * mongoDB自3.3后支持对document进行[schema限定与识别](https://docs.mongodb.com/manual/core/schema-validation/)
  * document以一种类似json的二进制形式表示，被称为bson
  * 一个document最大为16MB
  * 内部字段是[有序](https://docs.mongodb.com/manual/reference/bson-types/)的
* filed：document中的字段，对应为数据库中的属性/列
  * 对嵌套的字段/数组可以通过"."符号访问，具体格式为`<array>.<index>`或`<fieldName0>.<fieldName1>`
  * 更新文档字段名称会导致文档字段重排序
  * `_id`字段为对象ID字段，`name`字段为不可为null字符
* [view](https://docs.mongodb.com/manual/core/views/)：视图，对应数据库视图，但仅为逻辑视图，非物理视图
  * view由[聚合管道](https://docs.mongodb.com/manual/core/aggregation-pipeline/#std-label-aggregation-pipeline)进行定义和计算
  * view的计算为查询时进行的lazy计算，当用户查询时会将查询至于对应pipline下
  * view为只读的
  * 通过[merge](https://docs.mongodb.com/manual/reference/operator/aggregation/merge/#mongodb-pipeline-pipe.-merge)和管道可以按实际需要使得视图变为[物化视图](https://docs.mongodb.com/manual/core/materialized-views/)

