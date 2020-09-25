## regexParser数据结构
```rust
pub struct RegexParser {
    regexset: RegexSet,             // 支持同时进行多个正则匹配
    patterns: Vec<CompiledRegex>, // indexes correspend to RegexSet
    field: Atom,                  // 解析字段
    drop_field: bool,             // 解析完成后是否丢弃原字段
    drop_failed: bool,            
    target_field: Option<Atom>,   // 解析出来的字段将会作为目标字段的子字段
    overwrite_target: bool,       // 是否覆盖已存在字段
}

struct CompiledRegex {
    regex: Regex,               // 正则表达式
    capture_names: Vec<(usize, Atom, Conversion)>, 
    capture_locs: CaptureLocations,
}
```

## 数据处理逻辑解析
由Vector框架分析可知，每个Transform对象的数据处理函数为`fn transform(&mut self, mut event: Event) -> Option<Event>`
其大致处理流程如下：
1. 获取解析字段值
2. 对解析字段进行正则匹配
    1. 如果未成功匹配则结束则返回当前事件或返回空
3. 匹配成功后，获取保存的字段
4. 获取匹配结果
5. 如果配置中指明需要覆盖已有字段，则会发生已有字段的删除
6. 最后将匹配结果插入event中

## Tramsform源码
```rust
impl Transform for RegexParser {
    fn transform(&mut self, mut event: Event) -> Option<Event> {
        let log = event.as_mut_log();
        let value = log.get(&self.field).map(|s| s.as_bytes());  
        emit!(RegexParserEventProcessed);

        if let Some(value) = &value {
            let regex_id = self.regexset.matches(&value).into_iter().next();
            let id = match regex_id {
                Some(id) => id,
                None => {
                    emit!(RegexParserFailedMatch { value });
                    return if self.drop_failed { None } else { Some(event) };
                }
            };

            let target_field = self.target_field.as_ref();

            let pattern = self
                .patterns
                .get_mut(id)
                .expect("Mismatch between capture patterns and regexset");

            if let Some(captures) = pattern.captures(&value) {
                // Handle optional overwriting of the target field
                if let Some(target_field) = target_field {
                    if log.contains(target_field) {
                        if self.overwrite_target {
                            log.remove(target_field);
                        } else {
                            emit!(RegexParserTargetExists { target_field });
                            return Some(event);
                        }
                    }
                }

                log.extend(captures.map(|(name, value)| {
                    let name = target_field
                        .map(|target| Atom::from(format!("{}.{}", target, name)))
                        .unwrap_or_else(|| name.clone());
                    (name, value)
                }));
                if self.drop_field {
                    log.remove(&self.field);
                }
                return Some(event);
            }
        } else {
            emit!(RegexParserMissingField { field: &self.field });
        }

        if self.drop_failed {
            None
        } else {
            Some(event)
        }
    }
}
```
