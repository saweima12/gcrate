package shardmap

import (
	"fmt"
	"sync"
)

const SHARD_DEFAULT = 32

type Stringer interface {
	fmt.Stringer
	comparable
}

type KVShardBlock[K comparable, V any] struct {
	items map[K]V
	mu    sync.Mutex
}

func createMap[K comparable, V any](f ShardingFunc[K], opts ...Option[K, V]) *ShardMap[K, V] {
	resp := &ShardMap[K, V]{
		shardNum:     SHARD_DEFAULT,
		shardingFunc: f,
	}

	for _, opt := range opts {
		opt(resp)
	}

	// Initialize shards
	resp.shards = make([]*KVShardBlock[K, V], resp.shardNum)
	for i := range resp.shards {
		resp.shards[i] = &KVShardBlock[K, V]{items: make(map[K]V)}
	}

	return resp
}

// Create a ShardMap with string as the key.
func New[V any](opts ...Option[string, V]) *ShardMap[string, V] {
	return createMap[string, V](fnv32, opts...)
}

// Create a ShardMap where the key can be any struct implementing the String() method.
func NewStringer[K Stringer, V any](opts ...Option[K, V]) *ShardMap[K, V] {
	return createMap[K, V](strFnv32, opts...)
}

type ShardingFunc[K comparable] func(key K) uint32
type ShardMap[K comparable, V any] struct {
	shardNum     uint8
	shardingFunc ShardingFunc[K]
	shards       []*KVShardBlock[K, V]

	length int
}

func (sm *ShardMap[K, V]) Length() int {
	total := 0
	for i := range sm.shards {
		total += len(sm.shards[i].items)
	}
	return total
}

// Load returns the value stored the map for a key or nil, if no value is present.
func (sm *ShardMap[K, V]) Get(key K) (value V, ok bool) {
	index := sm.shardingFunc(key) % uint32(sm.shardNum)
	shard := sm.shards[index]

	shard.mu.Lock()
	defer shard.mu.Unlock()
	val, ok := shard.items[key]
	return val, ok
}

// Set the value for a key.
func (sm *ShardMap[K, V]) Set(key K, value V) {
	index := sm.shardingFunc(key) % uint32(sm.shardNum)
	shard := sm.shards[index]

	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.items[key] = value
	sm.length++
}

func (sm *ShardMap[K, V]) Remove(key K) {
	index := sm.shardingFunc(key) % uint32(sm.shardNum)
	shard := sm.shards[index]

	shard.mu.Lock()
	defer shard.mu.Unlock()
	delete(shard.items, key)
}

// recommend number.
const FNV_BASIS = uint32(2166136261)

// FNV-1a algorithm
func fnv32(key string) uint32 {
	const FNV_PRIME = uint32(16777619)
	nhash := FNV_BASIS
	for i := 0; i < len(key); i++ {
		nhash ^= uint32(key[i])
		nhash *= FNV_PRIME
	}
	return nhash
}

// Support someone who implement fmt.Stringer
func strFnv32[K fmt.Stringer](key K) uint32 {
	return fnv32(key.String())
}
