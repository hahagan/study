### overview

根据数字数值决定编码后数字所占字节数的一种编码方式。



### 算法原理

* **除了最后一个字节**外，varint编码中的每个字节都设置了最高有效位（most significant bit - msb
* msb为1则表明后面的字节还是属于当前数据的,如果是0那么这是当前数据的最后一个字节数据。
* 每个字节的低7位用于以7位为一组存储数字的二进制补码表示，最低有效组在前，或者叫最低有效字节在前。这表明varint编码后数据的字节是按照**小端序排列**。
  * 例如 `123456`二进制形式为`1 1110 0010 0100 0000`编码流程如下
    1. 取低七位` 100 0000`在前，并设msb为1，即`1100 0000`
    2. 再取七位`0 0010 01`，设msb为1，即`1100 0100`，结合已生成 ``1100 0000 1100 0100`
    3. 剩余不足7位，因此msb为0，不足8位补零。即`0000 0111`
    4. 最终varint编码为`1100 0000 1100 0100 0000 0111`
  * `1100 0000 1100 0100 0000 0111`解码流程
    1. 第一个字节msb为1，因此取剩余七字节`100 0000`
    2. 第二个字节msb为1，因此取剩余七字节并在第一个七字节之前`10 0000 0100 0000`
    3. 第三个字节msb为0,此时取剩余7字节，但不继续下一字节的解码。最终解码后的字节为`00 0001 1110 0000 0100 0000`。此时高位的零位去除后正好为`123456`的原始二进制表示。



### zigzag

ZigZag是将有符号数统一映射到无符号数的一种编码方案。在负数中最高符号位为1影响了varint的压缩，因此zigzag将符号位移动到低位，但由于负数的补码形式有较多的高位为1，因此可以考虑将其取反，从而完成对负数到正数的映射。

所以zigzag的编码算法为`(n << 1) ^ (n>> 31)`,这里假设数字长度为32。将符号位置于低位，并求异或。这里求异或可以将大量的高位置为1。因为负数绝对值较小时，高位为1的数量较多，通过异或可以将其转为0。

解码算法为编码算法的逆向。

```
func TestZigZag(t *testing.T) {
	var x int32 = -1
	r := zigzag32(x)
	r1 := zizagDecode32(r)
	fmt.Printf("%b %b\n", r, r1)
}

func zigzag32(x int32) int32 {
	// 左移一位 XOR (-1 / 0 的 64 位补码)
	fmt.Printf("%b %b, %b\n", x, x<<1, x>>31)
	return (x << 1) ^ (x >> 31)
}

func zizagDecode32(x int32) int32 {
	return (x >> 1) ^ -(x & 1)
}

```

补充
```
计算机中左移与右移操作：
	左移：
		正数：向左移动，低位补零
		负数：向左移动2位，低位补零
		
	右移：
		正数：向右移动2位，高位补0
		负数：向右移动2位,高位补1
	
```


varint的思想是对数值较小的数字进行编码以使用更少的字节数，但负数在字节流表示下其值较大，不利于使用varint的编码方式。此时可以使用zigzag对负数进行映射。







### 代码

```go
import (
	"testing"
)

type Varint struct {
}

func (v *Varint) encode(i uint32) ([]byte, error) {

	if i > 4294967295 {
		r := [5]byte{}
		return r[:], encode(i, r[:])
	} else if i > 16777214 {
		r := [4]byte{}
		return r[:], encode(i, r[:])
	} else if i >= 65535 {
		r := [3]byte{}
		return r[:], encode(i, r[:])
	} else if i > 254 {
		r := [2]byte{}
		return r[:], encode(i, r[:])
	} else {
		r := [1]byte{}
		return r[:], encode(i, r[:])
	}

}

func (v *Varint) decode(b []byte) uint32 {
	var r uint32
	for k, i := range b {
		if i&128 == 0 {
			r = r + uint32(i&255)<<(k*7)
			break
		} else {
			r = r + uint32(i&127)<<(k*7)
		}
	}
	return r
}

func encode(i uint32, r []byte) error {
	n := 0
	for i > 254 {
		r[n] = uint8(i) | 128
		i = i >> 7
		n++
	}

	if i >= 0 {
		r[n] = uint8(i) & 127
	}

	return nil
}

// TestEncode test varint encode
func TestEncode(t *testing.T) {
	v := new(Varint)
	testCase := []uint32{1, 8, 255, 256}
	result := [][]byte{
		{1},
		{8},
		{255, 1},
		{128, 2},
	}

	for k, i := range testCase {
		b, _ := v.encode(i)
		for h, j := range b {
			if result[k][h] != j {
				t.Errorf("encode %d, locate %d want: %b, get %b\n", i, h, result[k][h], j)
			}
		}
	}
}

func TestDecode(t *testing.T) {
	v := new(Varint)
	testCase := []uint32{1, 8, 255, 256}
	result := [][]byte{
		{1},
		{8},
		{255, 1},
		{128, 2},
	}

	for k := range testCase {
		b := v.decode(result[k][:])
		if testCase[k] != b {
			t.Errorf("want: %d, get %d\n", testCase[k], b)
		}
	}
}
```

