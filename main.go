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

func defaultHashFunc(k, l int) int {
	return k & (l - 1)
}

func defaultCollision(i int, compareFunc func(*HashValue) bool, d HashMapData) int {
	if i < 0 || d.Len() <= i {
		return -1
	}
	for index := i; index < d.Len(); index++ {
		if compareFunc(d.Get(index)) {
			return index
		}
	}
	return -1
}

type HashMapData interface {
	Len() int
	Set(int, *HashValue)
	Get(int) *HashValue
	GetV(int) int
	SetV(int, int)
	Reallocate(uint)
}

func MakeHashMapData(d HashMapData) HashMapData {
	return d
}

type ArrayHashMapData struct {
	array []*HashValue
}

func (d *ArrayHashMapData) Len() int {
	return len(d.array)
}

func (d *ArrayHashMapData) Set(i int, v *HashValue) {
	d.array[i] = v
}

func (d *ArrayHashMapData) Get(i int) *HashValue {
	return d.array[i]
}

func (d *ArrayHashMapData) GetV(i int) int {
	return d.array[i].v
}

func (d *ArrayHashMapData) SetV(i int, v int) {
	d.array[i].v = v
}

func (d *ArrayHashMapData) Reallocate(size uint) {
	if uint(len(d.array)) != size {
		d.array = make([]*HashValue, size)
	}
}

type HashMap struct {
	loadFactor float64     // allocator
	useCount   uint        // allocator
	data       HashMapData // data structure
	hashFunc   func(int, int) int
	collision  func(int, func(*HashValue) bool, HashMapData) int
}

func (h *HashMap) Set(k, v int) (int, bool) {
	hashIndex := h.collision(h.hashFunc(k, h.data.Len()), func(h *HashValue) bool {
		return h == nil
	}, h.data)
	if hashIndex == -1 {
		return hashIndex, false
	}

	if h.GetLoadFactor(1) >= h.loadFactor {
		fmt.Printf("HashMap should reallocate\n")
	}

	h.data.Set(hashIndex, &HashValue{
		k: k,
		v: v,
	})
	h.useCount++
	return hashIndex, true
}

func (h *HashMap) Get(k int) (int, bool) {
	hashIndex := h.collision(h.hashFunc(k, h.data.Len()), func(h *HashValue) bool {
		return h != nil && h.equal(k)
	}, h.data)
	if hashIndex == -1 {
		return hashIndex, false
	}
	return h.data.GetV(hashIndex), true
}

func (h *HashMap) Del(k int) (int, bool) {
	hashIndex := h.collision(h.hashFunc(k, h.data.Len()), func(h *HashValue) bool {
		return h != nil && h.equal(k)
	}, h.data)
	if hashIndex == -1 {
		return 0, false
	}
	v := h.data.GetV(hashIndex)
	h.data.SetV(hashIndex, 0)
	h.useCount--
	return v, true
}

func (h *HashMap) GetLoadFactor(delta uint) float64 {
	if h.data.Len() == 0 {
		return 0
	}
	return float64(h.useCount+delta) / float64(h.data.Len())
}

type HashMapOption func(*HashMap)

func MakeHashMap(options ...HashMapOption) *HashMap {
	hashMap := &HashMap{
		loadFactor: DEFAULT_LOAD_FACTOR,
		data: &ArrayHashMapData{
			array: make([]*HashValue, DEFAULT_HASH_MAP_SIZE),
		},
		hashFunc:  defaultHashFunc,
		collision: defaultCollision,
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

func WithHashMapData(data HashMapData) HashMapOption {
	return func(h *HashMap) {
		h.data = data
	}
}

func WithHashMapSize(size uint) HashMapOption {
	return func(h *HashMap) {
		if h.data != nil {
			h.data.Reallocate(size)
		}
	}
}

func WithHashMapHashFunc(f func(int, int) int) HashMapOption {
	return func(h *HashMap) {
		h.hashFunc = f
	}
}

func WithHashMapCollision(f func(int, func(*HashValue) bool, HashMapData) int) HashMapOption {
	return func(h *HashMap) {
		h.collision = f
	}
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
