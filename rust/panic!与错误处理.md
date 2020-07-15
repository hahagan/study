## panic!与不可恢复错误

### panic 栈的展开或终止
当出现 panic 时，程序默认会开始 **展开**（unwinding），这意味着 Rust 会回溯栈并清理它遇到的每一个函数的数据，不过这个回溯并清理的过程有很多工作。另一种选择是直接 **终止**（abort），这会不清理数据就退出程序。那么程序所使用的内存需要由操作系统来清理。如果你需要项目的最终二进制文件越小越好，panic 时通过在 Cargo.toml 的 `[profile]` 部分增加 `panic = 'abort'`，可以由展开切换为终止。

### 获取backtrace
backtrace 是一个执行到目前位置所有被调用的函数的列表。可以设`RUST_BACKTRACE` 环境变量来得到一个 backtrace。
为了获取带有这些信息的 backtrace，必须启用 debug 标识。当不使用 --release 参数运行 cargo build 或 cargo run 时 debug 标识会默认启用。


## Result与可恢复错误

### Result
```rust
enum Result<T, E> {
    Ok(T),
    Err(E),
}
```

这里我们告诉 Rust 当结果是 `Ok` 时，返回 `Ok` 成员中的 `file` 值，然后将这个文件句柄赋值给变量 f。`match` 之后，我们可以利用这个文件句柄来进行读写。

`match` 的另一个分支处理从 `File::open` 得到 `Err` 值的情况。在这种情况下，我们选择调用 panic! 宏。
```rust
use std::fs::File;

fn main() {
    let f = File::open("hello.txt");

    let f = match f {
        Ok(file) => file,
        Err(error) => {
            panic!("Problem opening the file: {:?}", error)
        },
    };
}
```

### 匹配不同错误
```rust
use std::fs::File;
use std::io::ErrorKind;

fn main() {
    let f = File::open("hello.txt");

    let f = match f {
        Ok(file) => file,
        Err(error) => match error.kind() {
            ErrorKind::NotFound => match File::create("hello.txt") {
                Ok(fc) => fc,
                Err(e) => panic!("Problem creating the file: {:?}", e),
            },
            other_error => panic!("Problem opening the file: {:?}", other_error),
        },
    };
}

```

### 失败时 panic 的简写：unwrap 和 expect
#### unwrap
`match` 能够胜任它的工作，不过它可能有点冗长并且不总是能很好的表明其意图。`Result<T, E>` 类型定义了很多辅助方法来处理各种情况。其中之一叫做 `unwrap`。如果 `Result` 值是成员 `Ok`，`unwrap` 会返回 `Ok` 中的值。如果 `Result` 是成员 `Err`，`unwrap` 会为我们调用 `panic!`。
```rust
use std::fs::File;

fn main() {
    let f = File::open("hello.txt").unwrap();
}
```
#### expect
类似于 `unwrap` 的方法它还允许我们选择 `panic!` 的错误信息：`expect`。使用 `expect` 而不是 `unwrap` 并提供一个好的错误信息可以表明你的意图并更易于追踪 panic 的根源
```rust
use std::fs::File;

fn main() {
    let f = File::open("hello.txt").expect("Failed to open hello.txt");
}
```

### 传播错误
除了在这个函数中处理错误外，还可以选择让调用者知道这个错误并决定该如何处理。这被称为 **传播**（propagating）错误

```rust
use std::io;
use std::io::Read;
use std::fs::File;

fn read_username_from_file() -> Result<String, io::Error> {
    let f = File::open("hello.txt");

    let mut f = match f {
        Ok(file) => file,
        Err(e) => return Err(e),
    };

    let mut s = String::new();

    match f.read_to_string(&mut s) {
        Ok(_) => Ok(s),
        Err(e) => Err(e),
    }
}
```

#### 传播错误的简写:?运算符
```rust
use std::io;
use std::io::Read;
use std::fs::File;

fn read_username_from_file() -> Result<String, io::Error> {
    let mut f = File::open("hello.txt")?;
    let mut s = String::new();
    f.read_to_string(&mut s)?;
    Ok(s)
}
```
`Result` 值之后的 `?` 被定义为与示例中定义的处理 `Result` 值的 `match` 表达式有着完全相同的工作方式。如果 `Result` 的值是 `Ok`，这个表达式将会返回 `Ok` 中的值而程序将继续执行。如果值是 `Err`，`Err` 中的值将作为整个函数的返回值，就好像使用了 `return` 关键字一样，这样错误值就被传播给了调用者。
`?` 运算符所使用的错误值被传递给了 `from` 函数，它定义于标准库的 `From` trait 中，其用来将错误从一种类型转换为另一种类型。当 `?` 运算符调用 `from` 函数时，收到的错误类型被转换为由当前函数返回类型所指定的错误类型。只要每一个错误类型都实现了 `from` 函数来定义如何将自身转换为返回的错误类型，`?` 运算符会自动处理这些转换。

```rust
use std::fs::File;

fn main() {
    let f = File::open("hello.txt")?;
}

// main不返回Result类型，因此无法通过编译
```
当你期望在不返回 `Result` 的函数中调用其他返回 `Result` 的函数时使用 `?` 的话，有两种方法修复这个问题。一种技巧是将函数返回值类型修改为 `Result<T, E>`，如果没有其它限制阻止你这么做的话。另一种技巧是通过合适的方法使用 `match` 或 `Result` 的方法之一来处理 `Result<T, E>`

## panic!或传播错误
如果代码 panic，就没有恢复的可能。你可以选择对任何错误场景都调用 panic!，不管是否有可能恢复，不过这样就是你代替调用者决定了这是不可恢复的。选择返回 Result 值的话，就将选择权交给了调用者，而不是代替他们做出决定。

### 错误处理指导原则
在当有可能会导致有害状态的情况下建议使用 panic! —— 在这里，有害状态是指当一些假设、保证、协议或不可变性被打破的状态，例如无效的值、自相矛盾的值或者被传递了不存在的值 —— 外加如下几种情况：
* 有害状态并不包含 预期 会偶尔发生的错误
* 在此之后代码的运行依赖于不处于这种有害状态
* 当没有可行的手段来将有害状态信息编码进所使用的类型中的情况

然而当错误预期会出现时，返回 Result 仍要比调用 panic! 更为合适。
当代码对值进行操作时，应该首先验证值是有效的，并在其无效时 panic!

### 创建自定义类型进行有效性验证
可以创建一个新类型来将验证放入创建其实例的函数中，而不是到处重复这些检查。这样就可以安全的在函数签名中使用新类型并相信他们接收到的值。
```rust

#![allow(unused)]
pub struct Guess {
    value: i32,
}

impl Guess {
    pub fn new(value: i32) -> Guess {
        if value < 1 || value > 100 {
            panic!("Guess value must be between 1 and 100, got {}.", value);
        }

        Guess {
            value
        }
    }

    pub fn value(&self) -> i32 {
        self.value
    }
}
```