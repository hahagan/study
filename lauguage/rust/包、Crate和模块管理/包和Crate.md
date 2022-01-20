## 概念
crate 是一个二进制项或者库。crate root 是一个源文件，Rust 编译器以它为起始点，并构成你的 crate 的根模块。
包（package） 是提供一系列功能的一个或者多个 crate。一个包会包含有一个 Cargo.toml 文件，阐述如何去构建这些 crate。

## Crate
一个 crate 会将一个作用域内的相关功能分组到一起，使得该功能可以很方便地在多个项目之间共享。
`rand`crate 提供的所有功能都可以通过该crate的名字：`rand`进行访问。
将一个crate的功能保持在其自身的作用域中，可以知晓一些特定的功能是在我们的crate中定义的还是在 `rand`crate 中定义的，这可以防止潜在的冲突。

## 包
包中所包含的内容由几条规则来确立：
* 一个包中至多 只能 包含一个库 crate(library crate)
* 包中可以包含任意多个二进制 crate(binary crate)
* 包中至少包含一个 crate，无论是库的还是二进制的

约定：
* 一个Cargo包中的src/main.rs 就是一个与包同名的二进制 crate 的 crate 根
* Cargo 知道如果包目录中包含 src/lib.rs，则包带有与其同名的库 crate，且 src/lib.rs 是 crate 根。crate 根文件将由 Cargo 传递给 rustc 来实际构建库或者二进制项目。