## kubernetes informer机制
- Informer 是 Client-go 中的⼀个核⼼⼯具包。从 Kubernetes 1.7 开始，所有需要监控资源变化情况的调⽤均推荐使⽤ Informer。
- Informer 提供了基于事件通知的只读缓存机制，可以注册资源变化的回调函数，并可以极⼤ 减少 API 的调⽤。

### 设计思路
**关键点**
- 为了让 Client-go 更快地返回 List/Get 请求的结果、减少对 Kubenetes API 的直接调⽤，Informer 被设计实现为⼀个依赖 Kubernetes List/Watch API 、可监听事件并触发回调函数的⼆级缓存⼯具包。

**减少对 Kubenetes API 的直接调⽤**
- 使⽤ Informer 实例的 Lister() ⽅法， List/Get Kubernetes 中的 Object 时，Informer 不会去请求 Kubernetes API，⽽是直接查找缓存在本地内存中的数据(这份数据由 Informer ⾃⼰维护)。通过这种⽅式，Informer 既可 以更快地返回结果，⼜能减少对 Kubernetes API 的直接调⽤。

**依赖 Kubernetes List/Watch API**
- Informer 只会调⽤ Kubernetes List 和 Watch 两种类型的 API。
- Informer 在初始化的时，先调⽤ Kubernetes List API 获得某种 resource 的全部 Object，缓存在内存中; 
- 然后，调⽤ Watch API 去 watch 这种 resource，去 维护这份缓存; 
- 最后，Informer 就不再调⽤ Kubernetes 的任何 API。
- ⽤ List/Watch 去维护缓存、保持⼀致性是⾮常典型的做法，Informer 只在初始化时调⽤⼀ 次 List API，之后完全依赖 Watch API 去维护缓存，没有任何 resync 机制。 

**监听事件并触发回调函数**
- Informer 通过 Kubernetes Watch API 监听某种 resource 下的所有事件。
- ⽽且，Informer 可以添加⾃定义的回调函数，回调函数实例只需实现 `OnAdd(obj interface{})`、 `OnUpdate(oldObj, newObj interface{})`、 和 `OnDelete(obj interface{})` 三个⽅法，这三个⽅法分别对应 informer 监听到创建、更新和删除这三种事件类型。

**⼆级缓存**
- ⼆级缓存属于 Informer 的底层缓存机制，这两级缓存分别是 DeltaFIFO 和 LocalStore。 
- 这两级缓存的⽤途各不相同。DeltaFIFO ⽤来存储 Watch API 返回的各种事件 ，LocalStore 只会被 Lister 的 List/Get ⽅法访问 。 虽然 Informer 和 Kubernetes 之间没有 resync 机制，但 Informer 内部的这两级缓存之间存在 resync 机制。

### 详细解析
**informer主要组件**

- Informer 中主要包含 Controller、Reflector、DeltaFIFO、LocalStore、Lister 和 Processor 六个组件；
- Controller 并不是 Kubernetes Controller，这两个 Controller 并没有任何联系；
- Reflector 的主要作⽤是通过 Kubernetes Watch API 监听某种 resource 下的所有事件；
- DeltaFIFO 和 LocalStore 是 Informer 的两级缓存；
- Lister 主要是被调⽤ List/Get ⽅法；
- Processor 中记录了所有的回调函数实例(即 ResourceEventHandler 实例)，并负责触发这些函数。

**Informer关键逻辑解析**

![informer](https://github.com/com-wushuang/goBasic/blob/main/image/informer.png)

1. Informer 在初始化时，Reflector 会先 List API 获得所有的 Pod
2. Reflect 拿到全部 Pod 后，会将全部 Pod 放到 Store 中 
3. 如果有⼈调⽤ Lister 的 List/Get ⽅法获取 Pod， 那么 Lister 会直接从 Store 中拿数据 
4. Informer 初始化完成之后，Reflector 开始 Watch Pod，监听 Pod 相关 的所有事件;如果此时 pod_1 被 删除，那么 Reflector 会监听到这个事件 
5. Reflector 将 pod_1 被删除 的这个事件发送到 DeltaFIFO 
6. DeltaFIFO ⾸先会将这个事件存储在⾃⼰的数据结构中(实际上是⼀个 queue)，然后会直接操作 Store 中 的数据，删除 Store 中的 pod_1 
7. DeltaFIFO 再 Pop 这个事件到 Controller 中 
8. Controller 收到这个事件，会触发 Processor 的回调函数 
9. LocalStore 会周期性地把所有的 Pod 信息重新放到 DeltaFIFO 中
