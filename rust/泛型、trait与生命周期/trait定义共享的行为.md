trait 告诉 Rust 编译器某个特定类型拥有可能与其他类型共享的功能。可以通过 trait 以一种抽象的方式定义共享的行为。可以使用 trait bounds 指定泛型是任何拥有特定行为的类型

## 定义trait
```rust
// src/lib.rs

pub trait Summary {
    fn summarize(&self) -> String;
}
```
这里使用 `trait`关键字来声明一个trait，后面是trait 名字，在这个例子中是 `Summary`。在大括号中声明描述实现这个trait的类型所需要的行为的方法签名.
在方法签名后跟分号，而不是在大括号中提供其实现。接着每一个实现这个 `trait` 的类型都需要提供其自定义行为的方法体，编译器也会确保任何实现 `Summary` trait 的类型都拥有与这个签名的定义完全一致的 `summarize` 方法。
trait 体中可以有多个方法：一行一个方法签名且都以分号结尾。

## 为类型实现trait

```rust
pub struct NewsArticle {
    pub headline: String,
    pub location: String,
    pub author: String,
    pub content: String,
}

impl Summary for NewsArticle {
    fn summarize(&self) -> String {
        format!("{}, by {} ({})", self.headline, self.author, self.location)
    }
}

pub struct Tweet {
    pub username: String,
    pub content: String,
    pub reply: bool,
    pub retweet: bool,
}

impl Summary for Tweet {
    fn summarize(&self) -> String {
        format!("{}: {}", self.username, self.content)
    }
}
```
别人想要利用我们 crate 的功能为其自己的库作用域中的结构体实现 `Summary` trait。首先他们需要将 trait 引入作用域。这可以通过指定 `use aggregator::Summary`; 实现，这样就可以为其类型实现 `Summary` trait 了。`Summary` 还必须是公有 trait 使得其他 crate 可以实现它。
实现 trait 时需要注意的一个限制是，只有当 trait 或者要实现 trait 的类型位于 crate 的本地作用域时，才能为该类型实现 trait。但是不能为外部类型实现外部 trait。
这个限制是被称为 相干性（coherence） 的程序属性的一部分，或者更具体的说是 孤儿规则（orphan rule），其得名于不存在父类型。这条规则确保了其他人编写的代码不会破坏你代码，反之亦然。

## 默认实现
有时为 trait 中的某些或全部方法提供默认的行为，而不是在每个类型的每个实现中都定义自己的行为是很有用的。这样当为某个特定类型实现 trait 时，可以选择保留或重载每个方法的默认行为。
```rust
pub trait Summary {
    fn summarize(&self) -> String {
        String::from("(Read more...)")
    }
}
```
如果想要对 `NewsArticle` 实例使用这个默认实现，而不是定义一个自己的实现，则可以通过 `impl Summary for NewsArticle {} `指定一个空的 `impl` 块

默认实现允许调用相同 trait 中的其他方法，哪怕这些方法没有默认实现。如此，trait 可以提供很多有用的功能而只需要实现指定一小部分内容。
```rust
pub trait Summary {
    fn summarize_author(&self) -> String;

    fn summarize(&self) -> String {
        format!("(Read more from {}...)", self.summarize_author())
    }
}
```

## trait 作为参数
```rust
pub fn notify(item: impl Summary) {
    println!("Breaking news! {}", item.summarize());
}
```
对于 `item` 参数，我们指定了 `impl` 关键字和 trait 名称，而不是具体的类型。

## Trait Bound 语法
```rust
pub fn notify<T: Summary>(item: T) {
    println!("Breaking news! {}", item.summarize());
}
```
impl Trait 很方便，适用于短小的例子。trait bound 则适用于更复杂的场景。例如，需要将两个参数限制为相同类型时
```rust
pub fn notify(item1: impl Summary, item2: impl Summary) {
pub fn notify<T: Summary>(item1: T, item2: T) { 
```

### 通过 + 指定多个 trait bound
如果 `notify` 需要显示 `item` 的格式化形式，同时也要使用 `summarize` 方法，那么 `item` 就需要同时实现两个不同的 trait：`Display` 和 `Summary`。这可以通过 `+` 语法实现
```rust
pub fn notify(item: impl Summary + Display) {
```

`+`语法也适用于泛型的 trait bound
```rust
pub fn notify<T: Summary + Display>(item: T) {
```

### 通过 where 简化 trait bound
然而，使用过多的 trait bound 也有缺点。每个泛型有其自己的 trait bound，所以有多个泛型参数的函数在名称和参数列表之间会有很长的 trait bound 信息，这使得函数签名难以阅读。为此，Rust 有另一个在函数签名之后的 `where` 从句中指定 trait bound 的语法。

```rust
fn some_function<T, U>(t: T, u: U) -> i32
    where T: Display + Clone,
          U: Clone + Debug
{
```

## 返回实现了 trait 的类型
也可以在返回值中使用 `impl Trait` 语法，来返回实现了某个 trait 的类型
```rust
fn returns_summarizable() -> impl Summary {
    Tweet {
        username: String::from("horse_ebooks"),
        content: String::from("of course, as you probably already know, people"),
        reply: false,
        retweet: false,
    }
}
```

不过这只适用于返回单一类型的情况

## 使用 trait bound 有条件地实现方法
通过使用带有 trait bound 的泛型参数的 `impl` 块，可以有条件地只为那些实现了特定 trait 的类型实现方法。
例如，示例 10-16 中的类型 `Pair<T>` 总是实现了 new 方法，不过只有那些为 `T` 类型实现了 `PartialOrd` trait （来允许比较） 和 `Display` trait （来启用打印）的 `Pair<T>` 才会实现 `cmp_display`
```rust

use std::fmt::Display;

struct Pair<T> {
    x: T,
    y: T,
}

impl<T> Pair<T> {
    fn new(x: T, y: T) -> Self {
        Self {
            x,
            y,
        }
    }
}

impl<T: Display + PartialOrd> Pair<T> {
    fn cmp_display(&self) {
        if self.x >= self.y {
            println!("The largest member is x = {}", self.x);
        } else {
            println!("The largest member is y = {}", self.y);
        }
    }
}
```

trait 和 trait bound 让我们使用泛型类型参数来减少重复，并仍然能够向编译器明确指定泛型类型需要拥有哪些行为。因为我们向编译器提供了 trait bound 信息，它就可以检查代码中所用到的具体类型是否提供了正确的行为。在动态类型语言中，如果我们尝试调用一个类型并没有实现的方法，会在运行时出现错误。Rust 将这些错误移动到了编译时，甚至在代码能够运行之前就强迫我们修复错误。