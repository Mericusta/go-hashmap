package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	HASH_MAP_SIZE = 1 << 10
)

type HashIntIntValue struct {
	k int
	v int
}

func (h *HashIntIntValue) equal(k int) bool {
	return h.k == k
}

type HashMap struct {
	array     []*HashIntIntValue
	hashFunc  func(int, int) int
	collision func(int, func(*HashIntIntValue) bool, []*HashIntIntValue) int
}

func defaultHashFunc(k, l int) int {
	return k & (l - 1)
}

func defaultCollision(i int, compareFunc func(*HashIntIntValue) bool, a []*HashIntIntValue) int {
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
		array:     make([]*HashIntIntValue, HASH_MAP_SIZE),
		hashFunc:  defaultHashFunc,
		collision: defaultCollision,
	}
	for _, option := range options {
		option(hashMap)
	}
	return hashMap
}

func WithHashMapSize(size uint) HashMapOption {
	return func(h *HashMap) {
		h.array = make([]*HashIntIntValue, size)
	}
}

func WithHashMapHashFunc(f func(int, int) int) HashMapOption {
	return func(h *HashMap) {
		h.hashFunc = f
	}
}

func WithHashMapCollision(f func(int, func(*HashIntIntValue) bool, []*HashIntIntValue) int) HashMapOption {
	return func(h *HashMap) {
		h.collision = f
	}
}

func (h *HashMap) Set(k, v int) (int, bool) {
	hashIndex := h.collision(h.hashFunc(k, len(h.array)), func(h *HashIntIntValue) bool {
		return h == nil
	}, h.array)
	if hashIndex == -1 {
		return hashIndex, false
	}
	h.array[hashIndex] = &HashIntIntValue{
		k: k,
		v: v,
	}
	return hashIndex, true
}

func (h *HashMap) Get(k int) (int, bool) {
	hashIndex := h.collision(h.hashFunc(k, len(h.array)), func(h *HashIntIntValue) bool {
		return h != nil && h.equal(k)
	}, h.array)
	if hashIndex == -1 {
		return hashIndex, false
	}
	return h.array[hashIndex].v, true
}

func main() {
	rand.NewSource(time.Now().UnixNano())
	keyValueMap := make(map[int]int)
	for index := 0; index != HASH_MAP_SIZE>>8; index++ {
		for {
			k := rand.Intn(HASH_MAP_SIZE) + 1
			if _, hasK := keyValueMap[k]; !hasK {
				keyValueMap[k] = index
				fmt.Printf("Key:value = [%v:%v]\n", k, index)
				break
			}
		}
	}

	defaultHashMap := MakeHashMap(WithHashMapSize(HASH_MAP_SIZE >> 9))
	for key, value := range keyValueMap {
		if _, ok := defaultHashMap.Set(key, value); !ok {
			fmt.Printf("defaultHashMap.Set(%v, %v) failed\n", key, value)
		} else {
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
}
