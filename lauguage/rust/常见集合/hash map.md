## 新建
像 vector 一样，哈希 map 将它们的数据储存在堆上

```rust
use std::collections::HashMap;

let mut scores = HashMap::new();

scores.insert(String::from("Blue"), 10);
scores.insert(String::from("Yellow"), 50);


//一个构建哈希 map 的方法是使用一个元组的 vector 的 collect 方法，其中每个元组包含一个键值对
let teams  = vec![String::from("Blue"), String::from("Yellow")];
let initial_scores = vec![10, 50];

let scores: HashMap<_, _> = teams.iter().zip(initial_scores.iter()).collect();
```

## 所有权
对于像 i32 这样的实现了 Copy trait 的类型，其值可以拷贝进哈希 map。对于像 String 这样拥有所有权的值，其值将被移动而哈希 map 会成为这些值的所有者.
如果将值的引用插入哈希 map，这些值本身将不会被移动进哈希 map。但是这些引用指向的值必须至少在哈希 map 有效时也是有效的。

## 访问
可以通过`get`方法并提供对应的键来从哈希 map 中获取值,`get`返回`Option<V>`，所以结果被装进`Some`；如果某个键在哈希 map 中没有对应的值，`get`会返回`None`,这时就要用某种第六章提到的方法之一来处理 Option。
可以使用与 vector 类似的方式来遍历哈希 map 中的每一个键值对，也就是 `for` 循环
```rust
use std::collections::HashMap;

let mut scores = HashMap::new();

scores.insert(String::from("Blue"), 10);
scores.insert(String::from("Yellow"), 50);

let team_name = String::from("Blue");
let score = scores.get(&team_name);

for (key, value) in &scores {
    println!("{}: {}", key, value);
}
```

## 更新
```rust
use std::collections::HashMap;

let mut scores = HashMap::new();

scores.insert(String::from("Blue"), 10);
scores.insert(String::from("Blue"), 25); // 覆盖
scores.entry(String::from("Blue")).or_insert(50); // 如果不存在则插入


let text = "hello world wonderful world";

let mut map = HashMap::new();

// 结合已有值进行新赋值
for word in text.split_whitespace() {
    let count = map.entry(word).or_insert(0);
    *count += 1;
}

println!("{:?}", scores);
```
`Entry`的`or_insert`方法在键对应的值存在时就返回这个值的可变引用，如果不存在则将参数作为新值插入并返回新值的可变引用。这比编写自己的逻辑要简明的多，另外也与借用检查器结合得更好。
