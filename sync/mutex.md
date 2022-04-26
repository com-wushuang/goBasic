## 数据结构
`mutex` 的源码主要是在 `src/sync/mutex.go`文件里，它的结构体比较简单，如下:
```go
type Mutex struct {
	state int32
	sema  uint32
}
```
## mutex状态位 
其实 `mutex` 本质上就是一个关于信号量的阻塞唤起操作。
- 当 `goroutine` 不能占有锁资源的时候会被阻塞挂起，此时不能继续执行后面的代码逻辑。
- 当 `mutex` 释放锁资源时，则会继续唤起之前的 `goroutine` 去抢占锁资源。
- `mutex` 的 `state` 有 32 位，它的低 3 位分别表示 3 种状态：`唤醒状态`、`上锁状态`、`饥饿状态`，剩下的位数则表示当前阻塞等待的 `goroutine` 数量。 `mutex` 会根据当前的 `state` 状态来进入正常模式、饥饿模式或者是自旋。

## mutex正常模式
- 当 `mutex` 调用 `Unlock()` 方法释放锁资源时，如果发现有等待唤起的 `Goroutine` 队列时，则会将队头的 `Goroutine` 唤起
- 队头的 `goroutine` 被唤起后，会调用 `CAS` 方法去尝试性的修改 `state` 状态，如果修改成功，则表示占有锁资源成功


## mutex饥饿模式
- 由于上面的 `Goroutine` 唤起后并不是直接的占用资源，还需要调用 `CAS` 方法去尝试性占有锁资源
- 如果此时有新来的 `Goroutine`，那么它也会调用 `CAS` 方法去尝试性的占有资源
- 但对于 `Go` 的调度机制来讲，会比较偏向于 `CPU` 占有时间较短的 `Goroutine` 先运行，而这将造成一定的几率让新来的 `Goroutine` 一直获取到锁资源，此时队头的 `Goroutine` 将一直占用不到，导致饿死
- 针对这种情况，`Go` 采用了饥饿模式。即通过判断队头 `Goroutine` 在超过一定时间后还是得不到资源时，会在 `Unlock` 释放锁资源时，直接将锁资源交给队头 `Goroutine`，并且将当前状态改为饥饿模式
- 后面如果有新来的 `Goroutine` 发现是饥饿模式时， 则会直接添加到等待队列的队尾
- 所以其实在饥饿模式下，`Goroutine` 的调度方式是`FIFO`

## 自旋
同一时刻只能有一个`Goroutine`获取到锁，没有获取到锁的`Goroutine`通常有两种处理方式：
- 一直循环等待判断该资源是否已经释放锁，这种锁叫做自旋锁，它不用将线程阻塞起来(NON-BLOCKING)；
- 把自己阻塞起来，等待重新调度请求，这种是互斥锁。 
自旋锁的原理比较简单，如果持有锁的线程能在短时间内释放锁资源，那么那些等待竞争锁的线程就不需要做内核态和用户态之间的切换进入阻塞状态，它们只需要等一等(自旋)，等到持有锁的线程释放锁之后即可获取，这样就避免了用户进程和内核切换的消耗。

如果 Goroutine 占用锁资源的时间比较短，那么每次都调用信号量来阻塞唤起 goroutine，将会很浪费资源。 因此在符合一定条件后，mutex 会让当前的 Goroutine 去空转 CPU，在空转完后再次调用 CAS 方法去尝试性的占有锁资源，直到不满足自旋条件，则最终会加入到等待队列里。

但是如果长时间上锁的话，自旋锁会非常耗费性能，它阻止了其他线程的运行和调度。线程持有锁的时间越长，则持有该锁的线程将被OS调度程序中断的风险越大。如果发生中断情况，那么其他线程将保持旋转状态(反复尝试获取锁)，而持有该锁的线程并不打算释放锁，这样导致的是结果是无限期推迟，直到持有锁的线程可以完成并释放它为止。

解决上面这种情况一个很好的方式是给自旋锁设定一个自旋时间，等时间一到立即释放自旋锁。自旋锁的目的是占着CPU资源不进行释放，等到获取锁立即进行处理。

## 自旋的条件
- 还没自旋超过 4 次
- 多核处理器
- `GOMAXPROCS` > 1
- `p` 上本地 `Goroutine` 队列为空

## mutex的Lock()过程
- 首先，如果 mutex 的 state = 0，即没有谁在占有资源，也没有阻塞等待唤起的 goroutine。则会调用 CAS 方法去尝试性占有锁，不做其他动作
- 如果不符合 m.state = 0，则进一步判断是否需要自旋
- 当不需要自旋又或者自旋后还是得不到资源时，此时会调用 `runtime_SemacquireMutex` 信号量函数，将当前的` goroutine` 阻塞并加入等待唤起队列里
- 当有锁资源释放，mutex 在唤起了队头的 goroutine 后，队头 goroutine 会尝试性的占有锁资源，而此时也有可能会和新到来的 goroutine 一起竞争
- 当队头 goroutine 一直得不到资源时，则会进入饥饿模式，直接将锁资源交给队头 goroutine，让新来的 goroutine 阻塞并加入到等待队列的队尾里
- 对于饥饿模式将会持续到没有阻塞等待唤起的 goroutine 队列时，才会解除

```go
// Lock mutex 的锁方法。
func (m *Mutex) Lock() {
	// 快速上锁.
	if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
		if race.Enabled {
			race.Acquire(unsafe.Pointer(m))
		}
		return
	}
	// 快速上锁失败，将进行操作较多的上锁动作。
	m.lockSlow()
}

func (m *Mutex) lockSlow() {
  var waitStartTime int64  // 记录当前 goroutine 的等待时间
  starving := false // 是否饥饿
  awoke := false // 是否被唤醒
  iter := 0 // 自旋次数
  old := m.state // 当前 mutex 的状态
  for {
    // 当前 mutex 的状态已上锁，并且非饥饿模式，并且符合自旋条件
    if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) {
      // 当前还没设置过唤醒标识
      if !awoke && old&mutexWoken == 0 && old>>mutexWaiterShift != 0 &&
        atomic.CompareAndSwapInt32(&m.state, old, old|mutexWoken) {
        awoke = true
      }
      runtime_doSpin()
      iter++
      old = m.state
      continue
    }
    new := old
    // 如果不是饥饿状态，则尝试上锁
    // 如果是饥饿状态，则不会上锁，因为当前的 goroutine 将会被阻塞并添加到等待唤起队列的队尾
    if old&mutexStarving == 0 {
      new |= mutexLocked
    }
    // 等待队列数量 + 1
    if old&(mutexLocked|mutexStarving) != 0 {
      new += 1 << mutexWaiterShift
    }
    // 如果 goroutine 之前是饥饿模式，则此次也设置为饥饿模式
    if starving && old&mutexLocked != 0 {
      new |= mutexStarving
    }
    //
    if awoke {
      // 如果状态不符合预期，则报错
      if new&mutexWoken == 0 {
        throw("sync: inconsistent mutex state")
      }
      // 新状态值需要清除唤醒标识，因为当前 goroutine 将会上锁或者再次 sleep
      new &^= mutexWoken
    }
    // CAS 尝试性修改状态，修改成功则表示获取到锁资源
    if atomic.CompareAndSwapInt32(&m.state, old, new) {
      // 非饥饿模式，并且未获取过锁，则说明此次的获取锁是 ok 的，直接 return
      if old&(mutexLocked|mutexStarving) == 0 {
        break
      }
      // 根据等待时间计算 queueLifo
      queueLifo := waitStartTime != 0
      if waitStartTime == 0 {
        waitStartTime = runtime_nanotime()
      }
      // 到这里，表示未能上锁成功
      // queueLife = true, 将会把 goroutine 放到等待队列队头
      // queueLife = false, 将会把 goroutine 放到等待队列队尾
      runtime_SemacquireMutex(&m.sema, queueLifo, 1)
      // 计算是否符合饥饿模式，即等待时间是否超过一定的时间
      starving = starving || runtime_nanotime()-waitStartTime > starvationThresholdNs
      old = m.state
      // 上一次是饥饿模式
      if old&mutexStarving != 0 {
        if old&(mutexLocked|mutexWoken) != 0 || old>>mutexWaiterShift == 0 {
          throw("sync: inconsistent mutex state")
        }
        delta := int32(mutexLocked - 1<<mutexWaiterShift)
        // 此次不是饥饿模式又或者下次没有要唤起等待队列的 goroutine 了
        if !starving || old>>mutexWaiterShift == 1 {
          delta -= mutexStarving
        }
        atomic.AddInt32(&m.state, delta)
        break
      }
      // 此处已不再是饥饿模式了，清除自旋次数，重新到 for 循环竞争锁。
      awoke = true
      iter = 0
    } else {
      old = m.state
    }
  }

  if race.Enabled {
    race.Acquire(unsafe.Pointer(m))
  }
}
```

## Unlock 过程
- mutex 的 Unlock() 则相对简单。同样的，会先进行快速的解锁，即没有等待唤起的 goroutine，则不需要继续做其他动作
- 如果当前是正常模式，则简单的唤起队头 Goroutine
- 如果是饥饿模式，则会直接将锁交给队头 Goroutine，然后唤起队头 Goroutine，让它继续运行

```go
// Unlock 对 mutex 解锁.
// 如果没有上过锁，缺调用此方法解锁，将会抛出运行时错误。
// 它将允许在不同的 Goroutine 上进行上锁解锁
func (m *Mutex) Unlock() {
	if race.Enabled {
		_ = m.state
		race.Release(unsafe.Pointer(m))
	}

	// 快速尝试解锁
	new := atomic.AddInt32(&m.state, -mutexLocked)
	if new != 0 {
		// 快速解锁失败，将进行操作较多的解锁动作。
		m.unlockSlow(new)
	}
}

func (m *Mutex) unlockSlow(new int32) {
  // 非上锁状态，直接抛出异常
  if (new+mutexLocked)&mutexLocked == 0 {
    throw("sync: unlock of unlocked mutex")
  }
  // 正常模式
  if new&mutexStarving == 0 {
    old := new
    for {
      // 没有需要唤起的等待队列
      if old>>mutexWaiterShift == 0 || old&(mutexLocked|mutexWoken|mutexStarving) != 0 {
        return
      }
      // 唤起等待队列并数量-1
      new = (old - 1<<mutexWaiterShift) | mutexWoken
      if atomic.CompareAndSwapInt32(&m.state, old, new) {
        runtime_Semrelease(&m.sema, false, 1)
        return
      }
      old = m.state
    }
  } else {
    //饥饿模式，将锁直接给等待队列的队头 goroutine
    runtime_Semrelease(&m.sema, true, 1)
  }
}
```

