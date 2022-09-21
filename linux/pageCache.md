# 节点Page Cache占用内存过多问题分析

## 现状分析
### 背景
超融合控制节点 `free -g` 命令，发现节点剩余物理内存几乎为 `0` ，从结果来看：
- `used`: 已被使用的物理内存为 `11g`
- `buff/cache`: 被 `buffer` 和 cache 使用的物理内存为 `18g`

初步得出结论，`buffer` 和 `cache` 使用了大量的内存。

![](https://raw.githubusercontent.com/com-wushuang/pics/main/free-g.png)

`free`命令统计的不够详细，为了查看是 `buffer` 和 `cache` 分别使用多少内存，查看 `/proc/meminfo` 文件：

![](https://raw.githubusercontent.com/com-wushuang/pics/main/proc-meminfo.png)

得出结论是 `cache` 内存占用过多。

### 影响
- `buffer` 和 `cache` 统称为 `page cache` 。是内核管理的内存，它属于内核不属于用户。
- `Page Cache` 存在的意义：减少 `I/O`，提升应用的 `I/O` 速度。
- 因此控制器节点上的 `page cache` 占用内存过多是内核运行的机制导致的，不是用户的行为导致。
- 服务器运行久了后，系统中 `free` 的内存会越来越少，应用在申请内存的时候，即使没有 `free` 内存，只要还有足够可回收的 `Page Cache`，就可以通过回收 `Page Cache` 的方式来申请到内存，回收的方式主要是两种：直接回收和后台回收。

![](https://raw.githubusercontent.com/com-wushuang/pics/main/%E5%BA%94%E7%94%A8%E7%94%B3%E8%AF%B7%E5%86%85%E5%AD%98%E7%9A%84%E8%BF%87%E7%A8%8B.png)

因为直接内存回收是在进程申请内存的过程中同步进行的回收，而这个回收过程可能会消耗很多时间，进而导致进程的后续行为都被迫等待，这样就会造成很长时间的延迟，以及系统的 `CPU` 利用率会升高，最终引起 `load` 飙高。

详细地描述一下过程，在开始内存回收后，首先进行后台异步回收（上图中蓝色标记的地方），这不会引起进程的延迟；如果后台异步回收跟不上进行内存申请的速度，就会开始同步阻塞回收，导致延迟（上图中红色和粉色标记的地方，这就是引起 `load` 高的地方）。

![](https://raw.githubusercontent.com/com-wushuang/pics/main/%E5%86%85%E5%AD%98%E5%9B%9E%E6%94%B6%E8%BF%87%E7%A8%8B.png)

## 解决方案一：主动清理 Page Cache
### 原理
```shell
echo 1 > /proc/sys/vm/drop_caches # 表示仅清除页面缓存（PageCache） 。
echo 2 > /proc/sys/vm/drop_caches # 表示清除回收 slab 分配器中的对象（包括目录项缓存和 inode 缓存）。 slab 分配器是内核中管理内存的一种机制，其中很多缓存数据实现都是用的 pagecache 。
echo 3 > /proc/sys/vm/drop_caches # 表示清除 pagecache 和 slab 分配器中的缓存对象（包括目录项缓存和 inode 缓存）。
```
三个等级1是影响最小的。设置成脚本的形式就是：
```shell
sync; echo 1 > /proc/sys/vm/drop_caches
```
`sync` 是个内存同步命令，删除 `Page Cache`前，先把缓存写入到物理 `disk` 上，防止数据丢失。

### 影响
在生产环境释放 `cache` 要慎重，避免引起性能抖动！原因是 `Page Cache` 设计之初就是为了减少 `I/O`，提升应用的 `I/O` 速度。当我们主动清理了 `Page Cache` 的那个时刻，由于系统中还没有建立新的 `Page Cache`，所有的应用 `I/O` 都会直接操作 `disk` 。从而引起了性能抖动。

## 解决方案二：内核参数优化，及早地触发后台回收来避免应用程序进行直接内存回收
### 原理
![](https://raw.githubusercontent.com/com-wushuang/pics/main/kswapd%E5%9B%9E%E6%94%B6%E7%BA%BF%E7%A8%8B%E5%8E%9F%E7%90%86.png)

后台回收的原理：当内存水位低于 `watermark low` 时，就会唤醒 `kswapd` 进行后台回收，然后 `kswapd` 会一直回收到 `watermark high`。我们可以增大 `/proc/sys/vm/min_free_kbytes` 这个配置选项来及早地触发后台回收。

### 影响
- 该值的设置和总的物理内存并没有一个严格对应的关系，如果配置不当会引起一些副作用，所以在调整该值之前：可以渐进式地增大该值，比如先调整为 `1G`，观察 `sar -B` 中 `pgscand` 是否还有不为 `0` 的情况；如果存在不为 `0` 的情况，继续增加到 `2G`，再次观察是否还有不为 `0` 的情况来决定是否增大，以此类推。
- 提高了内存水位后，应用程序可以直接使用的内存量就会减少，这在一定程度上浪费了内存。所以在调整这一项之前，你需要先思考一下，应用程序更加关注什么，如果关注延迟那就适当地增大该值，如果关注内存的使用量那就适当地调小该值。

