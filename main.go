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

func defaultHashFunc(k int, l uint) int {
	return k & int((l - 1))
}

type HashMapData interface {
	Len() int
	Get(int, int) (int, bool)
	Set(int, *HashValue) bool
	Del(int, int) (int, bool)
	Range(func(*HashValue) bool)
	Reallocate(uint)
}

// open address collision
type ArrayHashMapData struct {
	array []*HashValue
}

func (d *ArrayHashMapData) Len() int {
	return len(d.array)
}

func (d *ArrayHashMapData) get(hashIndex, key int, op func(int) (int, bool)) (int, bool) {
	for index := hashIndex; index != len(d.array); index++ {
		if d.array[index] != nil && d.array[index].k == key {
			return op(index)
		}
	}
	return 0, false
}

func (d *ArrayHashMapData) Set(hashIndex int, hashValue *HashValue) bool {
	for index := hashIndex; index != len(d.array); index++ {
		if d.array[index] == nil || d.array[index].k == hashValue.k {
			d.array[index] = hashValue
			return true
		}
	}
	return false
}

func (d *ArrayHashMapData) Get(hashIndex, key int) (int, bool) {
	return d.get(hashIndex, key, func(index int) (int, bool) {
		return d.array[index].v, true
	})
}

func (d *ArrayHashMapData) Del(hashIndex, key int) (int, bool) {
	return d.get(hashIndex, key, func(index int) (int, bool) {
		value := d.array[index].v
		d.array[index] = nil
		return value, true
	})
}

func (d *ArrayHashMapData) Range(op func(*HashValue) bool) {
	for index := 0; index != d.Len(); index++ {
		if d.array[index] == nil {
			continue
		}
		if !op(d.array[index]) {
			return
		}
	}
}

func (d *ArrayHashMapData) Reallocate(size uint) {
	if uint(len(d.array)) != size {
		d.array = make([]*HashValue, size)
	}
	// TODO: move data
}

type LinkedListHashMapData struct {
	buckets []*LinkedListNode
}

// chain address collision
type LinkedListNode struct {
	nextNode *LinkedListNode
	preNode  *LinkedListNode
	value    *HashValue
}

func (d *LinkedListHashMapData) Len() int {
	return len(d.buckets)
}

func (d *LinkedListHashMapData) Get(hashIndex, key int) (int, bool) {
	for p := d.buckets[hashIndex]; p != nil; p = p.nextNode {
		if p.value != nil && p.value.k == key {
			return p.value.v, true
		}
	}
	return 0, false
}

func (d *LinkedListHashMapData) Set(hashIndex int, hashValue *HashValue) bool {
	var preNode *LinkedListNode
	for p := d.buckets[hashIndex]; p != nil; p = p.nextNode {
		if p.value.k == hashValue.k {
			p.value = hashValue
			return true
		} else {
			preNode = p
		}
	}
	vNode := &LinkedListNode{
		value: hashValue,
	}
	if preNode == nil {
		d.buckets[hashIndex] = vNode
	} else {
		preNode.nextNode = vNode
		vNode.preNode = preNode
	}
	return true
}

func (d *LinkedListHashMapData) Del(hashIndex, key int) (int, bool) {
	for p := d.buckets[hashIndex]; p != nil; p = p.nextNode {
		if p.value != nil && p.value.k == key {
			value := p.value.v
			p.value = nil
			if p.preNode == nil { // bucket head
				if p.nextNode == nil {
					d.buckets[hashIndex] = nil
				} else {
					d.buckets[hashIndex] = p.nextNode
					d.buckets[hashIndex].preNode = nil
					p.preNode = nil
					p.nextNode = nil
				}
			} else if p.nextNode == nil { // bucket tail
				p.preNode.nextNode = nil
				p.preNode = nil
				p.nextNode = nil
			} else { // bucket middle
				p.preNode.nextNode = p.nextNode
				p.nextNode.preNode = p.preNode
				p.preNode = nil
				p.nextNode = nil
			}
			return value, true
		}
	}
	return 0, false
}

func (d *LinkedListHashMapData) Range(op func(*HashValue) bool) {
	for index := 0; index != d.Len(); index++ {
		for p := d.buckets[index]; p != nil; p = p.nextNode {
			if !op(p.value) {
				return
			}
		}
	}
}

func (d *LinkedListHashMapData) Reallocate(size uint) {
	if uint(len(d.buckets)) != size {
		d.buckets = make([]*LinkedListNode, size)
	}
	// TODO: move data
}

type HashMap struct {
	loadFactor float64     // allocator
	useCount   uint        // allocator
	data       HashMapData // data structure
	hashFunc   func(int, uint) int
}

func (h *HashMap) Set(k, v int) bool {
	hashIndex := h.hashFunc(k, uint(h.data.Len()))
	if hashIndex < 0 || h.data.Len() <= hashIndex {
		return false
	}
	if ok := h.data.Set(hashIndex, &HashValue{
		k: k,
		v: v,
	}); !ok {
		return false
	}
	h.useCount++
	return true
}

func (h *HashMap) Get(k int) (int, bool) {
	hashIndex := h.hashFunc(k, uint(h.data.Len()))
	if hashIndex < 0 || h.data.Len() <= hashIndex {
		return 0, false
	}
	return h.data.Get(hashIndex, k)
}

func (h *HashMap) Del(k int) (int, bool) {
	hashIndex := h.hashFunc(k, uint(h.data.Len()))
	if hashIndex < 0 || h.data.Len() <= hashIndex {
		return 0, false
	}
	v, ok := h.data.Del(hashIndex, k)
	if !ok {
		return 0, false
	}
	h.useCount--
	return v, true
}

func (h *HashMap) GetLoadFactor(delta uint) float64 {
	return float64(h.useCount+delta) / float64(h.data.Len())
}

func (h *HashMap) Range(op func(k, v int) bool) {
	h.data.Range(func(hashValue *HashValue) bool {
		return op(hashValue.k, hashValue.v)
	})
}

type HashMapOption func(*HashMap)

func MakeHashMap(options ...HashMapOption) *HashMap {
	hashMap := &HashMap{
		loadFactor: DEFAULT_LOAD_FACTOR,
		data: &ArrayHashMapData{
			array: make([]*HashValue, DEFAULT_HASH_MAP_SIZE),
		},
		hashFunc: defaultHashFunc,
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

func WithHashMapHashFunc(f func(int, uint) int) HashMapOption {
	return func(h *HashMap) {
		h.hashFunc = f
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
	hashMapTest(keyValueMap, WithHashMapData(&LinkedListHashMapData{
		buckets: make([]*LinkedListNode, DEFAULT_HASH_MAP_SIZE),
	}))
}

func hashMapTest(keyValueMap map[int]int, options ...HashMapOption) {
	testHashMap := MakeHashMap(options...)
	for key, value := range keyValueMap {
		if ok := testHashMap.Set(key, value); !ok {
			fmt.Printf("testHashMap.Set(%v, %v) failed\n", key, value)
		} else {
			fmt.Printf("after Set, testHashMap load factor is %v\n", testHashMap.GetLoadFactor(0))
			// fmt.Printf("testHashMap.Set(%v, %v) at hash index %v success\n", key, value, hashIndex)
		}
	}

	testHashMap.Range(func(k, v int) bool {
		fmt.Printf("key: %v, value: %v\n", k, v)
		return true
	})

	for key, value := range keyValueMap {
		if _value, hasKey := testHashMap.Get(key); !hasKey || _value != value {
			fmt.Printf("testHashMap.Get(%v), not has key or store value %v not equal to origin value %v\n", key, _value, value)
		} else {
			// fmt.Printf("testHashMap.Get(%v), key and store value equal to origin value %v\n", key, value)
		}
	}

	for key, value := range keyValueMap {
		if _value, hasKey := testHashMap.Del(key); !hasKey || _value != value {
			fmt.Printf("testHashMap.Del(%v), not has key or store value %v not equal to origin value %v\n", key, _value, value)
		} else {
			fmt.Printf("after Del, testHashMap load factor is %v\n", testHashMap.GetLoadFactor(0))
			// fmt.Printf("testHashMap.Get(%v), key and store value equal to origin value %v\n", key, value)
		}
	}

	testHashMap.Range(func(k, v int) bool {
		fmt.Printf("key: %v, value: %v\n", k, v)
		return true
	})
}
