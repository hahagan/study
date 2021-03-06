### app的能力范围

#### 涉及的已有功能

* 搜索
  * 已存搜索语句
  * 日志分组
  * 关系图谱
  * 仪表盘与可视化
  * 告警
* 索引管理(内部概念有点模糊)
  * 日志库
  * 索引库
* 数据源(替换logstash与调整解析规则管理)
  * 数据输入
  * 数据开放
  * 数据流(......)
  * 本地上传
    * 设置规则来源
  * 归档
  * 远程采集(python写的数据采集)
  * 日志库管理
  * 数据库连接(不替换vector，仅统一管理配置)
* Agent管理
  * 服务端的配置读取和客户端安装部署
* 机器学习
* 标签管理
* 巡检任务

#### 扩展功能

* 服务实例管理(服务的伸缩，服务的部署、删除等)
  * 特权服务提供接口
* 服务监控管理
  * 通用监控指标走平台监控
  * 特殊指标提供要求服务根据规范实现监控接口，通过prometheus或某些采集程序进行采集。最终的可视化通过产品内部功能或grafana绘图完成。
* app支持用户的编写安装代码
  * 不支持该行为，如果需要扩展第三方功能，通过app内的k8s对象完成对应的第三方安装。内部初始化由特权服务完成，避免代理预料之外的破坏。如果用户需要调用产品接口完成初始化，同样通过k8s对象完成。
* 框架支持按某种规则执行app内附带的程序代码扩展系统能力
  * 可以考虑通过用例指导用户通过k8s对象结合产品第三方开放接口和app扩展产品能力
  * 不像splunk一样直接执行app代码，是考虑到安全问题，如果出现安全问题可能会导致服务崩溃。
* 通过编码支持用户自定义spl命令
  * 待定，等等spl框架完善再考虑是否进行
* app所有权



### app配置文件规划

#### 配置文件规划

```
|----app.conf
|----service
	|---deployment
	|---statefulset
	|---daemonset
	|---job
	|---volumes
|----default
	|----datastream
		|----logstash
			|----<role>
				|----rule.conf
		|----vector
			|----<instance>
				|----rule.conf
		|----<other_service_name>
			|----<instance>
				|----some.conf
	|----search
		|----savedsearch.conf
		|----loggroup.conf
		|----view.conf
		|----dashboard.conf
		|----relationship.conf
	|----alert.conf
	|----sqlconnector.conf
	|----agent.conf
	|----label.conf
	|----presearch.conf
	|----index.conf
|----local
	|----datastream
		|----logstash
			|----<role>
				|----rule.conf
		|----vector
			|----<instance>
				|----rule.conf
		|----<other_service_name>
			|----<instance>
				|----some.conf
	|----search
		|----savedsearch.conf
		|----loggroup.conf
		|----view.conf
		|----dashboard.conf
		|----relationship.conf
	|----alert.conf
	|----sqlconnector.conf
	|----agent
	|----label.conf
	|----presearch
	|----index.conf
|----metadata
|----readme
	
	
```

* `local`与`default`：local和default目录包含了app的代表的功能配置，这些功能主要是产品内部的能力。与sp类似default目录代表app默认带有的配置项，而local为用户调整或被其他app调整后的配置。local具有较高优先级
  * `datastream`：目录为数据流配置，用于控制数据流。包含数据的接收，数据的处理和数据的输出配置。由于数据流可能有不同的组件组成，因此该目录下一级目录为各种服务组件的名称，用以应对不同组件的不同配置内容、安装方法等。
    * `<role>`：二级目录`<role>`名称对应一个组件服务应该扮演的角色，目录下各个配置文件代表了这个角色应该使用的配置。这么设置的原因是数据流由多种组件组合而成时，以不同的角色组成。当同一种组件存在多种角色是，相同角色的实例可以通过`<`role>`决定实例应该监听/分发的配置。例如AR的logstash目前有多个实例，每个实例都扮演了不同的角色(接收，处理，转发等)，但是没有统一管理，通过该方法可以将不同的行为统一管理。
  * `search`：目录下保存各种search-time相关的配置，例如已存搜索，日志分组，仪表盘等
  * `alter.conf`：内含告警策略，可以考虑是否合并到`search`目录下
  * `sqlconnector`：数据库连接信息，可以考虑进行加密处理
  * `agent`：相关管理
  * `label`：标签管理，考虑是否不要这个功能
  * `presearch`：在splunk中如果search命令在输出结果前会首先经过一次数据流的预处理，根据数据流的配置对索引中的数据进行处理。这里为后续需要进行与处理了的search功能预留
  * `index.conf`：索引相关配置，主要包含索引的管理项
* `service`：目录下划分多层，分别对应k8s的各种对象，用于扩展或启用服务实例。有特权服务进行部署。其等效于splunk中的`bin`、`appserver`、`lib`、`static`等目录的结合体。
  * 注意如果service中包含前端页面则规定其路由为`/app/<appname>/web/<webname>`，如果包含后端服务则规定其api路由为`/app/<appname>/api/<servicename>/....`和`/app/<appname>/openapi/<servicename>/....`
  * k8s对象要求：label自带app名称；label自带服务名称；namespace接收参数；有状态应用存储路径接收参数设置；自带符合规范的路由；
* `app.conf`文件则包含了主要的app定义
* `metadata`：包含app自身的一些元信息，含权限等
* `readme`：app相关文档



### app作用域与权限

#### 作用域和所有权

每个app功能需要额外定义app的所有权，app功能的生命周期为app的生命周期。

需要提供为app设置权限的能力，从而决定各个功能在系统中的可见范围和作用域。权限可以考虑基于角色和用户进行划分，为每个app中的功能设定权限。权限主要分为查看权限，修改权限和任务执行者。

#### 默认作用域

app的search-time(从索引中查询数据阶段)作用域，默认作用于app自身。即不同app间的预置字段提取(当前产品没有该功能，但需要为后续的schema on read 提供可能需要的隔离考虑)不会相互影响。尽可能保持保持app间能力的相互独立，因为一个app本身就代表了一个完善的能力集合。每个app存在自己的命名空间，如果特殊需要app间需要梦幻联动，需要获得app的权限进行访问。

app在index-time(数据接收、处理阶段)上作用域全局共享可见(或部分功能全局共享可见)，原因是多个app间可能会共享同一个数据输入(希望不共享，但是不显示)，app在index-time的数据接收能力无法做到完全独立。例如两个app的接收端口都为5140端口，除了端口资源冲突外，如果两个app间配置不可见，那么就会造成数据处理规则的冲突并且难以协调。

#### app功能的可见性和执行权限

为了使得app提供的功能可以给其他app或使用者提供，那么app的功能就可以通过各个功能对象的权限配置来控制可见性。在权限设置中可以设置功能的读写权限和执行该功能的执行者。例如某个app的功能"已存搜索1"

| 已存搜索1        | access | write | 执行权限（执行者） |
| ---------------- | ------ | ----- | ------------------ |
| admin(role)      | true   | true  | admin(role)        |
| app1（other app) | true   | false | app's role         |
| game（user)      | true   | false | user's role        |



为了达成app作用域和权限的划分，系统前端页面需要有所调整，后端可能需要考虑各种配置的存储结构。app间可共享的功能通过功能对象的权限设置，决定其是否可被其他应用使用。**以上除了所有权概念必须增加，其他都不是必须**。可以用额外的数据库表来表示对应配置的所有者，在删除和安装时根据配置所有者进行添加和删除。



### app的生成与安装

app实际上是代表了一系列能力的配置文件。程序代表了实际的能力集合，程序安装app即根据app配置触发两种行为：

* 更新数据库/缓存/内存数据

* 创建或刷新task/pod

那么app安装时需要支持以上两种行为程序和配置文件的一致性，并处理已有程序状态和配置文件的冲突。

#### 整体流程

1. 根据[配置文件规划](###app配置文件规划)，安装模块调度生成不同的安装任务并存入安装任务列表，以datastream为例。
   1. datastream安装模块，目录下各个目录，并生成对应服务的配置项的子任务，以vector/logstash为例。
   2. 读取vector/logstash目录下各个文件，合并为同一文本内容，并生成调度vector的子任务
2. 执行安装任务的各个阶段
   1. 执行安装任务的第一阶段，安装任务第一阶段承担对配置的合法性检查和资源检查，可以称为**资源预留任务**，该阶段可以成为**资源预留阶段**。例如datastream的子安装任务第一阶段向etl_server(vector管理服务，简称etl)发送资源资源预留请求。
      * etl根据结合请求配置与资源状态，首先对配置进行语法检查，随后进行资源冲突检查。
        1. 如果请求资源与预留资源项冲突，则返回`预留资源冲突`与说明
        2. 如果请求资源与已使用资源冲突，则返回`资源冲突`与说明
        3. 无冲突，以任务的形式向预留资源项添请求资源，记录vector配置项，并返回`成功`
      * 根据安装任务第一阶段状态决定以下操作
        1. 操作`成功`，进行下一个资源预留任务(下一个安装任务的第一阶段)
        2. 返回`资源冲突`，记录资源冲突说明，继续下一个资源预留任务或终止安装部署。
           * 继续下一个安装任务，是因为可以继续检查剩下任务的是否可行，有利于对app整体进行检查并生成报告
           * 直接终止安装部署，进入部署失败处理。直接终止目的在于减少安装部署等待时间，以及避免不必要的资源预留。
           * 两种行为有待考虑，一种有利于对整体app的异常行为诊断，第二种有利于减少app安装的资源占用和性能
        3. 返回`预留资源冲突`，则有可能是其他app正在安装中，因此可以休眠一段时间后再次进行当前任务的第一阶段。如果相同问题出现3次，则有可能是多个app间安装存在资源竞争，基于已有说明增加额外说明，当前任务视为`资源冲突`进行后续操作。
   2. 当所有资源预留任务成功执行后，执行安装任务第二阶段，该阶段完成对任务的准备，准备在安装失败后的服务回退，可以成为**安装准备任务**，该阶段称为**安装准备阶段**。如果**安装准备任务**失败则直接进入安装失败处理，但是理论上安装准备任务不会由于业务原因导致执行失败，更多的是内部原因导致，因此安装准备任务的实现需要考虑请求其他服务的超时以及重试，在一定的策略尝试失败后再视为任务失败。例如datastream的vector子安装任务第三阶段向etl发送配置准备请求。
      * etl首先会根据任务的ID，获取配置项，配合vector服务，结合实际vector数据流对象与任务的配置计算得出任务配置生效需要进行的change，new，delelte任务以及与回退需要进行的rollback_change，rollback_new，rolllback_delete，rollback任务。
   3. 当所有安装准备任务执行成功后，进行最后的安装任务第三阶段，称为**安装执行任务**，该阶段称为**执行确认阶段**。安装提交任务实际完成对应服务真正的任务启动，数据库存储等。如果某个安装提交任务失败，进入安装失败处理阶。例如datastream的子安装任务第三阶段向etl发送配置生效请求。
      * etl将向vector集群各个节点发送vector配置，由vector执行数据流更新，如果某个vector
   4. 安装失败处理阶段，根据安装任务失败位置，分为三种失败处理：
      1. 资源预留阶段失败，则对已完成资源预留任务进行预留资源撤销操作释放预留资源，销毁任务，并生成失败报告
      2. 安装准备阶段失败，对已完成资源预留任务进行撤销操作释放资源，销毁任务，并生成失败报告
      3. 执行确认阶段失败，首先对安装任务执行回退操作，使服务状态回滚，再对所有任务进行销毁操作，释放预留资源，销毁任务，并生成报告。

#### 创建任务

安装服务根据app配置文件路径规划创建出对应的安装任务。

* `service`：对应k8s对象，创建k8s部署任务，部署任务在k8s对象部署时，提供存储路由、命名空间、额外标签等参数。对应yaml文件如何使用参数由yaml决定，部署任务仅提供参数，不负责实际内容的改变。
  * `service`的安装会以`volume`任务为最优先进行安装，其成功后对`prevJob`目录下的任务进行安装
  * 提供额外的`prevJob`目录，允许在`volume`完成后进行安装，该目录的意义在于对各个允许进行初始化/升级任务的进行。随后执行其他对象的创建，并且不考虑这些对象实际状态，如果通过接口能够获取到对象状态则作为安装报告的一部分，但不考虑除了语法错误外的安装失败。
  * 提供`uninstallJob`目录，用于在进行回滚/卸载时首先执行的清理任务，随后通过特定k8s直接删除带有该app标签的对象(最后删除volume对象)，如果是rollback，则将对象回退到安装之前记录的配置(数据咋办呀)。
* `search`：search配置导入任务，每种任务的名称会在头部添加"app"名称作为前缀，以此与其他搜索区分(数据库中目前没有命名空间隔离，可以考虑)，并将多个app相互隔离。在回退或卸载时仅需要对已经导入的搜索配置进行删除操作。在资源预留阶段仅作配置检测，在安装准备阶段记录导入的配置，在执行安装阶段调用已有的导入接口进行导入，回退或卸载则调用删除接口删除配置。
* `alter`：告警功能，涉及任务实例创建，需要考虑是否参照安装任务的三阶段进行。告警任务如果不涉及共享资源问题，可以考虑如`search`功能一样进行隔离、检测和安装。
* `sqlconnector`：同`search`
* `agent`：
* `label`：
* `presearch`：暂无
* `index.conf`：暂无，目前的索引管理将会重做到`datastream`
* `datastream`： 目前考虑分为三种子任务，分别为vector，logstash和kafka(或其他数据队列)。
  * kafka：负责对kafka的topic配置进行调整(最优先进行)。
    * 资源预留任务：检查kafka状态，保存当前kafka的状态信息，检查配置的合法性。
    * 安装准备任务：无
    * 安装执行任务：执行kafka的topic配置
    * 回滚任务：完成对kafka已创建topic的清理(如果允许)
  * vector：负责对vector配置的调整
    * 资源预留任务：读取app配置，通过etl接口检查配置合法性和进行资源预留，etl结合当前配置、预留配置和app配置结合请求对应角色的vector服务实例进行检查(如果一定无实例，则通过平台接口判断是否应该等待一段时间重试或作为异常抛出)。如果vector返回通过检查，则将app配置和预留配置合并作为新预留配置，当前配置不变。
    * 安装准备任务：无
    * 安装执行任务：将配置预留配置合并到当前配置，并触发每个vector实例的变化。
    * 回滚任务: 从如果当前配置中含有app配置，则将当前配置中的app配置去除，并触发各个vector实例的变化。从预留配置中去除app配置。如果当前配置或预留配置中含有某个对象包含当前app的标签，将作为报告一部分进行说明，由使用者进行调整
  * logstash：负责对logstash配置的调整
    * 资源预留任务：读取app配置，通过etl接口检查配置合法性(logstash即将替换，剩下更能应该不会存在共享资源竞争问题，可以考虑不进行资源预留)。
    * 安装准备任务：将当前配置进行快照。
    * 安装执行任务：将配置文件和app配置合并后写回etcd，等待logstash进行同步。
    * 回滚任务: 将配置文件回滚到快照时。

#### 安装任务抽象

对于大部分不存在资源冲突的功能，其配置可以通过添加**namespace**进行隔离，每个app的namespace为其名称。因此无共享资源的配置可以通过添加app名称为前缀进行隔离(目前数据库无namespace概念和相关权限控制，可以考虑引入，不过会导致数据库表结构改变，不引入则需要在名称中确定前缀与名称的分隔符)。在资源预留任务中进行合法性检查，不需要资源预留。在安装准备任务中记录完成`todo`,`rollback`操作。在安装执行任务中进行`todo`操作，进行回退时执行`rollback`操作。

安装任务默认实现，实现以下任务的默认实现，各种安装任务可以基于安装任务的默认实现根据需要进行重写。

* `func create(meta)`：由子类实现，负责根据配置文创建安装任务的初始4化
* `func first()`：执行资源预留任务，默认返回成功

* `func second()`：执行安装准备任务，得出`todo`和`rollback`操作。
* `func third()`：执行安装任务，执行`todo`操作，失败后向上抛出错误
* `func rollback()`：执行`rollback`任务

#### 资源预留阶段与资源预留任务

资源预留任务对所有安装任务进行合法检测，以及对需要资源进行预留。确保在安装任务业务合理，存在足够的资源，尽早的发现业务和非业务错误。各个服务实际的资源预留也可以在安装准备阶段进行，但是考虑到事务的开始是从第一阶段开始，所以希望在事务通过检查后就为事务预留号足够的资源，避免在这个过程中由于其他安装任务的并发进行，导致第一阶段的通过资源检测造成后续的错误。

该阶段的资源预留任务，在合法性检测时出现错误，则直接作为错误抛出。在资源检测时，如果是已用资源冲突，则对作为错误抛出。如果是预留资源冲突，则在进行3次重试后如果仍然不能获取预留资源，则作为错误抛出。

对于一般的简单任务且确定不会出错任务，可以考虑什么也不做。

#### 安装准备阶段与安装准备任务

主要是为了计算出`todo`和`rollback`操作的行为。为了避免一些共享配置的`rollback`的正确性，需要考虑对这些配置进行加锁，确保不会由于其他操作导致配置的变更，产生超出预期的行为。

如果通过配置版本管理服务管理多个版本，且每个功能模块都具有相同格式的导出配置，提供统一的增/删/该接口，则能够提供默认实现计算出每种功能的`todo`和`rollabck`。但是成默认实现，要求配置的版本管理、所有功能的配置文件以统一的内容格式，并且要求各个服务提供统一标准的接口。例如splunk中大部分配置文件都是同一种`stanze`格式，所以他可以做到基于内容求出两个版本间的差集和交集，从而得出增/删/改任务。如果辅以统一的接口，各种任务配置可以直接转换为各个接口的参数进行调度。但是目前以我们的导出配置和接口做不到，需要各个功能模块，各自进行。

#### 安装执行阶段与安装执行任务

在此阶段之前整体系统不会产生过多的变化，主要是对资源的预留。在安装执行阶段进行后整体系统会开始进行变化，为了保证系统状态变更的事务性，对于竞争资源在资源预留阶段已经进行预留，确保不会被其他安装任务同时占用，在安装准备阶段已进行回滚的准备。

该阶段则进行`todo`操作，对系统进行变化，如果某个任务出现异常则对所有任务进行回滚。如果某个`todo`操作失败但并不希望对整个安装任务进行回滚，则应该抛出特定错误类型，并附加说明作为最终任务报告的一部分，指导用户进行调整。



