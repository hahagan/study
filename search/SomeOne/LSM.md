## LSM 存储引擎
假定现在我们要存储N对key-value，那么我们同样需要在索引里面保存N对key-offset。但是，如果数据文件本身是按序存放的，我们就没必要对每个key建索引了。我们可以将key划分成若干个block，只索引每个block的start_key。对于其它key，根据大小关系找到它存在的block，然后在block内部做顺序搜索即可

在LSM里面，我们把按序组织的数据文件称为**SSTable**（Sorted String Table）。
通过稀疏索引指向SSTable中的block，同时可以对block数据进行压缩，减少磁盘的IO吞吐量

### SSTable的维护和生成
在内存里面维护一个平衡二叉树（例如AVL树或者红黑树）。每当有Put(Key, Value)请求时，先将数据写入二叉树，保证其顺序性。当二叉树达到既定规模时，我们将其按序写入到磁盘，转换成SSTable存储下来。
在LSM里面，我们把内存里的二叉树称为memtable

### 故障恢复
如果在将memtable转存SSTable时，进程挂掉了，怎么保证未写入SSTable的数据不丢失呢？
参考数据库的**redo log**，我们也可以搞一个log记录当前memtable的写操作。在有Put请求过来时，除了写入memtable，还将操作追加到log。当memtable成功转成SSTable之后，它对应的log文件就可以删除了。在下次启动时，如果发现有残留的log文件，先通过它恢复上次的memtable。

todo：**redo log** 如何实现的，如果同样需要落盘，那么这种方式和直接将memtable落盘相比优点在哪？

### Bloom Filter
对于查询那些不存在的key，我们需要搜索完memtable和所有的SSTable，才能确定地说它不存在。

在数据量不大的情况下，这不是个问题。但是当数据量达到一定的量级后，这会对系统性能造成非常严重的问题。

我们可以借助Bloom Filter（布隆过滤器）来快速判断一个key是否存在。

思考：从某种意义上来看Bloom Filter和codesearch中获取索引中的file类似，通过一定的算法快速的筛选出更少量的数据。虽然不能保证这些数据中一定包含查询的key，但是可以保证除了这些数据外不会包含查询的key

### 合并策略
合并策略指：
* 什么时候合并
* 哪些SSTable会被合并

广泛合并策略应用的合并策略：
* size-tiered(STCS)
    * 当某个尺寸的SSTable数量达到既定个数时，合并为一个SSTable
        * 优点是实现简单，确定是合并时的空间放大效应
* leveled（LCS)
    * 就是将数据分成互不重叠的一系列固定大小（例如 2 MB）的SSTable文件，再将其分层（level）管理。对每个Level，我们都有一份清单文件记录着当前Level内每个SSTable文件存储的key的范围。
    * Level和Level的区别在于它所保存的SSTable文件的最大数量：Level-L最多只能保存 10 L 个SSTable文件（但是Level 0是个例外，后面再说）。



