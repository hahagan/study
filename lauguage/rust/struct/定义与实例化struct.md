## 定义与实例化
```rust
// 定义
struct StructName {
    field_name: Type,
    field_name1: Type1,
}

struct User {
    username: String,
    email: String,
    sign_in_count: u64,
    active: bool,
}

// 实例化
let user = User {
    email: String::from("someone@123.com"),
    username: String::from("someone"),
    actice: true,
    sign_in_count: 1,
}

// 可变结构体
let mut user1 = User {
    email: String::from("someone@example.com"),
    username: String::from("someusername123"),
    active: true,
    sign_in_count: 1,
};

user1.email = String::from("anotheremail@example.com");

// 字段初始化简写法
fn build_user(email: String, username: String) -> User {
    User {
        email,
        username,
        active: true,
        sign_in_count: 1,
    }
}

// 通过派生trait增加实用功能
#[derive(Debug)]
struct Rectangle {
    width: u32,
    height: u32,
}
```

结构体更新语法
```rust
fn main() {
struct User {
    username: String,
    email: String,
    sign_in_count: u64,
    active: bool,
}

let user1 = User {
    email: String::from("someone@example.com"),
    username: String::from("someusername123"),
    active: true,
    sign_in_count: 1,
};

let user2 = User {
    email: String::from("another@example.com"),
    username: String::from("anotherusername567"),
    active: user1.active,                           // 值来自user1
    sign_in_count: user1.sign_in_count,             // 值来自user1
};


let user2 = User {
    email: String::from("another@example.com"),
    username: String::from("anotherusername567"),
    ..user1
};
}

```

## 元组结构体
元组结构体有着结构体名称提供的含义，但没有具体的字段名，只有字段的类型。

```rust
fn main() {
struct Color(i32, i32, i32);
struct Point(i32, i32, i32);

let black = Color(0, 0, 0);
let origin = Point(0, 0, 0);
}
```
在其他方面，元组结构体实例类似于元组：可以将其解构为单独的部分，也可以使用`.`后跟索引来访问单独的值。注意虽然Color与Point两者成员类型相同，但是它们的实例不可以相互替换，它们是不同的数据类型。

## 没有任何字段的类单元结构体
我们也可以定义一个没有任何字段的结构体！它们被称为**类单元结构体**（unit-like structs）因为它们类似于 ()，即 unit 类型。类单元结构体常常在你想要在某个类型上实现 trait 但不需要在类型中存储数据的时候发挥作用。

## 结构体数据的所有权
`User`结构体的定义中，我们使用了自身拥有所有权的`String`类型而不是`&str`字符串`slice`类型。这是一个有意而为之的选择，因为我们想要这个结构体拥有它所有的数据，为此只要整个结构体是有效的话其数据也是有效的。

可以使结构体存储被其他对象拥有的数据的引用，不过这么做的话需要用上**生命周期**（lifetimes），这是一个第十章会讨论的 Rust 功能。生命周期确保结构体引用的数据有效性跟结构体本身保持一致。如果你尝试在结构体中存储一个引用而不指定生命周期将是无效的，比如这样：

```rust
// 这些代码不能编译！
struct User {
    username: &str,
    email: &str,
    sign_in_count: u64,
    active: bool,
}

fn main() {
    let user1 = User {
        email: "someone@example.com",
        username: "someusername123",
        active: true,
        sign_in_count: 1,
    };
```