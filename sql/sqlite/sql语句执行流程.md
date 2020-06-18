#概述
[sqlite3_preare_v2](https://www.sqlite.org/c3ref/prepare.html)和相关接口演示了一个编译器如何将sql文本转换为字节码<br/>
[sqlite3_stmt]()是实现单条sql语句的单个字节码程序的容器
[sqlite3_step](https://www.sqlite.org/c3ref/step.html)接口将一个字节码传递给虚拟机，并运行字节码程序直到结束或有结果可返回、发生错误、中断

## sql语句转换就绪陈述对象与执行
每个sql语句被认为是独立的计算程序，原始的sql文本作为源码。每条就绪的陈述对象都是编译好的对象码。所有的sql语句在运行前必须被转换为就绪陈述对象。
转换陈述对象流程：
1. 使用[sqlite3_prepare_v2](https://www.sqlite.org/c3ref/prepare.html)创建陈述对象
2. 绑定参数
3. 多次调用[sqlite3_step]()运行SQL语句
4. 使用[sqlite3_reset]()重设陈述对象，跳转到第二步。重复多次
5. 使用[sqlite3_finalize]()销毁对象

## 执行实例流程
![执行sql](images/3-1.jpg)
