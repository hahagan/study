### https://doc.rust-lang.org/std/collections/struct.HashMap.html#method.get



### 犯错误

```
pub fn get<Q: ?Sized>(&self, k: &Q) -> Option<&V>
where
    K: Borrow<Q>,
    Q: Hash + Eq, 
```

1. 使用`get`时没有传入不可变借用，而是直接传入变量
2. 获取`get`结果时没有注意对`option`进行处理

```
pub fn insert(&mut self, k: K, v: V) -> Option<V>
```

1. 在`insert`时没有将HashMap对象设为可变引用