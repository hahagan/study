todo

# 一、日志载体

假设S(a)表示对象a的schema。若对象A含有子对象时，若存在公式:

```
S(a) = F(a).S(a0).S(a1)......S(an)
```

其中`a0,a1...an`表示a的子对象，若用`G(a)`表示`S(a0).S(a1)......S(an)`,可得公式

```
S(a)=F(a).G(a)
```

继续下专可得公式

```
S(a) = F(a).F(a0).G(a0).F(a1).G(a1)......F(an)G(an)
```

根据G(x)定义可知，当x不存在子对象时G(x)为空。因此

```
S(a) = F(a).F(a0)......F(a1)......F(an)
```

如果F(a).F(a0)满足交换律则可以满足公式

```
S(a) = F(a).F(a0).F(a1)...F(an).F(a00).F(a01)......F(a0n)......F(ann)
```

如果满足该公式，那么在进行序列化时可以通过广度优先的方式完成序列化。在数据逆序列化时时可以仅进行头部的解析，对数据的真正解析延迟到数据使用时。因此日志在日志接收和处理过程中可以最大程度的避免数据解析。

```rust
pub Schema trait {
    new() -> Self
    add<T>(&mut self, T) error
    get(&self, key string) &mut entry
   
    
}

struct schema {
    keysCache []*Array
    keys []string
    quic hash
}

// 列式存储，将相同的数据存储到同一列中
struct Column {
    Type string
    Values []interface{}
    // 考虑到容量限制、不同处理端点的数据处理
    // 支持列分段，方便进行添加和扩容，可能会降低查找性能
    Next *[]interface    
}
```



# 二、思路

序列化输出时对数据进行整理。每当进入一个新的数据时，根据数据完成schema的整理，shcema可以重新更新，对于稳定的系统在一段时间的运行后，schema必然趋于稳定。如果schema出现冲突则可以认为是数据的产生者的错误使用。这里要求使用者对shcema有统一的认知。对于冲突的schema除了报错外，也许也可以通过生成新的shcema完成，但是需要额外考虑schema的选择，也许联合挂载技术可以用于在多个shcema间完成兼容和选择的能力。

