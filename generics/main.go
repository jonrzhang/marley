// 场景：泛型内存缓存
// 应用中多处需要缓存（用户信息、商品数据、配置项），类型各不相同
// 泛型让一套缓存逻辑适用于所有类型，避免重复代码和 interface{} 类型断言
package main

import (
	"fmt"
	"sync"
	"time"
)

// Cache 是类型安全的带 TTL 过期缓存
type Cache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]cacheItem[V]
	ttl   time.Duration
}

type cacheItem[V any] struct {
	value     V
	expiresAt time.Time
}

func NewCache[K comparable, V any](ttl time.Duration) *Cache[K, V] {
	c := &Cache[K, V]{
		items: make(map[K]cacheItem[V]),
		ttl:   ttl,
	}
	// 后台清理过期 key
	go c.cleanup()
	return c
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cacheItem[V]{value: value, expiresAt: time.Now().Add(c.ttl)}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.items[key]
	if !ok || time.Now().After(item.expiresAt) {
		var zero V
		return zero, false
	}
	return item.value, true
}

func (c *Cache[K, V]) cleanup() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, v := range c.items {
			if now.After(v.expiresAt) {
				delete(c.items, k)
			}
		}
		c.mu.Unlock()
	}
}

// --- 业务类型 ---

type User struct {
	ID   int
	Name string
}

type Product struct {
	ID    int
	Name  string
	Price float64
}

func main() {
	// 用户缓存：key=int，value=User
	userCache := NewCache[int, User](500 * time.Millisecond)
	userCache.Set(1, User{ID: 1, Name: "Alice"})
	userCache.Set(2, User{ID: 2, Name: "Bob"})

	// 商品缓存：key=string，value=Product
	productCache := NewCache[string, Product](500 * time.Millisecond)
	productCache.Set("sku-001", Product{ID: 101, Name: "Laptop", Price: 999.99})

	// 配置缓存：key=string，value=string
	configCache := NewCache[string, string](1 * time.Second)
	configCache.Set("theme", "dark")

	// 查询
	if u, ok := userCache.Get(1); ok {
		fmt.Printf("User  cache hit: %+v\n", u)
	}
	if p, ok := productCache.Get("sku-001"); ok {
		fmt.Printf("Product cache hit: %+v\n", p)
	}
	if v, ok := configCache.Get("theme"); ok {
		fmt.Printf("Config cache hit: theme=%s\n", v)
	}

	// miss
	if _, ok := userCache.Get(99); !ok {
		fmt.Println("User  cache miss: id=99")
	}

	// 等待过期
	fmt.Println("\nWaiting for user cache to expire...")
	time.Sleep(600 * time.Millisecond)
	if _, ok := userCache.Get(1); !ok {
		fmt.Println("User  cache expired: id=1")
	}
	// config cache still alive (1s TTL)
	if v, ok := configCache.Get("theme"); ok {
		fmt.Printf("Config still alive: theme=%s\n", v)
	}
}
