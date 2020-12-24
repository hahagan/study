### 问题描述

由于logstash的输出与处理为线性处理，同步输出。此时如果es对每个bulk请求的响应时间为0.16s，这意味着每秒仅能发送约16个请求。每个请求如果发送的数据条数为logstash的batch_size的大小，默认为125。此时意味着一个logstash仅能输出约2000条/s。

在原来的logstash使用方式中，worker配置数量为4-10个，意味着这将会有4-10个es输出线程，所以性能能够达到8000-20000条/s。但是由于目前采用多实例单workers的配置，导致一个logstash仅有一个es输出线程，因此只能达到2000条/s



### 解决方法

#### 一、增大batch_size

测试发现将batch_size从125提升到1000，性能提高为3000条/s，但是提升到batch_size为2000，没有明显提升



### 二、数据输出以单独的pipeline运行，支持多workers

详细的描述是数据的处理pipeline仍然以worker为1，用于处理数据。新增一个pipeline用于支持数据输出，workers可设置为N，代表并行多个es输出。

涉及的改动：

1. 在logstash的配置涉及pipeline.yml编写，编写两个pipeline设置，两个pipeline监听不同的目录，两个目录分别为数据处理规则和数据输出规则
2. confd数据组织时将output相关数据组织到另一个目录作为数据输出规则。