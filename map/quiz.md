## map的实现原理
### map头的数据结构
```go
// A header for a Go map.
type hmap struct {
    // 元素个数，调用 len(map) 时，直接返回此值
	count     int
	flags     uint8
	// buckets 的对数 log_2
	B         uint8
	// overflow 的 bucket 近似数
	noverflow uint16
	// 计算 key 的哈希的时候会传入哈希函数
	hash0     uint32
    // 指向 buckets 数组，大小为 2^B
    // 如果元素个数为0，就为 nil
	buckets    unsafe.Pointer
	// 等量扩容的时候，buckets 长度和 oldbuckets 相等
	// 双倍扩容的时候，buckets 长度会是 oldbuckets 的两倍
	oldbuckets unsafe.Pointer
	// 指示扩容进度，小于此地址的 buckets 迁移完成
	nevacuate  uintptr
	extra *mapextra // optional fields
}
```
比较重要的字段：
- count: `len(map)` 直接返回，所以 `len(map)` 的时间/空间复杂度都是`o(1)`
- buckets : 指向桶数组的指针，数组的长度为`2^B`
- B : `buckets` 指向的桶数组长度的对数 `log_2`
- hash0 : 计算 key 的哈希的时候会传入哈希函数，可以理解为`盐值`
- 剩下的字段暂时的学习还没有接触到，还不是很理解，并不是不重要

**从宏观上如下图所示**
![hmap](https://github.com/com-wushuang/goBasic/blob/main/image/hmap.png)

### 桶的数据结构
- 在map头数据结构中，我们讲到了，B 是 buckets 数组的长度的对数，也就是说 buckets 数组的长度就是 2^B
- bucket 里面存储了 key 和 value
**桶的数据结构如下**
```go
type bmap struct {
    topbits  [8]uint8
    keys     [8]keytype
    values   [8]valuetype
    pad      uintptr
    overflow uintptr
}
```
- topbits : 大小为8的数组，每个元素为uint8类型(所以大小是固定的64bit)。你可以认为这是 key-value在桶内的索引
- keys : 存放键
- vales : 存放值
- pad : 内存对齐
- overflow : 因为桶内只能放8个key-value,当放不下的时候就指向下一个桶, 用来扩展桶的长度

**从宏观上如下图所示**
![bmap](https://github.com/com-wushuang/goBasic/blob/main/image/bmap.png)
