时序数据特点：按时间增量

技术关键词：时间戳增量保存，varInt， 差值保存(XOR 浮点计算)，差值保存

# 官方技术博客

A series of blog posts explaining different components of TSDB:

\* [The Head Block](https://ganeshvernekar.com/blog/prometheus-tsdb-the-head-block/)

\* [WAL and Checkpoint](https://ganeshvernekar.com/blog/prometheus-tsdb-wal-and-checkpoint/)

\* [Memory Mapping of Head Chunks from Disk](https://ganeshvernekar.com/blog/prometheus-tsdb-mmapping-head-chunks-from-disk/)

\* [Persistent Block and its Index](https://ganeshvernekar.com/blog/prometheus-tsdb-persistent-block-and-its-index/)

\* [Queries](https://ganeshvernekar.com/blog/prometheus-tsdb-queries/)



# 磁盘存储结构

## head chunk

```
┌──────────────────────────────┐
│  magic(0x0130BC91) <4 byte>  │
├──────────────────────────────┤
│    version(1) <1 byte>       │
├──────────────────────────────┤
│    padding(0) <3 byte>       │
├──────────────────────────────┤
│ ┌──────────────────────────┐ │
│ │         Chunk 1          │ │
│ ├──────────────────────────┤ │
│ │          ...             │ │
│ ├──────────────────────────┤ │
│ │         Chunk N          │ │
│ └──────────────────────────┘ │
└──────────────────────────────┘

```

### chunk

此时不存在索引用于关联该chunk，因此需要记录chunk相关的原信息，官方文档中结构如下

```
┌─────────────────────┬───────────────────────┬───────────────────────┬───────────────────┬───────────────┬─
| series ref <8 byte> | mint <8 byte, uint64> | maxt <8 byte, uint64> | encoding <1 byte> | len <uvarint> | 
└─────────────────────┴───────────────────────┴───────────────────────┴───────────────────┴───────────────┴─
─────────────┬────────────────┐
data <bytes> │ CRC32 <4 byte> │
─────────────┴────────────────┘


```

实际代码解析中`data`记录了chunk的sample数量，所以其加载chunk时的结构为

```
┌─────────────────────┬───────────────────────┬───────────────────────┬───────────────────┬───────────────┬─
| series ref <8 byte> | mint <8 byte, uint64> | maxt <8 byte, uint64> | encoding <1 byte> | len <uvarint> | 
└─────────────────────┴───────────────────────┴───────────────────────┴───────────────────┴───────────────┴─
┬──────────────────────┬─────────────────┬────────────────┐
│  numSamples <uint16> │ samples <bytes> │ CRC32 <4 byte> │
┴──────────────────────┴─────────────────┴────────────────┘
```

对应的head_chunk文件的加载位于`tsdb/head.go`的`(h *Head) loadMmappedChunks`函数，该函数会调用`tsdb/chunks/head_chunks.go`的`(cdm *ChunkDiskMapper) IterateAllChunks`函数加载head_chunk中的每个chunk对象。加载chunk对象的部分代码如下
```go
			chunkRef := chunkRef(uint64(segID), uint64(idx))

			startIdx := idx
			seriesRef := binary.BigEndian.Uint64(mmapFile.byteSlice.Range(idx, idx+SeriesRefSize))
			idx += SeriesRefSize
			mint := int64(binary.BigEndian.Uint64(mmapFile.byteSlice.Range(idx, idx+MintMaxtSize)))
			idx += MintMaxtSize
			maxt := int64(binary.BigEndian.Uint64(mmapFile.byteSlice.Range(idx, idx+MintMaxtSize)))
			idx += MintMaxtSize

			// We preallocate file to help with m-mapping (especially windows systems).
			// As series ref always starts from 1, we assume it being 0 to be the end of the actual file data.
			// We are not considering possible file corruption that can cause it to be 0.
			// Additionally we are checking mint and maxt just to be sure.
			if seriesRef == 0 && mint == 0 && maxt == 0 {
				break
			}

			idx += ChunkEncodingSize // Skip encoding.
			dataLen, n := binary.Uvarint(mmapFile.byteSlice.Range(idx, idx+MaxChunkLengthFieldSize))
			idx += n

			numSamples := binary.BigEndian.Uint16(mmapFile.byteSlice.Range(idx, idx+2))
			idx += int(dataLen) // Skip the data.
```



## Index

```
┌────────────────────────────┬─────────────────────┐
│ magic(0xBAAAD700) <4b>     │ version(1) <1 byte> │
├────────────────────────────┴─────────────────────┤
│ ┌──────────────────────────────────────────────┐ │
│ │                 Symbol Table                 │ │
│ ├──────────────────────────────────────────────┤ │
│ │                    Series                    │ │
│ ├──────────────────────────────────────────────┤ │
│ │                 Label Index 1                │ │
│ ├──────────────────────────────────────────────┤ │
│ │                      ...                     │ │
│ ├──────────────────────────────────────────────┤ │
│ │                 Label Index N                │ │
│ ├──────────────────────────────────────────────┤ │
│ │                   Postings 1                 │ │
│ ├──────────────────────────────────────────────┤ │
│ │                      ...                     │ │
│ ├──────────────────────────────────────────────┤ │
│ │                   Postings N                 │ │
│ ├──────────────────────────────────────────────┤ │
│ │               Label Index Table              │ │
│ ├──────────────────────────────────────────────┤ │
│ │                 Postings Table               │ │
│ ├──────────────────────────────────────────────┤ │
│ │                      TOC                     │ │
│ └──────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────┘

```

### Symbo Table

```
┌────────────────────┬─────────────────────┐
│ len <4b>           │ #symbols <4b>       │
├────────────────────┴─────────────────────┤
│ ┌──────────────────────┬───────────────┐ │
│ │ len(str_1) <uvarint> │ str_1 <bytes> │ │
│ ├──────────────────────┴───────────────┤ │
│ │                . . .                 │ │
│ ├──────────────────────┬───────────────┤ │
│ │ len(str_n) <uvarint> │ str_n <bytes> │ │
│ └──────────────────────┴───────────────┘ │
├──────────────────────────────────────────┤
│ CRC32 <4b>                               │
└──────────────────────────────────────────┘

```

### series

一个series按持有的labelName和labelValue划分，即 A {a=b,c=d} 和 A { a=b, c=d'}即使lableName相同，由于labelValue不同会被划分为不同的series。

* 每个series按16字节对齐，因此每个seriesID为`offset/16`
* 每个chunk记录都记录了对应series在该chunk中的最早sample时间(`mint`)和最晚sample时间(`maxt`增量表示)，以及chunk的偏移量
* 按label集的字典序排序
* 时机上一条记录A {a=b,c=d}会被转变为{\__name__=A, a=b,c=d}

```
┌───────────────────────────────────────┐
│ ┌───────────────────────────────────┐ │
│ │   series_1                        │ │
│ ├───────────────────────────────────┤ │
│ │                 . . .             │ │
│ ├───────────────────────────────────┤ │
│ │   series_n                        │ │
│ └───────────────────────────────────┘ │
└───────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────┐
│ len <uvarint>                                                            │
├──────────────────────────────────────────────────────────────────────────┤
│ ┌──────────────────────────────────────────────────────────────────────┐ │
│ │                     labels count <uvarint64>                         │ │
│ ├──────────────────────────────────────────────────────────────────────┤ │
│ │              ┌────────────────────────────────────────────┐          │ │
│ │              │ ref(l_i.name) <uvarint32>                  │          │ │
│ │              ├────────────────────────────────────────────┤          │ │
│ │              │ ref(l_i.value) <uvarint32>                 │          │ │
│ │              └────────────────────────────────────────────┘          │ │
│ │                             ...                                      │ │
│ ├──────────────────────────────────────────────────────────────────────┤ │
│ │                     chunks count <uvarint64>                         │ │
│ ├──────────────────────────────────────────────────────────────────────┤ │
│ │              ┌────────────────────────────────────────────┐          │ │
│ │              │ c_0.mint <varint64>                        │          │ │
│ │              ├────────────────────────────────────────────┤          │ │
│ │              │ c_0.maxt - c_0.mint <uvarint64>            │          │ │
│ │              ├────────────────────────────────────────────┤          │ │
│ │              │ ref(c_0.data) <uvarint64>                  │----------| |-------> chunk
│ │              └────────────────────────────────────────────┘          │ │
│ │              ┌────────────────────────────────────────────┐          │ │
│ │              │ c_i.mint - c_i-1.maxt <uvarint64>          │          │ │
│ │              ├────────────────────────────────────────────┤          │ │
│ │              │ c_i.maxt - c_i.mint <uvarint64>            │          │ │
│ │              ├────────────────────────────────────────────┤          │ │
│ │              │ ref(c_i.data) - ref(c_i-1.data) <varint64> │          │ │
│ │              └────────────────────────────────────────────┘          │ │
│ │                             ...                                      │ │
│ └──────────────────────────────────────────────────────────────────────┘ │
├──────────────────────────────────────────────────────────────────────────┤
│ CRC32 <4b>                                                               │
└──────────────────────────────────────────────────────────────────────────┘

```

### Label Index

记录了label名称和label的values在符号表的偏移量

```
┌───────────────┬────────────────┬────────────────┐
│ len <4b>      │ #names <4b>    │ #entries <4b>  │
├───────────────┴────────────────┴────────────────┤
│ ┌─────────────────────────────────────────────┐ │
│ │ ref(value_0) <4b>                           │ │
│ ├─────────────────────────────────────────────┤ │
│ │ ...                                         │ │
│ ├─────────────────────────────────────────────┤ │
│ │ ref(value_n) <4b>                           │ │
│ └─────────────────────────────────────────────┘ │
│                      . . .                      │
├─────────────────────────────────────────────────┤
│ CRC32 <4b>                                      │
└─────────────────────────────────────────────────┘

```

### postings

某个labelName和LableValue的倒排索引项，指向多个series。被倒排索引表引用。

```
┌────────────────────┬────────────────────┐
│ len <4b>           │ #entries <4b>      │
├────────────────────┴────────────────────┤
│ ┌─────────────────────────────────────┐ │
│ │ ref(series_1) <4b>                  │ │
│ ├─────────────────────────────────────┤ │
│ │ ...                                 │ │
│ ├─────────────────────────────────────┤ │
│ │ ref(series_n) <4b>                  │-│---------------> series
│ └─────────────────────────────────────┘ │
├─────────────────────────────────────────┤
│ CRC32 <4b>                              │
└─────────────────────────────────────────┘

```

### label Index table

label的索引表

```
┌─────────────────────┬──────────────────────┐
│ len <4b>            │ #entries <4b>        │
├─────────────────────┴──────────────────────┤
│ ┌────────────────────────────────────────┐ │
│ │  n = 1 <1b>                            │ │
│ ├──────────────────────┬─────────────────┤ │
│ │ len(name) <uvarint>  │ name <bytes>    │ │
│ ├──────────────────────┴─────────────────┤ │
│ │  offset <uvarint64>                    │-│----------> label index
│ └────────────────────────────────────────┘ │
│                    . . .                   │
├────────────────────────────────────────────┤
│  CRC32 <4b>                                │
└────────────────────────────────────────────┘

```

### posting index label

倒排索引表，记录各个LabelName和LabelValue，指向`posting`区域，可查询相关的series

```
┌─────────────────────┬──────────────────────┐
│ len <4b>            │ #entries <4b>        │
├─────────────────────┴──────────────────────┤
│ ┌────────────────────────────────────────┐ │
│ │  n = 2 <1b>                            │ │
│ ├──────────────────────┬─────────────────┤ │
│ │ len(name) <uvarint>  │ name <bytes>    │ │
│ ├──────────────────────┼─────────────────┤ │
│ │ len(value) <uvarint> │ value <bytes>   │ │
│ ├──────────────────────┴─────────────────┤ │
│ │  offset <uvarint64>                    │-│---------> posting
│ └────────────────────────────────────────┘ │
│                    . . .                   │
├────────────────────────────────────────────┤
│  CRC32 <4b>                                │
└────────────────────────────────────────────┘

```



# 合并策略

在protmehus的在启动时会打开TSBD实例，该实例会启动额外的的gorountine，该gorountine每隔1分支定期的重新加载blocks，清理过期block并触发各个block的合并请求。

1. 该代码位于`tsdb/db.go`的`(db *DB) run()`函数

```
func (db *DB) run() {
	defer close(db.donec)

	backoff := time.Duration(0)

	for {
		select {
		case <-db.stopc:
			return
		case <-time.After(backoff):
		}

		select {
		case <-time.After(1 * time.Minute):
			db.cmtx.Lock()
			if err := db.reloadBlocks(); err != nil {
				level.Error(db.logger).Log("msg", "reloadBlocks", "err", err)
			}
			db.cmtx.Unlock()

			select {
			case db.compactc <- struct{}{}:
			default:
			}
		case <-db.compactc:
			db.metrics.compactionsTriggered.Inc()

			db.autoCompactMtx.Lock()
			if db.autoCompact {
				if err := db.Compact(); err != nil {
					level.Error(db.logger).Log("msg", "compaction failed", "err", err)
					backoff = exponential(backoff, 1*time.Second, 1*time.Minute)
				} else {
					backoff = 0
				}
			} else {
				db.metrics.compactionsSkipped.Inc()
			}
			db.autoCompactMtx.Unlock()
		case <-db.stopc:
			return
		}
	}
}
```

2. 从上述代码中可以看到触发了`func (db *DB) Compact() (returnErr error)`进行compact，该过程中会计算需要持久化的head chunk，truncate对应的WAL文件，并最终执行各个block的合并。具体代码位于`tsdb/db.go`

3. 执行具体block合并的对象为`tsdb/compact.go`的LeveledCompactor对象，该对象的`Plane`方法用于判断需要进行合并的blocks并生成对应的计划,并对block进行合并，合并流程如下。

   1. 首先打开未打开的block，并将各个block的meta通过函数`CompactBlockMetas`合并
   2. 根据合并meta对创建新的合并block对象实例临时目录，创建 index.Writer和chunks.Writer
   3. 调用`populateBlock`完成对合并索引和chunk文件的操作
   4. 完成对meta文件等的写操作，并最终在各种同步操作完成后将临时目录变更为正式目录，对外提供服务

4. `populateBlock`函数中将各个block的series按字典序排序(即lable的字典排序)，将各个block合并。

   1. 通过`newBlockChunkSeriesSet`函数将各个block转换为serires对象集，根据该对象集可以读取对应各个block的serires。

   2. 在通过`storage.NewMergeChunkSeriesSet`函数将各个serites对象集合并为一个serires集迭代器，使得最终合并block时遍历该seiresSet时按各个组成block的series按字典序访问，并合并具有相同label的series。排序算法为堆排序，对应seriesSet的实现代码为`storage/merge.go`的`genericMergeSeriesSet`对象。具体代码如下

      ```go
      func newGenericMergeSeriesSet(sets []genericSeriesSet, mergeFunc genericSeriesMergeFunc) genericSeriesSet {
      	if len(sets) == 1 {
      		return sets[0]
      	}
      
      	// We are pre-advancing sets, so we can introspect the label of the
      	// series under the cursor.
      	var h genericSeriesSetHeap
      	for _, set := range sets {
      		if set == nil {
      			continue
      		}
      		if set.Next() {
      			heap.Push(&h, set)
      		}
      		if err := set.Err(); err != nil {
      			return errorOnlySeriesSet{err}
      		}
      	}
      	return &genericMergeSeriesSet{
      		mergeFunc: mergeFunc,
      		sets:      sets,
      		heap:      h,
      	}
      }
      
      func (c *genericMergeSeriesSet) Next() bool {
      	// Run in a loop because the "next" series sets may not be valid anymore.
      	// If, for the current label set, all the next series sets come from
      	// failed remote storage sources, we want to keep trying with the next label set.
      	for {
      		// Firstly advance all the current series sets. If any of them have run out,
      		// we can drop them, otherwise they should be inserted back into the heap.
      		for _, set := range c.currentSets {
      			if set.Next() {
      				heap.Push(&c.heap, set)
      			}
      		}
      
      		if len(c.heap) == 0 {
      			return false
      		}
      
      		// Now, pop items of the heap that have equal label sets.
      		c.currentSets = nil
      		c.currentLabels = c.heap[0].At().Labels()
      		for len(c.heap) > 0 && labels.Equal(c.currentLabels, c.heap[0].At().Labels()) {
      			set := heap.Pop(&c.heap).(genericSeriesSet)
      			c.currentSets = append(c.currentSets, set)
      		}
      
      		// As long as the current set contains at least 1 set,
      		// then it should return true.
      		if len(c.currentSets) != 0 {
      			break
      		}
      	}
      	return true
      }
      
      func (c *genericMergeSeriesSet) At() Labels {
      	if len(c.currentSets) == 1 {
      		return c.currentSets[0].At()
      	}
      	series := make([]Labels, 0, len(c.currentSets))
      	for _, seriesSet := range c.currentSets {
      		series = append(series, seriesSet.At())
      	}
      	return c.mergeFunc(series...)
      }
      ```

   3. 迭代合并的seriesSet对象，将其chunk和series写入合并block和合并索引中
   
   ```go
   	for set.Next() {
   		select {
   		case <-c.ctx.Done():
   			return c.ctx.Err()
   		default:
   		}
   		s := set.At()
   		chksIter := s.Iterator()
           chks = chks[:0]
   		for chksIter.Next() {
   			// We are not iterating in streaming way over chunk as it's more efficient to do bulk write for index and
   			// chunk file purposes.
   			chks = append(chks, chksIter.At())
   		}
       }
   ```
   
   
   
5. 在合并的seiresSet对象中，迭代后会对相同的series的chunk进行合并。

   1. 该`mergeFunc`为`storage/generic.go`文件中的`chunkSeriesMergerAdapter.Merge()`函数，而该函数会进一步调用db打开时创建`LeveledCompactor`实例时传入的chunk合并方法

      ```go
      type chunkSeriesMergerAdapter struct {
      	VerticalChunkSeriesMergeFunc
      }
      
      func (a *chunkSeriesMergerAdapter) Merge(s ...Labels) Labels {
      	buf := make([]ChunkSeries, 0, len(s))
      	for _, ser := range s {
      		buf = append(buf, ser.(ChunkSeries))
      	}
      	return a.VerticalChunkSeriesMergeFunc(buf...)
      }
      ```
   
   2. `VerticalChunkSeriesMergeFunc`为`storage.NewCompactingChunkSeriesMerger(storage.ChainedSeriesMerge)`函数，该函数会返回一个`ChunkSeriesEntry`实例，该实例通过`Iterator()`得到的是一个`storage.compactChunkIterator`对象实例。`storage.compactChunkIterator`负责将传入的`ChunkSeries`的各个chunk进行"合并迭代"。代码位于`storage/merge.go`文件
   
   ```go
   // NewCompactingChunkSeriesMerger returns VerticalChunkSeriesMergeFunc that merges the same chunk series into single chunk series.
   // In case of the chunk overlaps, it compacts those into one or more time-ordered non-overlapping chunks with merged data.
   // Samples from overlapped chunks are merged using series vertical merge func.
   // It expects the same labels for each given series.
   //
   // NOTE: Use the returned merge function only when you see potentially overlapping series, as this introduces small a overhead
   // to handle overlaps between series.
   func NewCompactingChunkSeriesMerger(mergeFunc VerticalSeriesMergeFunc) VerticalChunkSeriesMergeFunc {
   	return func(series ...ChunkSeries) ChunkSeries {
   		if len(series) == 0 {
   			return nil
   		}
   		return &ChunkSeriesEntry{
   			Lset: series[0].Labels(),
   			ChunkIteratorFn: func() chunks.Iterator {
   				iterators := make([]chunks.Iterator, 0, len(series))
   				for _, s := range series {
   					iterators = append(iterators, s.Iterator())
   				}
   				return &compactChunkIterator{
   					mergeFunc: mergeFunc,
   					iterators: iterators,
   				}
   			},
   		}
   	}
   }
   ```
   
   3. `storage.compactChunkIterator`将组成该迭代器的各个serires的当前chunk按堆排序，堆排序的标准为各个chunk的MaxTime和MinTime。在堆中将堆首作为当前chunk，并堆其余chunk进行遍历，合并时间窗口有重叠的chunk。代码位于`storage/merge.go`文件
   
      ```go
      func (c *compactChunkIterator) Next() bool {
      	if c.h == nil {
      		for _, iter := range c.iterators {
      			if iter.Next() {
      				heap.Push(&c.h, iter)
      			}
      		}
      	}
      	if len(c.h) == 0 {
      		return false
      	}
      
      	iter := heap.Pop(&c.h).(chunks.Iterator)
      	c.curr = iter.At()
      	if iter.Next() {
      		heap.Push(&c.h, iter)
      	}
      
      	var (
      		overlapping []Series
      		oMaxTime    = c.curr.MaxTime
      		prev        = c.curr
      	)
      	// Detect overlaps to compact. Be smart about it and deduplicate on the fly if chunks are identical.
      	for len(c.h) > 0 {
      		// Get the next oldest chunk by min, then max time.
      		next := c.h[0].At()
      		if next.MinTime > oMaxTime {
      			// No overlap with current one.
      			break
      		}
      
      		if next.MinTime == prev.MinTime &&
      			next.MaxTime == prev.MaxTime &&
      			bytes.Equal(next.Chunk.Bytes(), prev.Chunk.Bytes()) {
      			// 1:1 duplicates, skip it.
      		} else {
      			// We operate on same series, so labels does not matter here.
      			overlapping = append(overlapping, newChunkToSeriesDecoder(nil, next))
      			if next.MaxTime > oMaxTime {
      				oMaxTime = next.MaxTime
      			}
      			prev = next
      		}
      
      		iter := heap.Pop(&c.h).(chunks.Iterator)
      		if iter.Next() {
      			heap.Push(&c.h, iter)
      		}
      	}
      	if len(overlapping) == 0 {
      		return true
      	}
      
      	// Add last as it's not yet included in overlap. We operate on same series, so labels does not matter here.
      	iter = NewSeriesToChunkEncoder(c.mergeFunc(append(overlapping, newChunkToSeriesDecoder(nil, c.curr))...)).Iterator()
      	if !iter.Next() {
      		if c.err = iter.Err(); c.err != nil {
      			return false
      		}
      		panic("unexpected seriesToChunkEncoder lack of iterations")
      	}
      	c.curr = iter.At()
      	if iter.Next() {
      		heap.Push(&c.h, iter)
      	}
      	return true
      }
      ```
   
   4. 对chunk进行合并的函数为在db的open函数中在通过`tsdb/compact.go`的`NewLeveledCompactorWithChunkSize`函数创建`LeveledCompactor`，时传入的合并方法，但实际此时传入合并方法为空，因此实际的合并方法为`storage/merge.go`文件的`storage.NewCompactingChunkSeriesMerger(storage.ChainedSeriesMerge)`函数
   
      ```go
      // ChainedSeriesMerge returns single series from many same, potentially overlapping series by chaining samples together.
      // If one or more samples overlap, one sample from random overlapped ones is kept and all others with the same
      // timestamp are dropped.
      //
      // This works the best with replicated series, where data from two series are exactly the same. This does not work well
      // with "almost" the same data, e.g. from 2 Prometheus HA replicas. This is fine, since from the Prometheus perspective
      // this never happens.
      //
      // It's optimized for non-overlap cases as well.
      func ChainedSeriesMerge(series ...Series) Series {
      	if len(series) == 0 {
      		return nil
      	}
      	return &SeriesEntry{
      		Lset: series[0].Labels(),
      		SampleIteratorFn: func() chunkenc.Iterator {
      			iterators := make([]chunkenc.Iterator, 0, len(series))
      			for _, s := range series {
      				iterators = append(iterators, s.Iterator())
      			}
      			return newChainSampleIterator(iterators)
      		},
      	}
      }
      ```
   
6. 结合以上流程可以发现整个blokc合并流程为

   1. 将block抽象为一个SeriesSet对象`storage.genericMergeSeriesSet`
   2. 迭代 `storage.genericMergeSeriesSet`对象，实际是通过堆排序合并了各个block的相同series对象，此时各个series的chunk未合并
   3. 获取`storage.genericMergeSeriesSet`的chunk迭代器时，将各个block同名的series的chunk迭代器组合获得`storage.compactChunkIterator`实例，`storage.compactChunkIterator`通过堆排序迭代各个各个series的chunk，如果此时有chunk存在重叠则进行合并。
   4. 迭代合并Series对象的chunk获得合并chunk并写入对应的chunk文件中
   5. 将Series写入index文件，并生成倒排索引等信息

https://liujiacai.net/blog/2021/04/11/prometheus-storage-engine/#headline-5

https://liujiacai.net/blog/2021/04/11/prometheus-storage-engine/#headline-5

