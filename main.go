package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
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

// ----------------------------------------------------------------

// open address collision

// liner detection and hashing LDH

// liner detection and hashing, awful but works...

type ldhHashMapData struct {
	array []*HashValue
}

func (d *ldhHashMapData) Len() int {
	return len(d.array)
}

func (d *ldhHashMapData) get(hashIndex, key int, op func(int) (int, bool)) (int, bool) {
	for index := hashIndex; index < len(d.array); index++ {
		if d.array[index] != nil && d.array[index].k == key {
			return op(index)
		}
	}
	return 0, false
}

func (d *ldhHashMapData) Set(hashIndex int, hashValue *HashValue) bool {
	for index := hashIndex; index != len(d.array); index++ {
		if d.array[index] == nil || d.array[index].k == hashValue.k {
			d.array[index] = hashValue
			return true
		}
	}
	return false
}

func (d *ldhHashMapData) Get(hashIndex, key int) (int, bool) {
	return d.get(hashIndex, key, func(index int) (int, bool) {
		return d.array[index].v, true
	})
}

func (d *ldhHashMapData) Del(hashIndex, key int) (int, bool) {
	return d.get(hashIndex, key, func(index int) (int, bool) {
		value := d.array[index].v
		d.array[index] = nil
		return value, true
	})
}

func (d *ldhHashMapData) Range(op func(*HashValue) bool) {
	for _, hashValue := range d.array {
		if hashValue == nil {
			continue
		}
		if !op(hashValue) {
			return
		}
	}
}

func (d *ldhHashMapData) Reallocate(size uint) {
	if uint(len(d.array)) != size {
		d.array = make([]*HashValue, size)
	}
	// TODO: move data
}

// second detection and hashing SDH

// second detection and hashing is nearly shit...

type sdhHashMapData struct {
	array []*HashValue
}

func (d *sdhHashMapData) Len() int {
	return len(d.array)
}

func (d *sdhHashMapData) get(hashIndex, key int, op func(int) (int, bool)) (int, bool) {
	for index := 1; index <= len(d.array)/2; index++ {
		lIndex := hashIndex - index*index
		rIndex := hashIndex + index*index
		if lIndex >= 0 && d.array[lIndex] != nil && d.array[lIndex].k == key {
			return op(lIndex)
		} else if rIndex < len(d.array) && d.array[rIndex] != nil && d.array[rIndex].k == key {
			return op(rIndex)
		}
	}
	return 0, false
}

func (d *sdhHashMapData) Set(hashIndex int, hashValue *HashValue) bool {
	for index := 1; index <= len(d.array)/2; index++ {
		lIndex := hashIndex - index*index
		rIndex := hashIndex + index*index
		if lIndex >= 0 && (d.array[lIndex] == nil || d.array[lIndex].k == hashValue.k) {
			d.array[lIndex] = hashValue
			return true
		} else if rIndex < len(d.array) && (d.array[rIndex] == nil || d.array[rIndex].k == hashValue.k) {
			d.array[rIndex] = hashValue
			return true
		}
	}
	return false
}

func (d *sdhHashMapData) Get(hashIndex, key int) (int, bool) {
	return d.get(hashIndex, key, func(index int) (int, bool) {
		return d.array[index].v, true
	})
}

func (d *sdhHashMapData) Del(hashIndex, key int) (int, bool) {
	return d.get(hashIndex, key, func(index int) (int, bool) {
		value := d.array[index].v
		d.array[index] = nil
		return value, true
	})
}

func (d *sdhHashMapData) Range(op func(*HashValue) bool) {
	for _, hashValue := range d.array {
		if hashValue == nil {
			continue
		}
		if !op(hashValue) {
			return
		}
	}
}

func (d *sdhHashMapData) Reallocate(size uint) {
	if uint(len(d.array)) != size {
		d.array = make([]*HashValue, size)
	}
	// TODO: move data
}

// random detection and hashing

// random detection and hashing is a shit...

// chain address collision

// doubly linked list DLL

type dllNode struct {
	nextNode *dllNode
	preNode  *dllNode
	value    *HashValue
}

type dllHashMapData struct {
	buckets []*dllNode
}

func (d *dllHashMapData) Len() int {
	return len(d.buckets)
}

func (d *dllHashMapData) Get(hashIndex, key int) (int, bool) {
	for p := d.buckets[hashIndex]; p != nil; p = p.nextNode {
		if p.value != nil && p.value.k == key {
			return p.value.v, true
		}
	}
	return 0, false
}

func (d *dllHashMapData) Set(hashIndex int, hashValue *HashValue) bool {
	var preNode *dllNode
	for p := d.buckets[hashIndex]; p != nil; p = p.nextNode {
		if p.value.k == hashValue.k {
			p.value = hashValue
			return true
		} else {
			preNode = p
		}
	}
	vNode := &dllNode{
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

func (d *dllHashMapData) Del(hashIndex, key int) (int, bool) {
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

func (d *dllHashMapData) Range(op func(*HashValue) bool) {
	for _, bucket := range d.buckets {
		for node := bucket; node != nil; node = node.nextNode {
			if !op(node.value) {
				return
			}
		}
	}
}

func (d *dllHashMapData) Reallocate(size uint) {
	if uint(len(d.buckets)) != size {
		d.buckets = make([]*dllNode, size)
	}
	// TODO: move data
}

// binary search tree BST

type bstNode struct {
	leftNode  *bstNode
	rightNode *bstNode
	value     *HashValue
}

func (n *bstNode) preOrderTraversal(op func(*HashValue) bool, deep int) bool {
	fmt.Printf("%v", strings.Repeat("\t", deep))
	if !op(n.value) {
		return false
	}
	if n.leftNode != nil {
		n.leftNode.preOrderTraversal(op, deep+1)
	}
	if n.rightNode != nil {
		n.rightNode.preOrderTraversal(op, deep+1)
	}
	return true
}

func (n *bstNode) inOrderTraversal(op func(*HashValue) bool) bool {
	if n.leftNode != nil {
		n.leftNode.inOrderTraversal(op)
	}
	if !op(n.value) {
		return false
	}
	if n.rightNode != nil {
		n.rightNode.inOrderTraversal(op)
	}
	return true
}

type bstHashMapData struct {
	buckets []*bstNode
}

func (d *bstHashMapData) Len() int {
	return len(d.buckets)
}

func (d *bstHashMapData) Get(hashIndex, key int) (int, bool) {
	if d.buckets[hashIndex] == nil {
		return 0, false
	} else {
		node := d.buckets[hashIndex]
		for {
			if key < node.value.k {
				if node.leftNode == nil {
					return 0, false
				} else {
					node = node.leftNode
				}
			} else if node.value.k < key {
				if node.rightNode == nil {
					return 0, false
				} else {
					node = node.rightNode
				}
			} else {
				return node.value.v, true
			}
		}
	}
}

func (d *bstHashMapData) Set(hashIndex int, hashValue *HashValue) bool {
	vNode := &bstNode{
		value: hashValue,
	}
	if d.buckets[hashIndex] == nil {
		d.buckets[hashIndex] = vNode
	} else {
		node := d.buckets[hashIndex]
		for {
			if hashValue.k < node.value.k {
				if node.leftNode == nil {
					node.leftNode = vNode
					return true
				} else {
					node = node.leftNode
				}
			} else if node.value.k < hashValue.k {
				if node.rightNode == nil {
					node.rightNode = vNode
					return true
				} else {
					node = node.rightNode
				}
			} else {
				node.value = hashValue
				return true
			}
		}
	}
	return true
}

// 移动左子树到右子树最小节点的左子树下（树易失衡）
func (d *bstHashMapData) del(hashIndex, key int) (int, bool) {
	if d.buckets[hashIndex] == nil {
		return 0, true
	} else {
		var parentNode *bstNode
		node := d.buckets[hashIndex]
		for {
			if key < node.value.k {
				if node.leftNode == nil {
					return 0, true
				} else {
					parentNode = node
					node = node.leftNode
				}
			} else if node.value.k < key {
				if node.rightNode == nil {
					return 0, true
				} else {
					parentNode = node
					node = node.rightNode
				}
			} else {
				value := node.value.v
				if parentNode == nil {
					d.buckets[hashIndex] = nil
				} else {
					if parentNode.leftNode == node {
						if node.rightNode == nil {
							parentNode.leftNode = node.leftNode
						} else {
							parentNode.leftNode = node.rightNode
							if node.leftNode != nil {
								leftSubNode := node.leftNode
								for node = node.rightNode; node.leftNode != nil; node = node.leftNode {
								}
								node.leftNode = leftSubNode
							}
						}
					} else if parentNode.rightNode == node {
						if node.rightNode == nil {
							parentNode.rightNode = node.leftNode
						} else {
							parentNode.rightNode = node.rightNode
							if node.leftNode != nil {
								leftSubNode := node.leftNode
								for node = node.rightNode; node.leftNode != nil; node = node.leftNode {
								}
								node.leftNode = leftSubNode
							}
						}
					} else {
						// TODO: BST error, need range and print tree
						return 0, false
					}
				}
				return value, true
			}
		}
	}
}

// 删除匹配节点，移动右子树最小节点到匹配节点
func (d *bstHashMapData) Del(hashIndex, key int) (int, bool) {
	if d.buckets[hashIndex] == nil {
		return 0, true
	} else {
		var parentNode *bstNode
		node := d.buckets[hashIndex]
		// fmt.Println()
		// fmt.Printf("Before Delete %v preOrder\n", key)
		// d.buckets[hashIndex].preOrderTraversal(func(h *HashValue) bool {
		// 	fmt.Printf("DEBUG: range key: %v, value: %v\n", h.k, h.v)
		// 	return true
		// }, 0)
		for {
			if key < node.value.k {
				if node.leftNode == nil {
					return 0, true
				} else {
					parentNode = node
					node = node.leftNode
				}
			} else if node.value.k < key {
				if node.rightNode == nil {
					return 0, true
				} else {
					parentNode = node
					node = node.rightNode
				}
			} else {
				value, deleteNode := node.value.v, node

				var newNode *bstNode
				leftNode := node.leftNode
				rightNode := node.rightNode
				minRightNodeParentNode := node
				for node = node.rightNode; node != nil && node.leftNode != nil; minRightNodeParentNode, node = node, node.leftNode {
				}
				if node == nil { // 单左链表
					newNode = leftNode
				} else if minRightNodeParentNode == deleteNode { // 单右链表
					newNode = deleteNode.rightNode
					newNode.leftNode = leftNode
				} else if minRightNodeParentNode != node {
					minRightNodeParentNode.leftNode = node.rightNode
					node.leftNode = leftNode
					node.rightNode = rightNode
					newNode = node
				} else {
					newNode = node.rightNode
				}

				if parentNode == nil {
					d.buckets[hashIndex] = newNode
				} else if parentNode.leftNode == deleteNode {
					parentNode.leftNode = newNode
				} else if parentNode.rightNode == deleteNode {
					parentNode.rightNode = newNode
				} else {
					// TODO: error
					return 0, false
				}

				deleteNode.leftNode = nil
				deleteNode.rightNode = nil

				// if d.buckets[hashIndex] != nil {
				// 	fmt.Printf("After Delete %v preOrder\n", key)
				// 	d.buckets[hashIndex].preOrderTraversal(func(h *HashValue) bool {
				// 		fmt.Printf("DEBUG: range key: %v, value: %v\n", h.k, h.v)
				// 		return true
				// 	}, 0)
				// }
				return value, true
			}
		}
		// fmt.Printf("DEBUG: search but not find!")
		// d.buckets[hashIndex].inOrderTraversal(func(h *HashValue) bool {
		// 	fmt.Printf("DEBUG: range key: %v, value: %v\n", h.k, h.v)
		// 	return true
		// })
	}
}

func (d *bstHashMapData) Range(op func(*HashValue) bool) {
	for _, bucket := range d.buckets {
		if bucket != nil {
			fmt.Printf("inOrderTraversal\n")
			if !bucket.inOrderTraversal(op) {
				return
			}
			fmt.Printf("preOrderTraversal\n")
			if !bucket.preOrderTraversal(op, 0) {
				return
			}
		}
	}
}

func (d *bstHashMapData) Reallocate(size uint) {
	if uint(len(d.buckets)) != size {
		d.buckets = make([]*bstNode, size)
	}
	// TODO: move data
}

// ----------------------------------------------------------------

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
		data: &ldhHashMapData{
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
	for index := 0; index != 1; index++ {
		rand.Seed(time.Now().UnixNano())
		keyValueMap := make(map[int]int)
		for index := 0; index != DEFAULT_HASH_MAP_SIZE>>7; index++ {
			for {
				k := rand.Intn(DEFAULT_HASH_MAP_SIZE) + 1
				if _, hasK := keyValueMap[k]; !hasK {
					keyValueMap[k] = index
					fmt.Printf("generate Key:Value = [%v:%v]\n", k, index)
					break
				}
			}
		}

		// keyValueMap = map[int]int{
		// 	3:   0,
		// 	974: 1,
		// 	554: 2,
		// 	875: 3,
		// 	44:  4,
		// 	834: 5,
		// 	296: 6,
		// 	458: 7,
		// }

		// hashMapTest(keyValueMap, WithHashMapSize(DEFAULT_HASH_MAP_SIZE>>9))
		// hashMapTest(keyValueMap, WithHashMapData(&sdhHashMapData{
		// 	array: make([]*HashValue, DEFAULT_HASH_MAP_SIZE>>9),
		// }))
		// hashMapTest(keyValueMap, WithHashMapData(&dllHashMapData{
		// 	buckets: make([]*dllNode, DEFAULT_HASH_MAP_SIZE>>9),
		// }))
		hashMapTest(keyValueMap, WithHashMapData(&bstHashMapData{
			buckets: make([]*bstNode, DEFAULT_HASH_MAP_SIZE>>10),
		}))
	}
}

type debugData struct {
	kvMap       map[int]int
	setSlice    []int
	delSlice    []int
	dataPreview strings.Builder
}

func hashMapTest(keyValueMap map[int]int, options ...HashMapOption) {
	testHashMap := MakeHashMap(options...)

	debugData := debugData{}
	debugData.kvMap = keyValueMap

	for key, value := range keyValueMap {
		fmt.Printf("testHashMap.Set(%v, %v)\n", key, value)
		debugData.setSlice = append(debugData.setSlice, key)
		if ok := testHashMap.Set(key, value); !ok {
			fmt.Printf("testHashMap.Set(%v, %v) failed\n", key, value)
		} else {
			// fmt.Printf("after testHashMap.Set(%v, %v), testHashMap load factor is %v\n", key, value, testHashMap.GetLoadFactor(0))
			// fmt.Printf("testHashMap.Set(%v, %v) at hash index %v success\n", key, value, hashIndex)
		}
	}

	// insertKeySlice := []int{10, 9} // 单向左子树
	// insertKeySlice := []int{9, 10} // 单向右子树
	// insertKeySlice := []int{974, 554, 875, 44, 834, 296, 458, 3}
	// for _, key := range insertKeySlice {
	// 	fmt.Printf("testHashMap.Set(%v, %v)\n", key, keyValueMap[key])
	// 	if ok := testHashMap.Set(key, keyValueMap[key]); !ok {
	// 		fmt.Printf("testHashMap.Set(%v, %v) failed\n", key, keyValueMap[key])
	// 	} else {
	// 		// fmt.Printf("after testHashMap.Set(%v, %v), testHashMap load factor is %v\n", key, value, testHashMap.GetLoadFactor(0))
	// 		// fmt.Printf("testHashMap.Set(%v, %v) at hash index %v success\n", key, value, hashIndex)
	// 	}
	// }

	testHashMap.Range(func(k, v int) bool {
		fmt.Printf("range key: %v, value: %v\n", k, v)
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
		fmt.Printf("testHashMap.Del(%v)\n", key)
		debugData.delSlice = append(debugData.delSlice, key)
		if _value, hasKey := testHashMap.Del(key); !hasKey || _value != value {
			fmt.Printf("testHashMap.Del(%v), not has key or store value %v not equal to origin value %v\n", key, _value, value)
			t := time.Now().UnixNano()
			outputFile, openError := os.Create(fmt.Sprintf("%v.log", t))
			if openError != nil {
				fmt.Printf("Error: open file occurs error: %v\n", openError)
				return
			}
			outputFile.WriteString(fmt.Sprintf("key value map: %v\n", debugData.kvMap))
			outputFile.WriteString(fmt.Sprintf("set slice: %v\n", debugData.setSlice))
			// outputFile.Write()
			outputFile.WriteString(fmt.Sprintf("del slice: %v\n", debugData.delSlice))
		} else {
			// fmt.Printf("after Del, testHashMap load factor is %v\n", testHashMap.GetLoadFactor(0))
			// fmt.Printf("testHashMap.Del(%v), key and store value equal to origin value %v\n", key, value)
		}
	}

	// for _, key := range []int{875, 974, 554, 44, 834, 296, 458, 3} {
	// 	fmt.Printf("testHashMap.Del(%v)\n", key)
	// 	if _value, hasKey := testHashMap.Del(key); !hasKey || _value != keyValueMap[key] {
	// 		fmt.Printf("testHashMap.Del(%v), not has key or store value %v not equal to origin value %v\n", key, _value, keyValueMap[key])
	// 	} else {
	// 		// fmt.Printf("after Del, testHashMap load factor is %v\n", testHashMap.GetLoadFactor(0))
	// 		// fmt.Printf("testHashMap.Del(%v), key and store value equal to origin value %v\n", key, value)
	// 	}
	// }

	testHashMap.Range(func(k, v int) bool {
		fmt.Printf("key: %v, value: %v\n", k, v)
		return true
	})
}
