### 描述

splunk tcp转发分为`splunktcp`类型与原始tcp类型，前者为splunk在tcp协议上完成的特殊应用层协议，需要破解splunktcp的V3协议。

原始tcp协议转发，splunk会转发所有splunk内的日志，包含splunk接收的日志和自身生成的日志。



例如 splunk接收的日志内容为

```
test tcp forward0
test tcp forward1
test tcp forward2
test tcp forward3
```

经过数据处理后splunk的event表现为

```
{"source": "192.168.84.80:9090", "_raw":"test tcp forward0", "_time": "......"}
{"source": "192.168.84.80:9090", "_raw":"test tcp forward1", "_time": "......"}
{"source": "192.168.84.80:9090", "_raw":"test tcp forward1", "_time": "......"}
{"source": "192.168.84.80:9090", "_raw":"test tcp forward1", "_time": "......"}
```



#### 原始tcp

那么配置splunk 原始tcp转发后，转发的日志内容如下,无法转发event，仅为原始日志内容。并且尝试时无法将其他日志剥离。

```
test tcp forward0
<web assece log0>
<some oprate log0>
<some statuts log0>
......

test tcp forward1
<some statuts log0>
......

test tcp forward2
<some statuts log2>
<some oprate log2>
......

test tcp forward3
<some oprate log3>
......
```



#### splunk tcp

splunktcp类型的转发，为splunk在tcp协议上完成的应用层协议，需要完成对splunktcp的接收协议方能接收对应的数据。同样的数据，使用splunktcp进行转发。接收方使用logstash的tcp接收，接收到的数据内容如下，在splunk页面可以观测到web报出警告，大意为""转发下游不可用"

```
{"message": "...splunk mod v3...", "@timestamp": ".....", "port": "<sp发送端口>"
{"message": "...splunk mod v3...", "@timestamp": ".....", "port": "<sp发送端口>"
{"message": "...splunk mod v3...", "@timestamp": ".....", "port": "<sp发送端口>"
{"message": "...splunk mod v3...", "@timestamp": ".....", "port": "<sp发送端口>"
{"message": "...splunk mod v3...", "@timestamp": ".....", "port": "<sp发送端口>"
{"message": "...splunk mod v3...", "@timestamp": ".....", "port": "<sp发送端口>"
{"message": "...splunk mod v3...", "@timestamp": ".....", "port": "<sp发送端口>"
{"message": "...splunk mod v3...", "@timestamp": ".....", "port": "<sp发送端口>"
{"message": "...splunk mod v3...", "@timestamp": ".....", "port": "<sp发送端口>"
{"message": "...splunk mod v3...", "@timestamp": ".....", "port": "<sp发送端口>"
```

接收方使用其他splunk实例，可在页面上搜索到新的日志内容



### 测试步骤

#### 创建splunk实例

创建两个splunk实例，分别用于接收数据并转发与接收转发。

```
// 接收转发
docker run -d -p 8000:8000 -p 5140:5140 -e "SPLUNK_START_ARGS=--accept-license" -e "SPLUNK_PASSWORD=12345678" --name sp_ac splunk/splunk:latest

// 接收并转发
docker run -d --network host -e "SPLUNK_START_ARGS=--accept-license" -e "SPLUNK_PASSWORD=12345678" --name sp_host splunk/splunk:latest

```



#### 测试原始转发

##### 配置splunk原始tcp转发

1. 创建接收

```
docker exec -it sp_host bash 
cd etc/system/local

## 修改配置文件，接收数据
vi inputs.conf

## inputs.conf文件内容如下
[tcp://5140]


[tcp://5140]
connection_host = dns
index = tcp-ac
source = ac
sourcetype = ac
_TCP_ROUTING = g1

```

```
## inputs.conf文件内容如下
[tcp://5140]


[tcp://5140]
connection_host = dns
index = tcp-ac
source = ac
sourcetype = ac
_TCP_ROUTING = g1
```

2. 配置转发

```
## 修改配置文件，配置转发
vi outputs.conf

## 文件内容如下
[tcpout:g1]
server = ip:6666

[tcpout-server://ip:6666]
disable=0
```

3. 重启容器或通过更新api使得配置生效
4. 创建一个logstash实例接收6666的tcp端口，略。
5. 发送数据，观测logstash输出

#### 测试splunktcp 转发

1. 基于前一测试的基础上，将ouput.conf部分配置删除，在splunk页面按顺序 "数据转发与接收" -> "tcp转发" -> "新建转发"，配置转发目的端口与目的地址。此时在`app/search/local/outputs.conf`目录或上文中的"outputs.conf"目录下会添加一个`splunktcp://ip:port`对象。

2. 如果此时测试的是logstash接收，则可以进行数据发送，并观测tcp接收的输出

3. 如果此时测试的是其他splunk的接收行为，进行一下步骤

   1. 进入sp_ac容器。`docker exec -it sp_ac bash`

   2. 修改文件`etc/system/local/inputs.conf`，创建splunktcp接收，内容如下

      ```
      [splunktcp:://6666]
      index = tcp-ac
      source = ac
      ```

   3.  重启容器或接口热更新splunk

4. 发送数据并观测web页面可搜索日志

5. 如果需要可以对特定端口，特定网卡进行抓包，可以进一步观测splunk间相互转发的数据内容，其根据一定格式在两者间进行数据交换，并且会发送额外的控制信号作为应用层的协议行为。