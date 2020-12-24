数据包内数据为大端存储

### Event memory layout

```
struct event {
	message_size: u32,
	map_count: u32,
	kv: Vec<kv>
}

srtuct kv {
	len_key: u32,
	key: bytes,
	'0': u8,
	len_value: u32,
	value: bytes
	'0'
}
```

### connect

```
1. send sig
	send splunkSignature to splunk 
2. read and discard
	1. connect to splunk and set 10ms time out
	2. read 4096 bytes from connetct
	3. 
```



```
type splunkSignature struct {
	signature  [128]byte	// --splunk-cooked-mode-v2-- 
	serverName [256]byte	// serverName, splunk host
	mgmtPort   [16]byte		// mgmtPort, splunk accept port
}
```



