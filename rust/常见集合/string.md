## 什么是字符串
Rust 的核心语言中只有一种字符串类型：`str`，字符串 `slice`，它通常以被借用的形式出现，`&str`。
称作`String`的类型是由标准库提供的，而没有写进核心语言部分，它是可增长的、可变的、有所有权的、UTF-8 编码的字符串类型。

## 新建字符串
```rust
let mut s = String::new();

let data = "initial contents";

let s = data.to_string();

// 该方法也可直接用于字符串字面值：
let s = "initial contents".to_string();

let s = String::from("initial contents");
```

## 更新
`String`的大小可以增加，其内容也可以改变，就像可以放入更多数据来改变`Vec`的内容一样。另外，可以方便的使用 `+` 运算符或 `format!` 宏来拼接 `String` 值

### 使用 push_str 和 push 附加字符串
```rust
let mut s1 = String::from("foo");
let s2 = "bar";
s1.push_str(s2);          // 疑问：为什么s2的所有权没有被方法夺走
println!("s2 is {}", s2); // 疑问：为什么s2的所有权没有被方法夺走

let mut s = String::from("lo");
s.push('l');              // push仅能添加一个单独的字符
```

### 使用 + 运算符或 format! 宏拼接字符串
```rust
let s1 = String::from("Hello, ");
let s2 = String::from("world!");
let s3 = s1 + &s2; // 注意 s1 被移动了，不能继续使用
```
`+`的函数声明类似于
```rust
fn add(self, s: &str) -> String {}
```
之所以能够在 add 调用中使用 &s2 是因为 &String 可以被 强转（coerced）成 &str
当add函数被调用时，Rust 使用了一个被称为 解引用强制多态（deref coercion）的技术，你可以将其理解为它把 &s2 变成了 &s2[..]。
其次，可以发现签名中 add 获取了 self 的所有权，因为 self 没有 使用 &。
换句话说，它看起来好像生成了很多拷贝，不过实际上并没有：这个实现比拷贝要更高效。
在有这么多 `+` 和 `"` 字符的情况下，很难理解具体发生了什么。对于更为复杂的字符串链接，可以使用`format!`宏。format! 与 println! 的工作原理相同，不过不同于将输出打印到屏幕上，它返回一个带有结果内容的 String。这个版本就好理解的多，并且不会获取任何参数的所有权。
```rust
let s1 = String::from("tic");
let s2 = String::from("tac");
let s3 = String::from("toe");

let s = format!("{}-{}-{}", s1, s2, s3);
```

## 索引字符串
在 Rust 中，如果你尝试使用索引语法访问 `String` 的一部分，会出现一个错误。

### 内部表现
```rust
let len = String::from("Hola").len(); // len方法返回的是utf-8编码所需要的字节数
```

最后一个 Rust 不允许使用索引获取 `String` 字符的原因是，索引操作预期总是需要常数时间 (O(1))。但是对于 `String` 不可能保证这样的性能，因为 Rust 必须从开头到索引位置遍历来确定有多少有效的字符

## 字符串slice
```
let hello = "Здравствуйте";

let s = &hello[0..4];
```

## 遍历字符串
```rust
for c in "नमस्ते".chars() {
    println!("{}", c);
}

for b in "नमस्ते".bytes() {
    println!("{}", b);
}
```