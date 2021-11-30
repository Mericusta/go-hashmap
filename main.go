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
	leftChild  *bstNode
	rightChild *bstNode
	value      *HashValue
}

func (n *bstNode) preOrderTraversal(op func(*HashValue) bool, deep int) bool {
	fmt.Printf("%v", strings.Repeat("\t", deep))
	if !op(n.value) {
		return false
	}
	if n.leftChild != nil {
		n.leftChild.preOrderTraversal(op, deep+1)
	}
	if n.rightChild != nil {
		n.rightChild.preOrderTraversal(op, deep+1)
	}
	return true
}

func (n *bstNode) inOrderTraversal(op func(*HashValue) bool) bool {
	if n.leftChild != nil {
		n.leftChild.inOrderTraversal(op)
	}
	if !op(n.value) {
		return false
	}
	if n.rightChild != nil {
		n.rightChild.inOrderTraversal(op)
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
				if node.leftChild == nil {
					return 0, false
				} else {
					node = node.leftChild
				}
			} else if node.value.k < key {
				if node.rightChild == nil {
					return 0, false
				} else {
					node = node.rightChild
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
				if node.leftChild == nil {
					node.leftChild = vNode
					return true
				} else {
					node = node.leftChild
				}
			} else if node.value.k < hashValue.k {
				if node.rightChild == nil {
					node.rightChild = vNode
					return true
				} else {
					node = node.rightChild
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
				if node.leftChild == nil {
					return 0, true
				} else {
					parentNode = node
					node = node.leftChild
				}
			} else if node.value.k < key {
				if node.rightChild == nil {
					return 0, true
				} else {
					parentNode = node
					node = node.rightChild
				}
			} else {
				value := node.value.v
				if parentNode == nil {
					d.buckets[hashIndex] = nil
				} else {
					if parentNode.leftChild == node {
						if node.rightChild == nil {
							parentNode.leftChild = node.leftChild
						} else {
							parentNode.leftChild = node.rightChild
							if node.leftChild != nil {
								leftSubNode := node.leftChild
								for node = node.rightChild; node.leftChild != nil; node = node.leftChild {
								}
								node.leftChild = leftSubNode
							}
						}
					} else if parentNode.rightChild == node {
						if node.rightChild == nil {
							parentNode.rightChild = node.leftChild
						} else {
							parentNode.rightChild = node.rightChild
							if node.leftChild != nil {
								leftSubNode := node.leftChild
								for node = node.rightChild; node.leftChild != nil; node = node.leftChild {
								}
								node.leftChild = leftSubNode
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
				if node.leftChild == nil {
					return 0, true
				} else {
					parentNode = node
					node = node.leftChild
				}
			} else if node.value.k < key {
				if node.rightChild == nil {
					return 0, true
				} else {
					parentNode = node
					node = node.rightChild
				}
			} else {
				value, deleteNode := node.value.v, node

				var newNode *bstNode
				leftChild := node.leftChild
				rightChild := node.rightChild
				minRightNodeParentNode := node
				for node = node.rightChild; node != nil && node.leftChild != nil; minRightNodeParentNode, node = node, node.leftChild {
				}
				if node == nil { // 单左链表
					newNode = leftChild
				} else if minRightNodeParentNode == deleteNode { // 单右链表
					newNode = deleteNode.rightChild
					newNode.leftChild = leftChild
				} else if minRightNodeParentNode != node {
					minRightNodeParentNode.leftChild = node.rightChild
					node.leftChild = leftChild
					node.rightChild = rightChild
					newNode = node
				} else {
					newNode = node.rightChild
				}

				if parentNode == nil {
					d.buckets[hashIndex] = newNode
				} else if parentNode.leftChild == deleteNode {
					parentNode.leftChild = newNode
				} else if parentNode.rightChild == deleteNode {
					parentNode.rightChild = newNode
				} else {
					// TODO: error
					return 0, false
				}

				deleteNode.leftChild = nil
				deleteNode.rightChild = nil

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

// avl tree

type avltNode struct {
	parentNode  *avltNode
	leftHeight  int
	leftChild   *avltNode
	rightHeight int
	rightChild  *avltNode
	value       *HashValue
}

// func (n *avltNode) preOrderTraversal(op func(*HashValue, int, int) bool, deep int) bool {
// 	fmt.Printf("%v", strings.Repeat("\t", deep))
// 	if !op(n.value, n.leftHeight, n.rightHeight) {
// 		return false
// 	}
// 	if n.leftChild != nil {
// 		n.leftChild.preOrderTraversal(op, deep+1)
// 	}
// 	if n.rightChild != nil {
// 		n.rightChild.preOrderTraversal(op, deep+1)
// 	}
// 	return true
// }

func (n *avltNode) preOrderTraversal(op func(*HashValue) bool, deep int) bool {
	fmt.Printf("%v", strings.Repeat("\t", deep))
	if !op(n.value) {
		return false
	}
	if n.leftChild != nil {
		n.leftChild.preOrderTraversal(op, deep+1)
	}
	if n.rightChild != nil {
		n.rightChild.preOrderTraversal(op, deep+1)
	}
	return true
}

func (n *avltNode) inOrderTraversal(op func(*HashValue) bool) bool {
	if n.leftChild != nil {
		n.leftChild.inOrderTraversal(op)
	}
	if !op(n.value) {
		return false
	}
	if n.rightChild != nil {
		n.rightChild.inOrderTraversal(op)
	}
	return true
}

// TODO: just check, no update
func (n *avltNode) checkBalance(height int) (*avltNode, int) {
	if n.parentNode == nil {
		return nil, 0
	}
	if n.parentNode.leftChild == n {
		n.parentNode.leftHeight = height
	} else if n.parentNode.rightChild == n {
		n.parentNode.rightHeight = height
	}
	if diff := n.parentNode.leftHeight - n.parentNode.rightHeight; diff < -1 || 1 < diff {
		fmt.Printf("node %v lost balance\n", n.parentNode.value.k)
		n.parentNode.preOrderTraversal(func(h *HashValue) bool {
			fmt.Printf("DEBUG: range key: %v, value: %v\n", h.k, h.v)
			return true
		}, 0)
		return n.parentNode, diff
	}
	return n.parentNode.checkBalance(height + 1)
}

type rotateType int

const (
	UNKNOWN rotateType = iota
	LL
	LR
	RL
	RR
)

func (n *avltNode) getRotateType(childNode *avltNode) rotateType {
	if n.leftChild != nil {
		if n.leftChild.leftChild != nil {
			if n.leftChild.leftChild == childNode || n.leftChild.leftChild.leftChild == childNode || n.leftChild.leftChild.rightChild == childNode {
				return LL
			}
		}
		if n.leftChild.rightChild != nil {
			if n.leftChild.rightChild == childNode || n.leftChild.rightChild.leftChild == childNode || n.leftChild.rightChild.rightChild == childNode {
				return LR
			}
		}
	}
	if n.rightChild != nil {
		if n.rightChild.rightChild != nil {
			if n.rightChild.rightChild == childNode || n.rightChild.rightChild.leftChild == childNode || n.rightChild.rightChild.rightChild == childNode {
				return RR
			}
		}
		if n.rightChild.leftChild != nil {
			if n.rightChild.leftChild == childNode || n.rightChild.leftChild.leftChild == childNode || n.rightChild.leftChild.rightChild == childNode {
				return RL
			}
		}
	}
	return UNKNOWN
}

func (n *avltNode) setLeftChild(childNode *avltNode) {
	n.leftChild = childNode
	if childNode != nil {
		n.leftHeight = childNode.getHeight() + 1
	} else {
		n.leftHeight = 0
	}
}

func (n *avltNode) setRightChild(childNode *avltNode) {
	n.rightChild = childNode
	if childNode != nil {
		n.rightHeight = childNode.getHeight() + 1
	} else {
		n.rightHeight = 0
	}
}

func (n *avltNode) getHeight() int {
	if n.leftHeight < n.rightHeight {
		return n.rightHeight
	}
	return n.leftHeight
}

//  5         7
// 1 7   ->  5 8
//  6 8     1 6 9
//     9
func (n *avltNode) leftRotate() *avltNode {
	newRootNode := n.rightChild
	if newRootNode != nil {
		n.setRightChild(newRootNode.leftChild)
		if newRootNode.leftChild != nil {
			newRootNode.leftChild.parentNode = n
		}
		newRootNode.setLeftChild(n)
		n.parentNode = newRootNode
	} else {
		n.setRightChild(nil)
	}
	return newRootNode
}

//    5        3
//   3 9  ->  2 5
//  2 4      1 4 9
// 1
func (n *avltNode) rightRotate() *avltNode {
	newRootNode := n.leftChild
	if newRootNode != nil {
		n.setLeftChild(newRootNode.rightChild)
		if newRootNode.rightChild != nil {
			newRootNode.rightChild.parentNode = n
		}
		newRootNode.setRightChild(n)
		n.parentNode = newRootNode
	} else {
		n.setLeftChild(nil)
	}
	return newRootNode
}

type avltHashMapData struct {
	buckets []*avltNode
}

func (d *avltHashMapData) Len() int {
	return len(d.buckets)
}

func (d *avltHashMapData) Get(hashIndex, key int) (int, bool) {
	if d.buckets[hashIndex] == nil {
		return 0, false
	} else {
		node := d.buckets[hashIndex]
		for {
			if key < node.value.k {
				if node.leftChild == nil {
					return 0, false
				} else {
					node = node.leftChild
				}
			} else if node.value.k < key {
				if node.rightChild == nil {
					return 0, false
				} else {
					node = node.rightChild
				}
			} else {
				return node.value.v, true
			}
		}
	}
}

func (d *avltHashMapData) Set(hashIndex int, hashValue *HashValue) bool {
	vNode := &avltNode{
		value: hashValue,
	}
	if d.buckets[hashIndex] == nil {
		d.buckets[hashIndex] = vNode
	} else {
		node := d.buckets[hashIndex]
		for {
			if hashValue.k < node.value.k {
				if node.leftChild == nil {
					vNode.parentNode = node
					node.leftChild = vNode
					break
				} else {
					node = node.leftChild
				}
			} else if node.value.k < hashValue.k {
				if node.rightChild == nil {
					vNode.parentNode = node
					node.rightChild = vNode
					break
				} else {
					node = node.rightChild
				}
			} else {
				node.value = hashValue
				return true
			}
		}
	}
	lostBalanceNode, _ := vNode.checkBalance(1)
	if lostBalanceNode != nil {
		lostBalanceNodeParent := lostBalanceNode.parentNode
		var newRootNode *avltNode
		rotateType := lostBalanceNode.getRotateType(vNode)
		fmt.Printf("avl-tree need change to keep balance, rotate type %v\n", rotateType)
		switch rotateType {
		case LR:
			lostBalanceNode.setLeftChild(lostBalanceNode.leftChild.leftRotate())
			lostBalanceNode.leftChild.parentNode = lostBalanceNode
			fallthrough
		case LL:
			newRootNode = lostBalanceNode.rightRotate()
		case RL:
			lostBalanceNode.setRightChild(lostBalanceNode.rightChild.rightRotate())
			lostBalanceNode.rightChild.parentNode = lostBalanceNode
			fallthrough
		case RR:
			newRootNode = lostBalanceNode.leftRotate()
		default:
			fmt.Printf("Error: lost balance node rotate type wrong\n")
			lostBalanceNode.preOrderTraversal(func(h *HashValue) bool {
				fmt.Printf("DEBUG: range key: %v, value: %v\n", h.k, h.v)
				return true
			}, 0)
		}

		if lostBalanceNodeParent == nil {
			d.buckets[hashIndex] = newRootNode
			d.buckets[hashIndex].parentNode = nil
		} else {
			if rotateType == LR || rotateType == LL {
				lostBalanceNodeParent.setLeftChild(newRootNode)
				lostBalanceNodeParent.leftChild.parentNode = lostBalanceNodeParent
			} else if rotateType == RL || rotateType == RR {
				lostBalanceNodeParent.setRightChild(newRootNode)
				lostBalanceNodeParent.rightChild.parentNode = lostBalanceNodeParent
			}
		}
	}

	// d.buckets[hashIndex].preOrderTraversal(func(h *HashValue, leftHeight, rightHeight int) bool {
	// 	fmt.Printf("DEBUG: range key: %v, value: %v, left height: %v, right height: %v\n", h.k, h.v, leftHeight, rightHeight)
	// 	return true
	// }, 0)

	return true
}

func (d *avltHashMapData) Del(hashIndex, key int) (int, bool) {
	return 0, false
}

func (d *avltHashMapData) Range(op func(*HashValue) bool) {
	for _, bucket := range d.buckets {
		if bucket != nil {
			fmt.Println()
			fmt.Printf("inOrderTraversal\n")
			if !bucket.inOrderTraversal(op) {
				return
			}
			fmt.Println()
			fmt.Printf("preOrderTraversal\n")
			if !bucket.preOrderTraversal(op, 0) {
				return
			}
		}
	}
}

func (d *avltHashMapData) Reallocate(size uint) {
	if uint(len(d.buckets)) != size {
		d.buckets = make([]*avltNode, size)
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
		fmt.Println()
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

		keyValueMap = map[int]int{
			27:  0,
			283: 1,
			379: 2,
			767: 3,
			4:   4,
			463: 5,
			930: 6,
			444: 7,
		}

		// hashMapTest(keyValueMap, WithHashMapSize(DEFAULT_HASH_MAP_SIZE>>9))
		// hashMapTest(keyValueMap, WithHashMapData(&sdhHashMapData{
		// 	array: make([]*HashValue, DEFAULT_HASH_MAP_SIZE>>9),
		// }))
		// hashMapTest(keyValueMap, WithHashMapData(&dllHashMapData{
		// 	buckets: make([]*dllNode, DEFAULT_HASH_MAP_SIZE>>9),
		// }))
		// hashMapTest(keyValueMap, WithHashMapData(&bstHashMapData{
		// 	buckets: make([]*bstNode, DEFAULT_HASH_MAP_SIZE>>10),
		// }))
		hashMapTest(keyValueMap, WithHashMapData(&avltHashMapData{
			buckets: make([]*avltNode, DEFAULT_HASH_MAP_SIZE>>10),
		}))
	}
}

type debugData struct {
	kvMap       map[int]int
	setSlice    []int
	delSlice    []int
	dataPreview strings.Builder
}

func (d debugData) outputFile() {
	t := time.Now().UnixNano()
	outputFile, openError := os.Create(fmt.Sprintf("%v.log", t))
	if openError != nil {
		fmt.Printf("Error: open file occurs error: %v\n", openError)
		return
	}
	outputFile.WriteString(fmt.Sprintf("key value map: %v\n", d.kvMap))
	outputFile.WriteString(fmt.Sprintf("set slice: %v\n", d.setSlice))
	outputFile.WriteString(fmt.Sprintf("del slice: %v\n", d.delSlice))
	outputFile.Close()
}

func hashMapTest(keyValueMap map[int]int, options ...HashMapOption) {
	testHashMap := MakeHashMap(options...)

	debugData := debugData{}
	debugData.kvMap = keyValueMap

	fmt.Println()

	// for key, value := range keyValueMap {
	// 	fmt.Printf("testHashMap.Set(%v, %v)\n", key, value)
	// 	debugData.setSlice = append(debugData.setSlice, key)
	// 	if ok := testHashMap.Set(key, value); !ok {
	// 		fmt.Printf("testHashMap.Set(%v, %v) failed\n", key, value)
	// 	} else {
	// 		// fmt.Printf("after testHashMap.Set(%v, %v), testHashMap load factor is %v\n", key, value, testHashMap.GetLoadFactor(0))
	// 		// fmt.Printf("testHashMap.Set(%v, %v) at hash index %v success\n", key, value, hashIndex)
	// 	}
	// }

	insertKeySlice := []int{4, 463, 930, 444, 27, 283, 379, 767}
	for _, key := range insertKeySlice {
		fmt.Printf("testHashMap.Set(%v, %v)\n", key, keyValueMap[key])
		if ok := testHashMap.Set(key, keyValueMap[key]); !ok {
			fmt.Printf("testHashMap.Set(%v, %v) failed\n", key, keyValueMap[key])
		} else {
			// fmt.Printf("after testHashMap.Set(%v, %v), testHashMap load factor is %v\n", key, value, testHashMap.GetLoadFactor(0))
			// fmt.Printf("testHashMap.Set(%v, %v) at hash index %v success\n", key, value, hashIndex)
		}
	}

	debugData.outputFile()

	testHashMap.Range(func(k, v int) bool {
		fmt.Printf("range key: %v, value: %v\n", k, v)
		return true
	})

	// for key, value := range keyValueMap {
	// 	if _value, hasKey := testHashMap.Get(key); !hasKey || _value != value {
	// 		fmt.Printf("testHashMap.Get(%v), not has key or store value %v not equal to origin value %v\n", key, _value, value)
	// 	} else {
	// 		// fmt.Printf("testHashMap.Get(%v), key and store value equal to origin value %v\n", key, value)
	// 	}
	// }

	// for key, value := range keyValueMap {
	// 	fmt.Printf("testHashMap.Del(%v)\n", key)
	// 	debugData.delSlice = append(debugData.delSlice, key)
	// 	if _value, hasKey := testHashMap.Del(key); !hasKey || _value != value {
	// 		fmt.Printf("testHashMap.Del(%v), not has key or store value %v not equal to origin value %v\n", key, _value, value)
	// 		t := time.Now().UnixNano()
	// 		outputFile, openError := os.Create(fmt.Sprintf("%v.log", t))
	// 		if openError != nil {
	// 			fmt.Printf("Error: open file occurs error: %v\n", openError)
	// 			return
	// 		}
	// 		outputFile.WriteString(fmt.Sprintf("key value map: %v\n", debugData.kvMap))
	// 		outputFile.WriteString(fmt.Sprintf("set slice: %v\n", debugData.setSlice))
	// 		// outputFile.Write()
	// 		outputFile.WriteString(fmt.Sprintf("del slice: %v\n", debugData.delSlice))
	// 	} else {
	// 		// fmt.Printf("after Del, testHashMap load factor is %v\n", testHashMap.GetLoadFactor(0))
	// 		// fmt.Printf("testHashMap.Del(%v), key and store value equal to origin value %v\n", key, value)
	// 	}
	// }

	// for _, key := range []int{875, 974, 554, 44, 834, 296, 458, 3} {
	// 	fmt.Printf("testHashMap.Del(%v)\n", key)
	// 	if _value, hasKey := testHashMap.Del(key); !hasKey || _value != keyValueMap[key] {
	// 		fmt.Printf("testHashMap.Del(%v), not has key or store value %v not equal to origin value %v\n", key, _value, keyValueMap[key])
	// 	} else {
	// 		// fmt.Printf("after Del, testHashMap load factor is %v\n", testHashMap.GetLoadFactor(0))
	// 		// fmt.Printf("testHashMap.Del(%v), key and store value equal to origin value %v\n", key, value)
	// 	}
	// }

	// testHashMap.Range(func(k, v int) bool {
	// 	fmt.Printf("key: %v, value: %v\n", k, v)
	// 	return true
	// })
}
