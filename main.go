package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	DEFAULT_HASH_MAP_SIZE = 1 << 10
	DEFAULT_LOAD_FACTOR   = 0.75
)

type HashValue struct {
	k int
	v int
}

func (h *HashValue) equal(k int) bool {
	return h.k == k
}

type HashMap struct {
	loadFactor float64      // allocator
	useCount   uint         // allocator
	array      []*HashValue // data structure
	hashFunc   func(int, int) int
	collision  func(int, func(*HashValue) bool, []*HashValue) int
}

func defaultHashFunc(k, l int) int {
	return k & (l - 1)
}

func defaultCollision(i int, compareFunc func(*HashValue) bool, a []*HashValue) int {
	if i < 0 || len(a) <= i {
		return -1
	}
	for index := i; index < len(a); index++ {
		if compareFunc(a[index]) {
			return index
		}
	}
	return -1
}

type HashMapOption func(*HashMap)

func MakeHashMap(options ...HashMapOption) *HashMap {
	hashMap := &HashMap{
		loadFactor: DEFAULT_LOAD_FACTOR,
		array:      make([]*HashValue, DEFAULT_HASH_MAP_SIZE),
		hashFunc:   defaultHashFunc,
		collision:  defaultCollision,
	}
	for _, option := range options {
		option(hashMap)
	}
	return hashMap
}

func WithHashMapLoadFactor(factor float64) HashMapOption {
	return func(h *HashMap) {
		h.loadFactor = factor
	}
}

func WithHashMapSize(size uint) HashMapOption {
	return func(h *HashMap) {
		h.array = make([]*HashValue, size)
	}
}

func WithHashMapHashFunc(f func(int, int) int) HashMapOption {
	return func(h *HashMap) {
		h.hashFunc = f
	}
}

func WithHashMapCollision(f func(int, func(*HashValue) bool, []*HashValue) int) HashMapOption {
	return func(h *HashMap) {
		h.collision = f
	}
}

func (h *HashMap) Set(k, v int) (int, bool) {
	hashIndex := h.collision(h.hashFunc(k, len(h.array)), func(h *HashValue) bool {
		return h == nil
	}, h.array)
	if hashIndex == -1 {
		return hashIndex, false
	}

	if h.GetLoadFactor(1) >= h.loadFactor {
		fmt.Printf("HashMap should reallocate\n")
	}

	h.array[hashIndex] = &HashValue{
		k: k,
		v: v,
	}
	h.useCount++
	return hashIndex, true
}

func (h *HashMap) Get(k int) (int, bool) {
	hashIndex := h.collision(h.hashFunc(k, len(h.array)), func(h *HashValue) bool {
		return h != nil && h.equal(k)
	}, h.array)
	if hashIndex == -1 {
		return hashIndex, false
	}
	return h.array[hashIndex].v, true
}

func (h *HashMap) Del(k int) (int, bool) {
	hashIndex := h.collision(h.hashFunc(k, len(h.array)), func(h *HashValue) bool {
		return h != nil && h.equal(k)
	}, h.array)
	if hashIndex == -1 {
		return 0, false
	}
	v := h.array[hashIndex].v
	h.array[hashIndex].v = 0
	h.useCount--
	return v, true
}

func (h *HashMap) GetLoadFactor(delta uint) float64 {
	if len(h.array) == 0 {
		return 0
	}
	return float64(h.useCount+delta) / float64(len(h.array))
}

func main() {
	rand.NewSource(time.Now().UnixNano())
	keyValueMap := make(map[int]int)
	for index := 0; index != DEFAULT_HASH_MAP_SIZE>>8; index++ {
		for {
			k := rand.Intn(DEFAULT_HASH_MAP_SIZE) + 1
			if _, hasK := keyValueMap[k]; !hasK {
				keyValueMap[k] = index
				fmt.Printf("Key:value = [%v:%v]\n", k, index)
				break
			}
		}
	}

	hashMapTest(keyValueMap)
	hashMapTest(keyValueMap, WithHashMapSize(DEFAULT_HASH_MAP_SIZE>>9))
}

func hashMapTest(keyValueMap map[int]int, options ...HashMapOption) {
	defaultHashMap := MakeHashMap(options...)
	for key, value := range keyValueMap {
		if _, ok := defaultHashMap.Set(key, value); !ok {
			fmt.Printf("defaultHashMap.Set(%v, %v) failed\n", key, value)
		} else {
			fmt.Printf("after Set, defaultHashMap load factor is %v\n", defaultHashMap.GetLoadFactor(0))
			// fmt.Printf("defaultHashMap.Set(%v, %v) at hash index %v success\n", key, value, hashIndex)
		}
	}

	for key, value := range keyValueMap {
		if _value, hasKey := defaultHashMap.Get(key); !hasKey || _value != value {
			fmt.Printf("defaultHashMap.Get(%v), not has key or store value %v not equal to origin value %v\n", key, _value, value)
		} else {
			// fmt.Printf("defaultHashMap.Get(%v), key and store value equal to origin value %v\n", key, value)
		}
	}

	for key, value := range keyValueMap {
		if _value, hasKey := defaultHashMap.Del(key); !hasKey || _value != value {
			fmt.Printf("defaultHashMap.Del(%v), not has key or store value %v not equal to origin value %v\n", key, _value, value)
		} else {
			fmt.Printf("after Del, defaultHashMap load factor is %v\n", defaultHashMap.GetLoadFactor(0))
			// fmt.Printf("defaultHashMap.Get(%v), key and store value equal to origin value %v\n", key, value)
		}
	}
}
