## Slice
另一个没有所有权的数据类型是 slice。slice 允许你引用集合中一段连续的元素序列，而不用引用整个集合。

### 字符串slice
**字符串 slice**（string slice）是 String 中一部分值的引用，它看起来像这样
```rust
fn main() {
let s = String::from("hello world");

let hello = &s[0..5];
let world = &s[6..11];

// let slice = &S[start..end]
// let slice = &s[..end]
// let slice = &s[start..]
}
```
![slice存储结构](https://kaisery.github.io/trpl-zh-cn/img/trpl04-06.svg)
`注意：字符串 slice range 的索引必须位于有效的 UTF-8 字符边界内，如果尝试从一个多字节字符的中间位置创建字符串 slice，则程序将会因错误而退出。`

```rust
fn first_word(s: &String) -> &str {
    let bytes = s.as_bytes();

    for (i, &item) in bytes.iter().enumerate() {
        if item == b' ' {
            return &s[0..i];
        }
    }

    &s[..]
}

fn main() {
    let mut s = String::from("hello world");

    let word = first_word(&s);

    s.clear(); // 错误!

    println!("the first word is: {}", word);  // word为s的切片引用，在上一行中s已被清理
}

```

所有权、借用和 slice 这些概念让 Rust 程序在编译时确保内存安全。Rust 语言提供了跟其他系统编程语言相同的方式来控制你使用的内存，但拥有数据所有者在离开作用域后自动清除其数据的功能意味着你无须额外编写和调试相关的控制代码。