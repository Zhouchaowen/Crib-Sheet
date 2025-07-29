# Map
> 本质是有两个表，一个可读和修改的快照读表，一个可增删改查的脏表；配合锁+原子操作实现数据的写入与更新。miss等于dirty长度后开始脏表提升为快照读表。

```go
// https://github.com/golang/go/blob/master/src/sync/map.go
type Map struct {
    mu Mutex
    read atomic.Pointer[readOnly]
    dirty map[any]*entry
    misses int
}

type readOnly struct {
    m       map[any]*entry
    amended bool
}

type entry struct {
    p atomic.Pointer[any]
}
```
read：原子指针，指向只读的 map
dirty：包含最新写入的数据的 map
mu：互斥锁，保护 dirty map
misses：记录在 read map 中未找到 key 的次数
amended：标记 dirty map 中是否包含 read map 中没有的 key

## 读取流程
```go
// Load 返回 map 中指定 key 对应的值
// 如果值不存在则返回 nil
// ok 参数表示该 key 是否在 map 中找到
func (m *Map) Load(key any) (value any, ok bool) {
    // 原子加载 read map
    read := m.loadReadOnly()
    // 尝试从 read map 中获取值
    e, ok := read.m[key]
    
    // 如果在 read map 中没找到，且 dirty map 中可能有新数据
    if !ok && read.amended {
        // 加锁进入慢路径
        m.mu.Lock()
        
        // 避免虚假未命中：
        // 如果在等待锁的过程中 dirty map 被提升为 read map
        // 则需要重新检查 read map
        // (如果后续对同一个 key 的加载不会未命中，就不值得为这个 key 复制 dirty map)
        read = m.loadReadOnly()
        e, ok = read.m[key]
        
        // 再次检查：如果仍然没找到且 dirty map 确实有新数据
        if !ok && read.amended {
            // 从 dirty map 中查找
            e, ok = m.dirty[key]
            
            // 无论该条目是否存在，都记录一次未命中：
            // 这个 key 在 dirty map 被提升为 read map 之前
            // 都会走慢路径
            m.missLocked()
        }
        m.mu.Unlock()
    }
    
    // 如果最终没找到，返回 nil 和 false
    if !ok {
        return nil, false
    }
    
    // 找到了则返回值
    return e.load()
}
```
先无锁读取 read（只读快照），如果命中，直接返回。
若未命中且 read.amended 为 true（说明 dirty 有新 key），则加锁后：
再次检查 read（期间可能被 promote），
若还是未命中，则去 dirty 查找，
查找后计数 misses，misses 达到阈值会 promote dirty 为新的 read。

## 写入流程
```go
// Store 为指定的 key 设置 value
func (m *Map) Store(key, value any) {
    _, _ = m.Swap(key, value)
}

// Swap 交换指定 key 的值，并返回之前的值（如果存在的话）
// loaded 参数表示 key 是否存在于 map 中
func (m *Map) Swap(key, value any) (previous any, loaded bool) {
    // 1. 快速路径：尝试直接在 read map 中更新
    read := m.loadReadOnly()
    if e, ok := read.m[key]; ok {
        // 如果 key 存在于 read map 中，尝试原子交换
        if v, ok := e.trySwap(&value); ok {
            if v == nil {
                return nil, false
            }
            return *v, true
        }
    }

    // 2. 慢路径：需要加锁处理
    m.mu.Lock()
    read = m.loadReadOnly()
    if e, ok := read.m[key]; ok {
        // 2.1 key 在 read map 中存在
        if e.unexpungeLocked() {
            // 如果该条目之前被标记为删除
            // 这意味着存在非空的 dirty map，且该条目不在其中
            // 需要将条目添加到 dirty map 中
            m.dirty[key] = e
        }
        // 执行加锁的值交换
        if v := e.swapLocked(&value); v != nil {
            loaded = true
            previous = *v
        }
    } else if e, ok := m.dirty[key]; ok {
        // 2.2 key 在 dirty map 中存在
        if v := e.swapLocked(&value); v != nil {
            loaded = true
            previous = *v
        }
    } else {
        // 2.3 key 在两个 map 中都不存在
        if !read.amended {
            // 这是第一次向 dirty map 添加新的 key
            // 确保 dirty map 已分配，并标记 read map 为不完整
            m.dirtyLocked()
            m.read.Store(&readOnly{m: read.m, amended: true})
        }
        // 在 dirty map 中创建新条目
        m.dirty[key] = newEntry(value)
    }
    m.mu.Unlock()
    return previous, loaded
}
```
写操作必须加锁。
若 key 在 read 中，需 unexpunge（从已删除/expunged 恢复），并加入 dirty。
若 key 只在 dirty，直接更新。
新 key 则先确保 dirty 存在（必要时从 read 复制），写入 dirty，同时设置 read.amended = true。

## 脏表提升
脏表提升的实现就在 missLocked 方法中，提升时机是 miss 次数达到 dirty 的长度。
```go
func (m *Map) missLocked() {
	m.misses++
	if m.misses < len(m.dirty) {
		return
	}
	m.read.Store(&readOnly{m: m.dirty})
	m.dirty = nil
	m.misses = 0
}
```
sync.Map 通过分层存储（read/dirty）、无锁读、脏表提升及智能 miss 统计等机制，实现了高效的并发读写。其关键在于读写分离和热点数据的快速迁移，最大限度减少锁竞争，提升多协程场景下 map 的性能。