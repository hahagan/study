## API
```
— Function link.empty link

Predicate used to test if a link is empty. Returns true if link is empty and false otherwise.

— Function link.full link

Predicate used to test if a link is full. Returns true if link is full and false otherwise.

— Function link.nreadable link

Returns the number of packets on link.

— Function link.nwriteable link

Returns the remaining number of packets that fit onto link.

— Function link.receive link

Returns the next available packet (and advances the read cursor) on link. If the link is empty an error is signaled.

— Function link.front link

Return the next available packet without advancing the read cursor on link. If the link is empty, nil is returned.

— Function link.transmit link, packet

Transmits packet onto link. If the link is full packet is dropped (and the drop counter increased).

— Function link.stats link
```

## 数据结构
```c++
enum { LINK_RING_SIZE    = 1024,
       LINK_MAX_PACKETS  = LINK_RING_SIZE - 1
};
struct link {
  // this is a circular ring buffer, as described at:
  //   http://en.wikipedia.org/wiki/Circular_buffer
  struct packet *packets[LINK_RING_SIZE];
  struct {
    struct counter *dtime, *txbytes, *rxbytes, *txpackets, *rxpackets, *txdrop;
  } stats;
  // Two cursors:
  //   read:  the next element to be read
  //   write: the next element to be written
  int read, write;
};
```
