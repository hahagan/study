## 用法
可以一次性将路径引入作用域，然后使用`use`关键字调用该路径中的项，就如同它们是本地项一样。
```rust
mod front_of_house {
    pub mod hosting {
        pub fn add_to_waitlist() {}
    }
}

use crate::front_of_house::hosting;

pub fn eat_at_restaurant() {
    hosting::add_to_waitlist();
    hosting::add_to_waitlist();
    hosting::add_to_waitlist();
}

```

## as提供新名称
```rust
use std::fmt::Result;
use std::io::Result as IoResult;

fn function1() -> Result {
    // --snip--
    Ok(())
}

fn function2() -> IoResult<()> {
    // --snip--
    Ok(())
}
```

## 使用 pub use 重导出名称
当使用`use`关键字将名称导入作用域时，在新作用域中可用的名称是私有的。如果为了让调用你编写的代码的代码能够像在自己的作用域内引用这些类型，可以结合`pub`和`use`。这个技术被称为 “重导出（re-exporting）”，因为这样做将项引入作用域并同时使其可供其他代码引入自己的作用域。
```rust
mod front_of_house {
    pub mod hosting {
        pub fn add_to_waitlist() {}
    }
}

pub use crate::front_of_house::hosting;

pub fn eat_at_restaurant() {
    hosting::add_to_waitlist();
    hosting::add_to_waitlist();
    hosting::add_to_waitlist();
}
```

## 使用外部包
在Cargo.toml中加入`rand`依赖告诉了 Cargo 要从 crates.io 下载`rand`和其依赖，并使其可在项目代码中使用。并在代码中通过use引用其作用域的项
```rust
use rand::Rng;

fn main() {
    let secret_number = rand::thread_rng().gen_range(1, 101);
}

```

## 嵌套路径消除use行
```rust
use std::{cmp::Ordering, io};

use std::io::{self, Write}; // ==> use std::io;
                            //     use std::io::Write; 
```

## 通过 glob 运算符将所有的公有定义引入作用域
如果希望将一个路径下 所有 公有项引入作用域，可以指定路径后跟 *，glob 运算符
```rust
use std::collections::*;
```