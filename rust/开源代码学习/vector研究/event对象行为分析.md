## event数据结构
```rust
pub enum Event {
    Log(LogEvent),
    Metric(Metric),
}

#[derive(PartialEq, Debug, Clone, Default)]
pub struct LogEvent {
    fields: BTreeMap<String, Value>,
}

#[derive(Debug, Clone, PartialEq, Deserialize, Serialize)]
pub struct Metric {
    pub name: String,
    pub timestamp: Option<DateTime<Utc>>,
    pub tags: Option<BTreeMap<String, String>>,
    pub kind: MetricKind,
    #[serde(flatten)]
    pub value: MetricValue,
}

#[derive(PartialEq, Debug, Clone, is_enum_variant)]
pub enum Value {
    Bytes(Bytes),
    Integer(i64),
    Float(f64),
    Boolean(bool),
    Timestamp(DateTime<Utc>),
    Map(BTreeMap<String, Value>),
    Array(Vec<Value>),
    Null,
}
```

## LogEvent
Event分为两种类型，LogEvent(以后简称log)类型由一个`BTreeMap`存储，其key为`string`类型代表一个event的字段名，Value为底层多种数据的枚举代表一个字段的值。

### Value方法
源码位于"event/value.rs"
在Value上的方法主要有
* `pub fn From<T>(T) -> self`: 该函数用于从各种类型中生成一个Value实例。注意在这里vector并没有实现泛型的From，文中这么写仅未来代表各种数据类型
* `fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>`: 将Value值序列化为各种其他类型的数据
* `fn try_into(self) -> Result<serde_json::Value, Self::Error>`: 尝试将Value实例序列化为json的Value对象
* `pub fn to_string_lossy(&self) -> String`: 转换为`string`对象
* `pub fn as_bytes(&self) -> Bytes`: 将Value实例根据不同类型按不同方法转换为`bytes`对象
* `pub fn into_bytes(self) -> Bytes`: 将Value实例根据不同类型按不同方法转换为`bytes`对象，直接调用了`as_bytes`函数，所以两者行为基本相同。
* `pub fn as_timestamp(&self) -> Option<&DateTime<Utc>>`: 转换为时间引用
