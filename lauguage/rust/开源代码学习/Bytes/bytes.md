## 存储结构
一片连续的引用计数的内存，可以有效的保存和操作连续的内存。主要用于网络类型代码。
允许多个Bytes对象使用相同的低层内存，并且通过引用计数跟踪内存是否可以释放或不再使用。
bytes内存布局如下
```text
+-------+
| Bytes |
+-------+
 /      \_____
|              \
v               v
+-----+------------------------------------+
| Arc |         |      Data     |          |
+-----+------------------------------------+
```

```rust
pub struct Bytes {
    inner: Inner,
}

#[cfg(target_endian = "little")]
#[repr(C)]
struct Inner {
    // WARNING: Do not access the fields directly unless you know what you are
    // doing. Instead, use the fns. See implementation comment above.
    arc: AtomicPtr<Shared>,
    ptr: *mut u8,
    len: usize,
    cap: usize,
}
```

## bytes方法

### clone
```rust
impl Clone for Bytes {
    fn clone(&self) -> Bytes {
        Bytes {
            inner: unsafe { self.inner.shallow_clone(false) },
        }
    }
}
```