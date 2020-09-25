## API
```
— Constant packet.max_payload

The maximum payload length of a packet.

— Function packet.allocate

Returns a new empty packet. An an error is raised if there are no packets left on the freelist. Initially the length of the allocated is 0, and its data is uninitialized garbage.

— Function packet.free packet

Frees packet and puts in back onto the freelist.

— Function packet.clone packet

Returns an exact copy of packet.

— Function packet.resize packet, length

Sets the payload length of packet, truncating or extending its payload. In the latter case the contents of the extended area at the end of the payload are filled with zeros.

— Function packet.append packet, pointer, length

Appends length bytes starting at pointer to the end of packet. An error is raised if there is not enough space in packet to accomodate length additional bytes.

— Function packet.prepend packet, pointer, length

Prepends length bytes starting at pointer to the front of packet, taking ownership of the packet and returning a new packet. An error is raised if there is not enough space in packet to accomodate length additional bytes.

— Function packet.shiftleft packet, length

Take ownership of packet, truncate it by length bytes from the front, and return a new packet. Length must be less than or equal to length of packet.

— Function packet.shiftright packet, length

Take ownership of packet, moves packet payload to the right by length bytes, growing packet by length. Returns a new packet. The sum of length and length of packet must be less than or equal to packet.max_payload.

— Function packet.from_pointer pointer, length

Allocate packet and fill it with length bytes from pointer.

— Function packet.from_string string

Allocate packet and fill it with the contents of string.

— Function *packet.clone_to_memory pointer packet

Creates an exact copy of at memory pointed to by pointer. Pointer must point to a packet.packet_t.
```

## 数据结构
```c++
/* Use of this source code is governed by the Apache 2.0 license; see COPYING. */

// The maximum amount of payload in any given packet.
enum { PACKET_PAYLOAD_SIZE = 10*1024 };

// Packet of network data, with associated metadata.
struct packet {
    uint16_t length;           // data payload length
    unsigned char data[PACKET_PAYLOAD_SIZE];
};
```

## 使用
packet用于存储当前正在处理的数据。
1. 每个packet必须有明确的生命周期。
2. 数据包通过两个接口明确的分配和释放。
3. 通过`link.receive`接口可以获取数据包的所有权。
4. app必须确保通过`link.trnasmit`将数据包将数据所有权传递给其他app或通过`free`释放数据包
5. app仅能在数据包没有被`transmit`或`free`前使用
6. 数据包分配，从一个数据包池中进行分配

## 源码
### 空闲列表
packets_fl为在"engine/packets.freelist"创建的一个共享内存中的一个"struct freelist",用于保存空闲的packet地址。对外接口allocate()分配packet时会从packets_fl为在中保存的packet中分配
```c++
struct freelist {
    int32_t lock[1];
    uint64_t nfree;
    uint64_t max;
    struct packet *list[max_packets];
};
```

### packet分配与释放
`local function freelist_remove`: 从空闲列表中获取一个packet存储对象，并将空余数量减一
`local function freelist_add(freelist, element)`: element为一个packets对象指针，将element赋值加入freelist，并且使得空闲数量加一

`function allocate()`: 如果group_fl有空闲元素，则将其加入packets_fl中，最后返回一个packet_fl的列表对象，并将packet_fl剩余减一。对外开发的packet分配函数。

`function new_packet()`: 从DMA内存中申请一个空packet存储，并返回其指针
`function preallocate_step()`: 从DMA内存中申请空packet存储并放入packet_fl中，同时确保总申请数量不会超出限制
`function free_internal(p)`: 内部释放packet，并将其存入packet_fl。

```lua
function allocate ()
   if freelist_nfree(packets_fl) == 0 then
      if group_fl then
         freelist_lock(group_fl)
         while freelist_nfree(group_fl) > 0
         and freelist_nfree(packets_fl) < packets_allocated do
            freelist_add(packets_fl, freelist_remove(group_fl))
         end
         freelist_unlock(group_fl)
      end
      if freelist_nfree(packets_fl) == 0 then
         preallocate_step()
      end
   end
   return freelist_remove(packets_fl)
end
```