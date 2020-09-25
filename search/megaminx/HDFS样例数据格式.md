## 通用部分
{时间} {未知} {日志级别} dfs.{系统}${对象} {日志内容}

时间(strptime格式): %y%m%d %H%M%S
未知: 该字段含义未知
日志级别(log_level): 英文字符
系统: 目前发现值有
* DataNode
* FSNamesystem
* DataBlockScanner
* FSDataset
对象: 根据系统不同，有不同对象，甚至可能没有
* DataNode(系统类型)含有对象
    * DataXceiver
    * PacketResponder

日志内容: 根据日志对象不同，日志内容不同



re:
(?P<time>\d{6}\s+\d{6})\s*(?P<u1>\d+)\s*(?P<level>\w+)\s*dfs.(?P<device>[\w]+)(\$(?P<object>\w+))*:(?P<ctx>.+)*

## 