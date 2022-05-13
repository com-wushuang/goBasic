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

## key 定位过程
- key 经过哈希计算后得到哈希值，共 64 个 bit 位,计算它到底要落在哪个桶时，只会用到最后 B 个 bit 位
- 还记得前面提到过的 B 吗？如果 B = 5，那么桶的数量，也就是 buckets 数组的长度是 2^5 = 32
- 再用哈希值的高 8 位，找到此 key 在 bucket 中的位置，这是在寻找已有的 key。最开始桶内还没有 key，新加入的 key 会找到第一个空位，放入。

大致的定位过程如上，结合例子看就是如下：
![find_key](https://github.com/com-wushuang/goBasic/blob/main/image/find_key.png)
- 上图中，假定 B = 5，所以 bucket 总数就是 2^5 = 32。
- 首先计算出待查找 key 的哈希，使用低 5 位 00110，找到对应的 6 号 bucket。
- 使用高 8 位 10010111，对应十进制 151，在 6 号 bucket 中寻找 tophash 值（HOB hash）为 151 的 key，找到了 2 号槽位，这样整个查找过程就结束了。
- bucket 中没找到，并且 overflow 不为空，还要继续去 overflow bucket 中寻找，直到找到或是所有的 key 槽位都找遍了，包括所有的 overflow bucket。

## map是如何扩容的
**map的扩容是在如下的条件下发生的**
```go
// src/runtime/hashmap.go/mapassign

// 触发扩容时机
if !h.growing() && (overLoadFactor(int64(h.count), h.B) || tooManyOverflowBuckets(h.noverflow, h.B)) {
		hashGrow(t, h)
	}

// 装载因子超过 6.5
func overLoadFactor(count int64, B uint8) bool {
	return count >= bucketCnt && float32(count) >= loadFactor*float32((uint64(1)<<B))
}

// overflow buckets 太多
func tooManyOverflowBuckets(noverflow uint16, B uint8) bool {
	if B < 16 {
		return noverflow >= uint16(1)<<B // 1*2^B = 2^B
	}
	return noverflow >= 1<<15
}
```
- overLoadFactor : 装载因子超过了6.5
- tooManyOverflowBuckets : overflow buckets 太多

**overLoadFactor**
- 我们知道，每个 bucket 有 8 个空位，在没有溢出，且所有的桶都装满了的情况下，装载因子算出来的结果是 8。
- 因此当装载因子超过 6.5 时，表明很多 bucket 都快要装满了，查找效率和插入效率都变低了。在这个时候进行扩容是有必要的。

**tooManyOverflowBuckets**
- 就是说在装载因子比较小的情况下，这时候 map 的查找和插入效率也很低，而第 1 点识别不出来这种情况。
- 表面现象就是计算装载因子的分子比较小，即 map 里元素总数少，但是 bucket 数量多（真实分配的 bucket 数量多，包括大量的 overflow bucket）。

**扩容触发条件总结**
- 元素太多而桶太少
- 元素太少而桶太多
- 这两种情况都会导致map的效率变低下，所以要进行扩容，提升效率，其实是一种平衡

**情形一**

容量变化：

- 元素太多，而 bucket 数量太少，很简单：将 B 加 1，bucket 最大数量（2^B）直接变成原来 bucket 数量的 2 倍。
- 于是，就有新老 bucket 了。注意，这时候元素都在老 bucket 里，还没迁移到新的 bucket 来。
- 而且，新 bucket 只是最大数量变为原来最大数量（2^B）的 2 倍（2^B * 2）。

搬迁过程：
- 要重新计算 key 的哈希，才能决定它到底落在哪个 bucket。
- 例如，原来 B = 5，计算出 key 的哈希后，只用看它的低 5 位，就能决定它落在哪个 bucket。
- 扩容后，B 变成了 6，因此需要多看一位，它的低 6 位决定 key 落在哪个 bucket,这称为 rehash。
![rehash](https://github.com/com-wushuang/goBasic/blob/main/image/rehash.png)
- 因此，某个 key 在搬迁前后 bucket 序号可能和原来相等，也可能是相比原来加上 2^B（原来的 B 值），取决于 hash 值 第 6 bit 位是 0 还是 1。
- 扩容后，B 增加了 1，意味着 buckets 总数是原来的 2 倍，原来一个桶“裂变”到两个桶。

例子：

原始 B = 2，map中有 2 个 key 的哈希值低 3 位分别为：010，110。由于原来 B = 2，所以低 2 位 10 决定它们落在 2 号桶，现在 B 变成 3，所以 010、110 分别落入 2、6 号桶。
![append_example](https://github.com/com-wushuang/goBasic/blob/main/image/append_example.png)

**情形二**

容量变化：

- 新的 buckets 数量和之前相等。

搬迁过程：

- 从老的 buckets 搬迁到新的 buckets，由于 bucktes 数量不变，因此可以按序号来搬，比如原来在 0 号 bucktes，到新的地方后，仍然放在 0 号 buckets。

**扩容前后map宏观视图**

扩容前，B = 2，共有 4 个 buckets，lowbits 表示 hash 值的低位。假设我们不关注其他 buckets 情况，专注在 2 号 bucket：
![before_map_append](https://github.com/com-wushuang/goBasic/blob/main/image/before_map_append.png)
假设 overflow 太多，触发了等量扩容（对应于前面的情形二),扩容完成后，overflow bucket 消失了，key 都集中到了一个 bucket，更为紧凑了，提高了查找的效率:
![after_map_append_1](https://github.com/com-wushuang/goBasic/blob/main/image/after_map_append_1.png)
假设触发了 2 倍的扩容，那么扩容完成后，老 buckets 中的 key 分裂到了 2 个 新的 bucket。一个在 x part，一个在 y 的 part。依据是 hash 的 lowbits。新 map 中 0-3 称为 x part，4-7 称为 y part。
![after_map_append_2](https://github.com/com-wushuang/goBasic/blob/main/image/after_map_append_2.png)


## key为什么是无序的？
- map 在扩容后，会发生 key 的搬迁，原来落在同一个 bucket 中的 key，搬迁后，有些 key 就要远走高飞了（bucket 序号加上了 2^B）。
- 而遍历的过程，就是按顺序遍历 bucket，同时按顺序遍历 bucket 中的 key。搬迁后，key 的位置发生了重大的变化，有些 key 飞上高枝，有些 key 则原地不动。
- 这样，遍历 map 的结果就不可能按原来的顺序了。
----
- 如果 hard code 的 map。同时，也不向 map 进行插入删除的操作，按理说每次遍历这样的 map 都会返回一个固定顺序的 key/value 序列吧。
- 的确是这样，但是 Go 杜绝了这种做法，因为这样会给新手程序员带来误解，以为这是一定会发生的事情，在某些情况下，可能会酿成大错。
- 当然，Go 做得更绝，当我们在遍历 map 时，并不是固定地从 0 号 bucket 开始遍历，每次都是从一个随机值序号的 bucket 开始遍历，并且是从这个 bucket 的一个随机序号的 cell 开始遍历。这样，即使你是一个写死的 map，仅仅只是遍历它，也不太可能会返回一个固定序列的 key/value 对了。
- “迭代 map 的结果是无序的”这个特性是从 go 1.0 开始加入的。

## 可以边遍历边删除吗？
- map 并不是一个线程安全的数据结构。多个协程同时读写一个 map 是未定义的行为，如果被检测到，会直接 panic。
- 在同一个协程内边遍历边删除，并不会检测到同时读写，理论上是可以这样做的。遍历的结果就可能不会是相同的了，有可能结果遍历结果集中包含了删除的 key，也有可能不包含，这取决于删除 key 的时间
**线程安全的读写map**
- 读之前调用 RLock() 函数，读完之后调用 RUnlock() 函数解锁；写之前调用 Lock() 函数，写完之后，调用 Unlock() 解锁。
- sync.Map 是线程安全的 map，也可以使用。

## map是线程安全的吗？
- 在查找、赋值、遍历、删除的过程中都会检测写标志，一旦发现写标志置位（bit位等于1），则直接 panic。
- 赋值和删除函数在检测完写标志是复位之后，先将写标志位置位，才会进行之后的操作。

**检测写标志**
```go
if h.flags&hashWriting == 0 { // 按位与
		throw("concurrent map writes")
	}
```
**设置写标志**
```go
h.flags |= hashWriting
```

## map的比较
- 两个map变量之间使用判等表达式是没有办法通过编译的
- map变量只能和nil做判等比较
```go
package main

import "fmt"

func main() {
	var m map[string]int
	var n map[string]int

	fmt.Println(m == nil)
	fmt.Println(n == nil)

	// 不能通过编译
	//fmt.Println(m == n)
}
```

**逻辑判等**
1.都为 nil
2.非空、长度相等，指向同一个 map 实体对象
3.相应的 key 指向的 value “深度”相等

## 可以对map的元素取地址么？
无法对map 的 key 或 value 进行取址。以下代码不能通过编译：
```go
package main

import "fmt"

func main() {
	m := make(map[string]int)

	fmt.Println(&m["chengjun"]) // 编译报错：cannot take the address of m["chengjun"]
}
```
如果通过其他 hack 的方式，例如 unsafe.Pointer 等获取到了 key 或 value 的地址，也不能长期持有，因为一旦发生扩容，key 和 value 的位置就会改变，之前保存的地址也就失效了。



