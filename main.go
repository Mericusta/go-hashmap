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

// doubly linked list - DLL

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

// binary search tree - BST

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
		// fmt.Println()
		// fmt.Printf("Before Delete %v preOrder\n", key)
		// d.buckets[hashIndex].preOrderTraversal(func(h *HashValue) bool {
		// 	fmt.Printf("DEBUG: range key: %v, value: %v\n", h.k, h.v)
		// 	return true
		// }, 0)
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

// avl tree - AVLT

type avltNode struct {
	parentNode  *avltNode
	leftHeight  int
	leftChild   *avltNode
	rightHeight int
	rightChild  *avltNode
	value       *HashValue
}

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

func (n *avltNode) preOrderTraversalWithHeight(op func(*HashValue, int, int) bool, deep int) bool {
	fmt.Printf("%v", strings.Repeat("\t", deep))
	if !op(n.value, n.leftHeight, n.rightHeight) {
		return false
	}
	if n.leftChild != nil {
		n.leftChild.preOrderTraversalWithHeight(op, deep+1)
	}
	if n.rightChild != nil {
		n.rightChild.preOrderTraversalWithHeight(op, deep+1)
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

// TODO: just check, no update -> rebalance just update & checkBalance just check
// checkAndRebalance 向上检查平衡并且再平衡
func (n *avltNode) checkAndRebalance(height int) *avltNode {
	if n.parentNode == nil {
		return nil
	}
	if n.parentNode.leftChild == n {
		n.parentNode.leftHeight = height
	} else if n.parentNode.rightChild == n {
		n.parentNode.rightHeight = height
	}
	if diff := n.parentNode.leftHeight - n.parentNode.rightHeight; diff < -1 || 1 < diff {
		fmt.Printf("node %v lost balance\n", n.parentNode.value.k)
		n.parentNode.preOrderTraversalWithHeight(func(h *HashValue, leftHeight, rightHeight int) bool {
			fmt.Printf("DEBUG: range key: %v, value: %v, leftHeight = %v, rightHeight = %v\n", h.k, h.v, leftHeight, rightHeight)
			return true
		}, 0)
		return n.parentNode
	}
	return n.parentNode.checkAndRebalance(height + 1)
}

// rebalance 向下再平衡
func (n *avltNode) rebalance() int {
	if n.leftChild != nil {
		n.leftHeight = n.leftChild.rebalance() + 1
	} else {
		n.leftHeight = 0
	}
	if n.rightChild != nil {
		n.rightHeight = n.rightChild.rebalance() + 1
	} else {
		n.rightHeight = 0
	}
	return n.getHeight()
}

// checkBalance 向下检查平衡
func (n *avltNode) checkBalance() *avltNode {
	var leftLostBalanceNode, rightLostBalanceNode *avltNode
	if n.leftChild != nil {
		leftLostBalanceNode = n.leftChild.checkBalance()
	}
	if n.rightChild != nil {
		rightLostBalanceNode = n.rightChild.checkBalance()
	}
	if leftLostBalanceNode != nil && rightLostBalanceNode != nil {
		fmt.Printf("node %v left %v and right %v node both lost balance\n", n.value.k, leftLostBalanceNode.value.k, rightLostBalanceNode.value.k)
		panic(fmt.Sprintf("node %v left %v and right %v node both lost balance\n", n.value.k, leftLostBalanceNode.value.k, rightLostBalanceNode.value.k))
	} else if leftLostBalanceNode != nil {
		return leftLostBalanceNode
	} else if rightLostBalanceNode != nil {
		return rightLostBalanceNode
	}
	if diff := n.leftHeight - n.rightHeight; diff < -1 || 1 < diff {
		// fmt.Printf("DEBUG: checkBalance node %v lost balance\n", n.value.k)
		// n.preOrderTraversalWithHeight(func(h *HashValue, leftHeight, rightHeight int) bool {
		// 	fmt.Printf("DEBUG: range key: %v, value: %v, leftHeight = %v, rightHeight = %v\n", h.k, h.v, leftHeight, rightHeight)
		// 	return true
		// }, 0)
		return n
	}
	return nil
}

type rotateType int

const (
	UNKNOWN rotateType = iota
	LL
	LR
	RL
	RR
)

func (n *avltNode) getRotateType() rotateType {
	factor := n.getBalanceFactor()
	if n.leftChild != nil {
		if factor > 1 && n.leftChild.getBalanceFactor() >= 0 {
			return LL
		}
		if factor > 1 && n.leftChild.getBalanceFactor() < 0 {
			return LR
		}
	}
	if n.rightChild != nil {
		if factor < -1 && n.rightChild.getBalanceFactor() <= 0 {
			return RR
		}
		if factor < -1 && n.rightChild.getBalanceFactor() > 0 {
			return RL
		}
	}
	return UNKNOWN
}

func (n *avltNode) getRotateTypeByTargetNode(childNode *avltNode) rotateType {
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
		childNode.parentNode = n
		n.leftHeight = childNode.getHeight() + 1
	} else {
		n.leftHeight = 0
	}
}

func (n *avltNode) setRightChild(childNode *avltNode) {
	n.rightChild = childNode
	if childNode != nil {
		childNode.parentNode = n
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

func (n *avltNode) getBalanceFactor() int {
	return n.leftHeight - n.rightHeight
}

func (n *avltNode) leftRotate() *avltNode {
	newRootNode := n.rightChild
	if newRootNode != nil {
		n.setRightChild(newRootNode.leftChild)
		newRootNode.setLeftChild(n)
	} else {
		n.setRightChild(nil)
	}
	return newRootNode
}

func (n *avltNode) rightRotate() *avltNode {
	newRootNode := n.leftChild
	if newRootNode != nil {
		n.setLeftChild(newRootNode.rightChild)
		newRootNode.setRightChild(n)
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

	// balance
	// fmt.Println()
	// fmt.Printf("root node %v rebalance\n", d.buckets[hashIndex].value.k)
	// d.buckets[hashIndex].rebalance() // 自根节点向下再平衡
	// d.buckets[hashIndex].preOrderTraversalWithHeight(func(h *HashValue, leftHeight, rightHeight int) bool {
	// 	fmt.Printf("DEBUG: range key: %v, value: %v, left height: %v, right height: %v\n", h.k, h.v, leftHeight, rightHeight)
	// 	return true
	// }, 0)

	// fmt.Println()
	// fmt.Printf("root node %v checkBalance\n", d.buckets[hashIndex].value.k)
	// lostBalanceNode := d.buckets[hashIndex].checkBalance() // 自根节点向下检查平衡
	// if lostBalanceNode != nil {
	// 	lostBalanceNode.preOrderTraversalWithHeight(func(h *HashValue, leftHeight, rightHeight int) bool {
	// 		fmt.Printf("DEBUG: range key: %v, value: %v, left height: %v, right height: %v\n", h.k, h.v, leftHeight, rightHeight)
	// 		return true
	// 	}, 0)
	// }

	// fmt.Println()
	// fmt.Printf("vNode %v checkAndRebalance(1)\n", vNode.value.k)
	lostBalanceNode := vNode.checkAndRebalance(1) // 自插入节点向上检查平衡并且再平衡
	// if lostBalanceNode != nil {
	// 	lostBalanceNode.preOrderTraversalWithHeight(func(h *HashValue, leftHeight, rightHeight int) bool {
	// 		fmt.Printf("DEBUG: range key: %v, value: %v, left height: %v, right height: %v\n", h.k, h.v, leftHeight, rightHeight)
	// 		return true
	// 	}, 0)
	// }
	// if lostBalanceNode != lostBalanceNode {
	// 	panic(fmt.Sprintf("lostBalanceNode %v != lostBalanceNode %v, root node %v, vNode %v\n", lostBalanceNode.value.k, lostBalanceNode.value.k, d.buckets[hashIndex].value.k, vNode.value.k))
	// }

	if lostBalanceNode != nil {
		lostBalanceNodeParent := lostBalanceNode.parentNode
		var newRootNode *avltNode
		rotateType := lostBalanceNode.getRotateType()
		// rotateTypeByTargetNode := lostBalanceNode.getRotateTypeByTargetNode(vNode)
		// if rotateType != rotateTypeByTargetNode {
		// 	panic(fmt.Sprintf("lostBalanceNode %v getRotateType() %v != getRotateTypeByTargetNode(%v) %v", lostBalanceNode.value.k, vNode.value.k, rotateType, rotateTypeByTargetNode))
		// }
		// fmt.Printf("avl-tree need change to keep balance, rotate type %v\n", rotateType)
		switch rotateType {
		case LR:
			lostBalanceNode.setLeftChild(lostBalanceNode.leftChild.leftRotate())
			fallthrough
		case LL:
			newRootNode = lostBalanceNode.rightRotate()
		case RL:
			lostBalanceNode.setRightChild(lostBalanceNode.rightChild.rightRotate())
			fallthrough
		case RR:
			newRootNode = lostBalanceNode.leftRotate()
		default:
			fmt.Printf("Error: lost balance node %v rotate type wrong\n", lostBalanceNode)
			lostBalanceNode.preOrderTraversal(func(h *HashValue) bool {
				fmt.Printf("DEBUG: range key: %v, value: %v\n", h.k, h.v)
				return true
			}, 0)
			panic(fmt.Sprintf("Error: lost balance node %v rotate type wrong\n", lostBalanceNode))
		}

		if lostBalanceNodeParent == nil {
			d.buckets[hashIndex] = newRootNode
			d.buckets[hashIndex].parentNode = nil
		} else {
			if lostBalanceNodeParent.leftChild == lostBalanceNode {
				lostBalanceNodeParent.setLeftChild(newRootNode)
			} else if lostBalanceNodeParent.rightChild == lostBalanceNode {
				lostBalanceNodeParent.setRightChild(newRootNode)
			} else {
				fmt.Printf("Error: lost balance node %v is not exists in its parent %v child\n", lostBalanceNode.value.k, lostBalanceNodeParent.value.k)
				panic(fmt.Sprintf("Error: lost balance node %v is not exists in its parent %v child\n", lostBalanceNode.value.k, lostBalanceNodeParent.value.k))
			}
		}
	}

	// d.buckets[hashIndex].preOrderTraversalWithHeight(func(h *HashValue, leftHeight, rightHeight int) bool {
	// 	fmt.Printf("DEBUG: range key: %v, value: %v, left height: %v, right height: %v\n", h.k, h.v, leftHeight, rightHeight)
	// 	return true
	// }, 0)

	return true
}

// type 1 delete leaf
//   7              7
//  5 8  -> Del(1) 5 8
// 1 6 9            6 9
// type 3 delete node only has right tree
//   7               7
//  5 8  -> Del(8)  5 9
// 1 6 9           1 6
// type 4 delete node has both left and right tree
//   7               7
//  5 8  -> Del(5)  6 8
// 1 6 9           1   9
func (d *avltHashMapData) Del(hashIndex, key int) (int, bool) {
	if d.buckets[hashIndex] == nil {
		return 0, true
	} else {
		// fmt.Println()
		// fmt.Printf("Before Delete %v preOrder\n", key)
		// d.buckets[hashIndex].preOrderTraversalWithHeight(func(h *HashValue, leftHeight, rightHeight int) bool {
		// 	fmt.Printf("DEBUG: range key: %v, value: %v, left height: %v, right height: %v\n", h.k, h.v, leftHeight, rightHeight)
		// 	return true
		// }, 0)
		var parentNode *avltNode
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
				value, deleteNode := node.value.v, node
				var newNode *avltNode
				leftChild := node.leftChild
				rightChild := node.rightChild
				minRightNodeParentNode := node
				for node = node.rightChild; node != nil && node.leftChild != nil; minRightNodeParentNode, node = node, node.leftChild {
				}
				if node == nil { // 单左链表
					newNode = leftChild
				} else if minRightNodeParentNode == deleteNode { // 单右链表
					newNode = deleteNode.rightChild
					newNode.setLeftChild(leftChild)
				} else if minRightNodeParentNode != node {
					minRightNodeParentNode.setLeftChild(node.rightChild)
					node.setLeftChild(leftChild)
					node.setRightChild(rightChild)
					newNode = node
				} else {
					newNode = node.rightChild
				}

				var checkNode *avltNode
				if parentNode == nil {
					d.buckets[hashIndex] = newNode
					if newNode == nil {
						return value, true
					}
					newNode.parentNode = nil
					checkNode = d.buckets[hashIndex]
				} else if parentNode.leftChild == deleteNode {
					parentNode.setLeftChild(newNode)
					checkNode = parentNode
				} else if parentNode.rightChild == deleteNode {
					parentNode.setRightChild(newNode)
					checkNode = parentNode
				} else {
					panic(fmt.Sprintf("new node %v does has parent node %v but parent node not has new node\n", newNode.value.k, parentNode.value.k))
				}

				deleteNode.leftChild = nil
				deleteNode.rightChild = nil

				if d.buckets[hashIndex] != nil {
					fmt.Printf("After Delete %v preOrder\n", key)
					d.buckets[hashIndex].preOrderTraversalWithHeight(func(h *HashValue, leftHeight, rightHeight int) bool {
						fmt.Printf("DEBUG: range key: %v, value: %v, left height: %v, right height: %v\n", h.k, h.v, leftHeight, rightHeight)
						return true
					}, 0)
				}

				// balance
				// fmt.Println()
				// fmt.Printf("node %v rebalance\n", checkNode.value.k)
				checkNode.rebalance() // 自变更节点向下再平衡
				// checkNode.preOrderTraversalWithHeight(func(h *HashValue, leftHeight, rightHeight int) bool {
				// 	fmt.Printf("DEBUG: range key: %v, value: %v, left height: %v, right height: %v\n", h.k, h.v, leftHeight, rightHeight)
				// 	return true
				// }, 0)

				// fmt.Println()
				// fmt.Printf("node %v checkBalance\n", checkNode.value.k)
				lostBalanceNode := checkNode.checkBalance() // 自变更节点向下检查平衡
				// if lostBalanceNode != nil {
				// 	lostBalanceNode.preOrderTraversalWithHeight(func(h *HashValue, leftHeight, rightHeight int) bool {
				// 		fmt.Printf("DEBUG: range key: %v, value: %v, left height: %v, right height: %v\n", h.k, h.v, leftHeight, rightHeight)
				// 		return true
				// 	}, 0)
				// }

				if lostBalanceNode == nil {
					// fmt.Println()
					// fmt.Printf("node %v checkAndRebalance(1)\n", vNode.value.k)
					lostBalanceNode = checkNode.checkAndRebalance(checkNode.getHeight() + 1) // 自变更节点向上检查平衡并且再平衡
					// if lostBalanceNode != nil {
					// 	lostBalanceNode.preOrderTraversalWithHeight(func(h *HashValue, leftHeight, rightHeight int) bool {
					// 		fmt.Printf("DEBUG: range key: %v, value: %v, left height: %v, right height: %v\n", h.k, h.v, leftHeight, rightHeight)
					// 		return true
					// 	}, 0)
					// }
					// if lostBalanceNode != lostBalanceNode {
					// 	panic(fmt.Sprintf("lostBalanceNode %v != lostBalanceNode %v, root node %v, vNode %v\n", lostBalanceNode.value.k, lostBalanceNode.value.k, d.buckets[hashIndex].value.k, vNode.value.k))
					// }
				}

				if lostBalanceNode != nil {
					lostBalanceNodeParent := lostBalanceNode.parentNode
					var newRootNode *avltNode
					rotateType := lostBalanceNode.getRotateType()
					// fmt.Printf("avl-tree need change to keep balance, rotate type %v\n", rotateType)
					switch rotateType {
					case LR:
						lostBalanceNode.setLeftChild(lostBalanceNode.leftChild.leftRotate())
						fallthrough
					case LL:
						newRootNode = lostBalanceNode.rightRotate()
					case RL:
						lostBalanceNode.setRightChild(lostBalanceNode.rightChild.rightRotate())
						fallthrough
					case RR:
						newRootNode = lostBalanceNode.leftRotate()
					default:
						fmt.Printf("Error: lost balance node %v rotate type wrong\n", lostBalanceNode)
						lostBalanceNode.preOrderTraversal(func(h *HashValue) bool {
							fmt.Printf("DEBUG: range key: %v, value: %v\n", h.k, h.v)
							return true
						}, 0)
						panic(fmt.Sprintf("Error: lost balance node %v rotate type wrong\n", lostBalanceNode))
					}

					if lostBalanceNodeParent == nil {
						d.buckets[hashIndex] = newRootNode
						d.buckets[hashIndex].parentNode = nil
					} else {
						if lostBalanceNodeParent.leftChild == lostBalanceNode {
							lostBalanceNodeParent.setLeftChild(newRootNode)
						} else if lostBalanceNodeParent.rightChild == lostBalanceNode {
							lostBalanceNodeParent.setRightChild(newRootNode)
						} else {
							fmt.Printf("Error: lost balance node %v is not exists in its parent %v child\n", lostBalanceNode.value.k, parentNode.value.k)
							panic(fmt.Sprintf("Error: lost balance node %v is not exists in its parent %v child\n", lostBalanceNode.value.k, parentNode.value.k))
						}
					}
				}
				return value, true
			}
		}
	}
}

func (d *avltHashMapData) Range(op func(*HashValue) bool) {
	for _, bucket := range d.buckets {
		if bucket != nil {
			// fmt.Println()
			// fmt.Printf("inOrderTraversal\n")
			// if !bucket.inOrderTraversal(op) {
			// 	return
			// }
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

// 2-3 tree - TTT

type tttNode struct {
	leftValue, middleValue, rightValue            *HashValue // middleValue 辅助数据
	leftChild, middleChild, rightChild            *tttNode   // 子树
	parentNode, middleLeftChild, middleRightChild *tttNode   // 辅助树
}

func (n *tttNode) preOrderTraversal(op func(*HashValue) bool, deep int) bool {
	if n.leftChild != nil && !n.leftChild.preOrderTraversal(op, deep+1) {
		return false
	}
	if n.leftValue != nil {
		fmt.Printf("%v", strings.Repeat("\t", deep))
		if !op(n.leftValue) {
			return false
		}
	}
	if n.middleChild != nil && !n.middleChild.preOrderTraversal(op, deep+1) {
		return false
	}
	if n.rightValue != nil {
		fmt.Printf("%v", strings.Repeat("\t", deep))
		if !op(n.rightValue) {
			return false
		}
	}
	if n.rightChild != nil && !n.rightChild.preOrderTraversal(op, deep+1) {
		return false
	}
	return true
}

func (n *tttNode) getKeyString() string {
	var keyString string
	if n.leftValue != nil {
		keyString = fmt.Sprintf("%v", n.leftValue.k)
	}
	if n.rightValue != nil {
		if len(keyString) == 0 {
			keyString = fmt.Sprintf("%v", n.rightValue.k)
		} else {
			keyString = fmt.Sprintf("%v,%v", keyString, n.rightValue.k)
		}
	}
	return keyString
}

func (n *tttNode) validateCheck() {
	if n.leftValue == nil && n.rightValue != nil {
		panic(fmt.Sprintf("node %+v left value is nil but right value is not nil\n", n))
	}
	if n.leftValue == nil && (n.leftChild != nil || n.middleChild != nil) {
		panic(fmt.Sprintf("node %+v left value is nil but left or middle child is not nil\n", n))
	}
	if n.rightValue == nil && n.rightChild != nil {
		panic(fmt.Sprintf("node %+v right value is nil but right child is not nil\n", n))
	}
	if n.leftChild != nil {
		n.leftChild.validateCheck()
	}
	if n.rightChild != nil {
		n.rightChild.validateCheck()
	}
}

func (n *tttNode) resetNodeValue(leftValue, middleValue, rightValue *HashValue) {
	n.leftValue, n.middleValue, n.rightValue = leftValue, middleValue, rightValue
}

func (n *tttNode) splitNode() *tttNode {
	newLeftChild := &tttNode{
		leftValue: n.leftValue,
	}
	if n.leftChild != nil {
		newLeftChild.setLeftChild(n.leftChild)
	}
	if n.middleLeftChild != nil {
		newLeftChild.setMiddleChild(n.middleLeftChild)
	}

	newMiddleChild := &tttNode{
		leftValue: n.rightValue,
	}
	if n.middleRightChild != nil {
		newMiddleChild.setLeftChild(n.middleRightChild)
	}
	if n.rightChild != nil {
		newMiddleChild.setMiddleChild(n.rightChild)
	}

	newRootNode := &tttNode{
		leftValue: n.middleValue,
	}
	newRootNode.setLeftChild(newLeftChild)
	newRootNode.setMiddleChild(newMiddleChild)

	return newRootNode
}

func (n *tttNode) newSplitNodeType1(insertHashValue *HashValue, insertType InsertType) *tttNode {
	newLeftChild := &tttNode{}
	if n.leftChild != nil {
		newLeftChild.setLeftChild(n.leftChild)
	}

	newMiddleChild := &tttNode{}
	if n.rightChild != nil {
		newMiddleChild.setMiddleChild(n.rightChild)
	}

	newRootNode := &tttNode{}
	newRootNode.setLeftChild(newLeftChild)
	newRootNode.setMiddleChild(newMiddleChild)

	switch insertType {
	case InsertLeft:
		newLeftChild.leftValue = insertHashValue
		newRootNode.leftValue = n.leftValue
		newMiddleChild.leftValue = n.rightValue
	case InsertMiddle:
		newLeftChild.leftValue = n.leftValue
		newRootNode.leftValue = insertHashValue
		newMiddleChild.leftValue = n.rightValue
	case InsertRight:
		newLeftChild.leftValue = n.leftValue
		newRootNode.leftValue = n.rightValue
		newMiddleChild.leftValue = insertHashValue
	}

	return newRootNode
}

func (n *tttNode) newSplitNodeType2(originNode, newNode *tttNode) {
	if n.leftChild == originNode {
		//     D     C     C,D
		//    /| +  /| =  / | \
		// A,C E   A B   A  B  E
		n.setRightChild(n.middleChild)
		n.setMiddleChild(newNode.middleChild)
		n.setLeftChild(newNode.leftChild)
		n.resetNodeValue(newNode.leftValue, nil, n.leftValue)
	} else if n.middleChild == originNode {
		//   D       F     D,F
		//  /|   +  /| =  / | \
		// C E,G   E G   C  E  G
		n.setLeftChild(n.leftChild)
		n.setMiddleChild(newNode.leftChild)
		n.setRightChild(newNode.middleChild)
		n.resetNodeValue(n.leftValue, nil, newNode.leftValue)
	} else {
		panic("error node")
	}
}

func (n *tttNode) newSplitNodeType3(originNode, newNode *tttNode) *tttNode {
	newRootNode := &tttNode{}
	newLeftChild := &tttNode{}
	newMiddleChild := &tttNode{}
	switch {
	//                      E
	//                     /|
	//     E,J       C    C J
	//    / | \  +  /| = /| |\
	// A,D  H  L   A D  A D H L
	case n.leftChild == originNode:
		newRootNode.leftValue = n.leftValue
		newLeftChild = newNode
		newMiddleChild.leftValue = n.rightValue
		newMiddleChild.setLeftChild(n.middleChild)
		newMiddleChild.setMiddleChild(n.rightChild)
	//                        G
	//                       /|
	//     E,J       G      E J
	//    / | \  +  /| =   /| |\
	// A,D F,G L   F H  A,D F H L
	case n.middleChild == originNode:
		newRootNode.leftValue = newNode.leftValue
		newLeftChild.leftValue = n.leftValue
		newLeftChild.setLeftChild(n.leftChild)
		newLeftChild.setMiddleChild(newNode.leftChild)
		newMiddleChild.leftValue = n.rightValue
		newMiddleChild.setLeftChild(newNode.middleChild)
		newMiddleChild.setMiddleChild(n.rightChild)
	//                        J
	//                       /|
	//     E,J         L    C L
	//    / | \    +  /| = /| |\
	// A,D  H  K,L   K M  A D K M
	case n.rightChild == originNode:
		newRootNode.leftValue = n.rightValue
		newMiddleChild = newNode
		newLeftChild.leftValue = n.leftValue
		newLeftChild.setLeftChild(n.leftChild)
		newLeftChild.setMiddleChild(n.middleChild)
	default:
		panic("error node")
	}
	newRootNode.setLeftChild(newLeftChild)
	newRootNode.setMiddleChild(newMiddleChild)
	return newRootNode
}

func (n *tttNode) setLeftChild(child *tttNode) {
	n.leftChild = child
	if child != nil {
		child.parentNode = n
	}
}

func (n *tttNode) setMiddleChild(child *tttNode) {
	n.middleChild = child
	if child != nil {
		child.parentNode = n
	}
}

func (n *tttNode) setRightChild(child *tttNode) {
	n.rightChild = child
	if child != nil {
		child.parentNode = n
	}
}

func (n *tttNode) setMiddleLeftChild(child *tttNode) {
	n.middleLeftChild = child
	if child != nil {
		child.parentNode = n
	}
}

func (n *tttNode) setMiddleRightChild(child *tttNode) {
	n.middleRightChild = child
	if child != nil {
		child.parentNode = n
	}
}

type tttNodeType int

const (
	errorNode tttNodeType = iota
	twoChildren
	threeChildren
	fourChildren
)

func (n *tttNode) getNodeType() tttNodeType {
	if n.leftValue != nil {
		if n.rightValue != nil {
			return threeChildren
		} else {
			return twoChildren
		}
	}
	if n.rightValue != nil {
		// TODO: panic here
		return twoChildren
	}
	// TODO: panic here
	return errorNode
}

type tttHashMapData struct {
	buckets []*tttNode
}

func (d *tttHashMapData) Len() int {
	return len(d.buckets)
}

func (d *tttHashMapData) Get(hashIndex, key int) (int, bool) {
	if d.buckets[hashIndex] == nil {
		return 0, false
	} else {
		node := d.buckets[hashIndex]
		for {
			if node.leftValue != nil {
				switch {
				case node.leftValue.k == key:
					return node.leftValue.v, true
				case key < node.leftValue.k:
					if node.leftChild != nil {
						node = node.leftChild
						continue
					} else {
						return 0, false
					}
				}
			}
			if node.rightValue != nil {
				switch {
				case node.rightValue.k == key:
					return node.rightValue.v, true
				case node.rightValue.k < key:
					if node.rightChild != nil {
						node = node.rightChild
						continue
					} else {
						return 0, false
					}
				}
			}
			if node.middleChild != nil {
				node = node.middleChild
			} else {
				return 0, false
			}
		}
	}
}

func (d *tttHashMapData) set(hashIndex int, insertHashValue *HashValue) bool {
	if d.buckets[hashIndex] == nil {
		d.buckets[hashIndex] = &tttNode{
			leftValue: insertHashValue,
		}
	} else {
		node := d.buckets[hashIndex]
		hashValue := insertHashValue
		var leftChild, middleChild *tttNode
		fromSplit := false
	INSERT:
		for {
			// 左边界
			if node.leftValue != nil {
				switch {
				case node.leftValue.k == hashValue.k: // 左等
					node.leftValue.v = hashValue.v
					goto SPLIT
				case hashValue.k < node.leftValue.k: // 左左
					if !fromSplit && node.leftChild != nil {
						node = node.leftChild
						continue
					} else {
						if node.rightValue == nil {
							// insert 1
							// 2 -> 1,2
							node.resetNodeValue(hashValue, nil, node.leftValue)
							// split
							//      5        2,5
							// 1,2,3 6 -> 1,3   6
							if fromSplit {
								node.setRightChild(node.middleChild)
								node.setMiddleChild(middleChild)
								node.setLeftChild(leftChild)
							}
						} else {
							// insert 1
							// 2,3 -> 1,2,3
							node.resetNodeValue(hashValue, node.leftValue, node.rightValue)
							node.setLeftChild(leftChild)
							node.setMiddleLeftChild(middleChild)
							node.setMiddleRightChild(node.middleChild)
							node.setMiddleChild(nil)
						}
						goto SPLIT
					}
				}
			}
			// 右边界
			if node.rightValue != nil {
				switch {
				case node.rightValue.k == hashValue.k: // 右等
					node.rightValue.v = hashValue.v
					goto SPLIT
				case node.rightValue.k < hashValue.k: // 右右
					if !fromSplit && node.rightChild != nil {
						node = node.rightChild
						continue
					} else {
						if node.leftValue == nil {
							// TODO: panic here
							// insert 3
							// 2 -> 2,3
							node.resetNodeValue(node.rightValue, nil, hashValue)
						} else {
							// insert 3
							// 1,2 -> 1,2,3
							node.resetNodeValue(node.leftValue, node.rightValue, hashValue)
							node.setRightChild(middleChild)
							node.setMiddleLeftChild(node.middleChild)
							node.setMiddleRightChild(leftChild)
							node.setMiddleChild(nil)
						}
						goto SPLIT
					}
				}
			}
			// 中子树
			if node.middleChild != nil { // 左右/右左
				// 从中间分裂上来
				if fromSplit {
					if node.rightValue != nil {
						node.resetNodeValue(node.leftValue, hashValue, node.rightValue)
						node.setMiddleLeftChild(leftChild)
						node.setMiddleRightChild(middleChild)
						node.setMiddleChild(nil)
					} else {
						node.resetNodeValue(node.leftValue, nil, hashValue)
						node.setMiddleChild(leftChild)
						node.setRightChild(middleChild)
					}
					goto SPLIT
				} else {
					node = node.middleChild
					continue
				}
			} else {
				// insert 2
				// 1,3 -> 1,2,3
				if fromSplit {
					node.resetNodeValue(node.leftValue, nil, hashValue)
					node.setMiddleChild(leftChild)
					node.setRightChild(middleChild)
				} else {
					if node.rightValue != nil {
						node.resetNodeValue(node.leftValue, hashValue, node.rightValue)
					} else {
						node.resetNodeValue(node.leftValue, nil, hashValue)
					}
				}
				goto SPLIT
			}
		}
	SPLIT:
		parentNode := node.parentNode
		if node.middleValue != nil {
			newRootNode := node.splitNode()
			node.middleValue = nil
			if parentNode != nil {
				node = parentNode
				hashValue = newRootNode.leftValue
				leftChild = newRootNode.leftChild
				middleChild = newRootNode.middleChild
				fromSplit = true
				goto INSERT
			} else {
				d.buckets[hashIndex] = newRootNode
			}
		}
	}

	fmt.Println()
	fmt.Printf("after insert %v\n", insertHashValue.k)
	d.buckets[hashIndex].preOrderTraversal(func(h *HashValue) bool {
		fmt.Printf("DEBUG: range key: %v, value: %v\n", h.k, h.v)
		return true
	}, 0)

	return true
}

type InsertType int

const (
	InsertError InsertType = iota
	InsertLeft
	InsertRight
	InsertMiddle
)

func (d *tttHashMapData) Set(hashIndex int, insertHashValue *HashValue) bool {
	if d.buckets[hashIndex] == nil {
		d.buckets[hashIndex] = &tttNode{
			leftValue: insertHashValue,
		}
	} else {
		insertType := InsertError
		node := d.buckets[hashIndex]
		for {
			if node.leftValue != nil {
				switch {
				case node.leftValue.k == insertHashValue.k:
					node.leftValue.v = insertHashValue.v
					return true
				case insertHashValue.k < node.leftValue.k:
					if node.leftChild != nil {
						node = node.leftChild
						continue
					} else {
						insertType = InsertLeft
						goto SPLIT
					}
				}
			}
			if node.rightValue != nil {
				switch {
				case node.rightValue.k == insertHashValue.k:
					node.rightValue.v = insertHashValue.v
					return true
				case node.rightValue.k < insertHashValue.k:
					if node.rightChild != nil {
						node = node.rightChild
						continue
					} else {
						insertType = InsertRight
						goto SPLIT
					}
				}
			}
			if node.middleChild != nil {
				node = node.middleChild
			} else {
				if node.rightValue == nil {
					insertType = InsertRight
				} else if node.leftValue == nil {
					insertType = InsertLeft
					panic("error node")
				} else {
					insertType = InsertMiddle
				}
				goto SPLIT
			}
		}
	SPLIT:
		switch node.getNodeType() {
		case twoChildren:
			switch insertType {
			case InsertLeft:
				node.resetNodeValue(insertHashValue, nil, node.leftValue)
			case InsertRight:
				node.resetNodeValue(node.leftValue, nil, insertHashValue)
			default:
				panic("error insert type")
			}
		case threeChildren:
			if insertType == InsertError {
				panic("error insert type")
			}
			if node.parentNode == nil {
				d.buckets[hashIndex] = node.newSplitNodeType1(insertHashValue, insertType)
			} else {
				switch node.parentNode.getNodeType() {
				case twoChildren:
					newRootNode := node.newSplitNodeType1(insertHashValue, insertType)
					node.parentNode.newSplitNodeType2(node, newRootNode)
				case threeChildren:
					newRootNode := node.newSplitNodeType1(insertHashValue, insertType)
				RESPLIT:
					newRootNode = node.parentNode.newSplitNodeType3(node, newRootNode)
					node = node.parentNode
					if node.parentNode == nil {
						d.buckets[hashIndex] = newRootNode
					} else {
						goto RESPLIT
					}
				default:
					panic("error node type")
				}
			}
		default:
			panic("error node type")
		}
	}

	fmt.Println()
	fmt.Printf("after insert %v\n", insertHashValue.k)
	d.buckets[hashIndex].preOrderTraversal(func(h *HashValue) bool {
		fmt.Printf("DEBUG: range key: %v, value: %v\n", h.k, h.v)
		return true
	}, 0)

	return true
}

func (d *tttHashMapData) Del(hashIndex, key int) (int, bool) {
	if d.buckets[hashIndex] == nil {
		return 0, false
	} else {
		node := d.buckets[hashIndex]
		var value int
		var deleteNode *tttNode
		for {
			if node.leftValue != nil {
				switch {
				case node.leftValue.k == key:
					value = node.leftValue.v
					deleteNode = node
					node.resetNodeValue(node.rightValue, nil, nil)
					goto REBALANCE
				case key < node.leftValue.k:
					node = node.leftChild
					continue
				}
			}
			if node.rightValue != nil {
				switch {
				case node.rightValue.k == key:
					value = node.rightValue.v
					deleteNode = node
					node.resetNodeValue(node.leftValue, nil, nil)
					goto REBALANCE
				case node.rightValue.k < key:
					node = node.rightChild
					continue
				}
			}
			node = node.middleChild
		}
	REBALANCE:
		if deleteNode.leftValue != nil || deleteNode.rightValue != nil {
			return value, true
		}
		return value, true
	}
}

func (d *tttHashMapData) Range(op func(*HashValue) bool) {
	for _, bucket := range d.buckets {
		if bucket != nil {
			fmt.Println()
			fmt.Printf("preOrderTraversal\n")
			if !bucket.preOrderTraversal(op, 0) {
				return
			}
		}
	}
}

func (d *tttHashMapData) Reallocate(size uint) {
	if uint(len(d.buckets)) != size {
		d.buckets = make([]*tttNode, size)
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
	seed := time.Now().UnixNano()
	fmt.Printf("seed is %v\n", seed)
	rand.Seed(seed)
	for index := 0; index != 10000; index++ {
		fmt.Println()
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

		// hashMapTest(seed, index, keyValueMap, WithHashMapData(&avltHashMapData{
		// 	buckets: make([]*avltNode, DEFAULT_HASH_MAP_SIZE>>10),
		// }))

		hashMapTest(seed, index, keyValueMap, WithHashMapData(&tttHashMapData{
			buckets: make([]*tttNode, DEFAULT_HASH_MAP_SIZE>>10),
		}))

		// hashMapDebug(seed, index, debugKeyValueMap, WithHashMapData(&tttHashMapData{
		// 	buckets: make([]*tttNode, DEFAULT_HASH_MAP_SIZE>>10),
		// }))
	}
}

type debugData struct {
	seed      int64
	inex      int
	kvMap     map[int]int
	setSlice  []int
	delSlice  []int
	panicInfo strings.Builder
}

func (d debugData) outputFile() {
	outputFile, openError := os.Create(fmt.Sprintf("%v-%v.log", d.seed, d.inex))
	if openError != nil {
		fmt.Printf("Error: open file occurs error: %v\n", openError)
		return
	}
	outputFile.WriteString(fmt.Sprintf("key value map: %v\n", func() string {
		var mapStringBuilder strings.Builder
		for k, v := range d.kvMap {
			mapStringBuilder.WriteString(fmt.Sprintf("%v", k))
			mapStringBuilder.WriteString(":")
			mapStringBuilder.WriteString(fmt.Sprintf("%v", v))
			mapStringBuilder.WriteString(",")
		}
		return mapStringBuilder.String()
	}()))
	outputFile.WriteString(fmt.Sprintf("set slice: %v\n", func() string {
		var mapStringBuilder strings.Builder
		for _, k := range d.setSlice {
			mapStringBuilder.WriteString(fmt.Sprintf("%v,", k))
		}
		return mapStringBuilder.String()
	}()))
	outputFile.WriteString(fmt.Sprintf("del slice: %v\n", func() string {
		var mapStringBuilder strings.Builder
		for _, k := range d.delSlice {
			mapStringBuilder.WriteString(fmt.Sprintf("%v,", k))
		}
		return mapStringBuilder.String()
	}()))
	outputFile.WriteString(d.panicInfo.String())
	outputFile.Close()
}

var (
	debugKeyValueMap = map[int]int{
		1023: 6, 982: 7, 528: 0, 20: 1, 199: 2, 702: 3, 224: 4, 170: 5,
	}

	debugSetSlice = []int{
		170, 1023, 982, 528, 20, 199, 702, 224,
	}

	debugGetSlice = []int{}

	debugDelSlice = []int{}
)

func hashMapDebug(seed int64, index int, keyValueMap map[int]int, options ...HashMapOption) {
	debugHashMap := MakeHashMap(options...)

	debugData := debugData{
		seed:  seed,
		inex:  index,
		kvMap: keyValueMap,
	}

	fmt.Println()

	for _, key := range debugSetSlice {
		fmt.Printf("debugHashMap.Set(%v, %v)\n", key, keyValueMap[key])
		debugData.setSlice = append(debugData.setSlice, key)
		if ok := debugHashMap.Set(key, keyValueMap[key]); !ok {
			fmt.Printf("debugHashMap.Set(%v, %v) failed\n", key, keyValueMap[key])
		} else {
			// fmt.Printf("after debugHashMap.Set(%v, %v), debugHashMap load factor is %v\n", key, value, debugHashMap.GetLoadFactor(0))
			// fmt.Printf("debugHashMap.Set(%v, %v) at hash index %v success\n", key, value, hashIndex)
		}
		fmt.Println()
	}

	fmt.Printf("after debugHashMap.Set\n")
	debugHashMap.Range(func(k, v int) bool {
		fmt.Printf("range key: %v, value: %v\n", k, v)
		return true
	})

	for key, value := range keyValueMap {
		if _value, hasKey := debugHashMap.Get(key); !hasKey || _value != value {
			fmt.Printf("debugHashMap.Get(%v), not has key or store value %v not equal to origin value %v\n", key, _value, value)
		} else {
			// fmt.Printf("debugHashMap.Get(%v), key and store value equal to origin value %v\n", key, value)
		}
	}

	// for _, key := range debugDelSlice {
	// 	fmt.Printf("debugHashMap.Del(%v)\n", key)
	// 	if _value, hasKey := debugHashMap.Del(key); !hasKey || _value != keyValueMap[key] {
	// 		fmt.Printf("debugHashMap.Del(%v), not has key or store value %v not equal to origin value %v\n", key, _value, keyValueMap[key])
	// 	} else {
	// 		// fmt.Printf("after Del, debugHashMap load factor is %v\n", debugHashMap.GetLoadFactor(0))
	// 		// fmt.Printf("debugHashMap.Del(%v), key and store value equal to origin value %v\n", key, value)
	// 	}
	// }

	// debugHashMap.Range(func(k, v int) bool {
	// 	fmt.Printf("key: %v, value: %v\n", k, v)
	// 	return true
	// })
}

func hashMapTest(seed int64, index int, keyValueMap map[int]int, options ...HashMapOption) {
	testHashMap := MakeHashMap(options...)

	debugData := debugData{
		seed:  seed,
		inex:  index,
		kvMap: keyValueMap,
	}

	defer func() {
		if err := recover(); err != nil {
			debugData.panicInfo.WriteString(fmt.Sprintf("panic: %v", err))
			debugData.outputFile()
		}
	}()

	fmt.Println()

	for key, value := range keyValueMap {
		fmt.Printf("testHashMap.Set(%v, %v)\n", key, value)
		debugData.setSlice = append(debugData.setSlice, key)
		if ok := testHashMap.Set(key, value); !ok {
			fmt.Printf("testHashMap.Set(%v, %v) failed\n", key, value)
		} else {
			// fmt.Printf("after testHashMap.Set(%v, %v), testHashMap load factor is %v\n", key, value, testHashMap.GetLoadFactor(0))
			// fmt.Printf("testHashMap.Set(%v, %v) at hash index %v success\n", key, value, hashIndex)
		}
		fmt.Println()
	}

	fmt.Printf("after testHashMap.Set\n")
	testHashMap.Range(func(k, v int) bool {
		fmt.Printf("range key: %v, value: %v\n", k, v)
		return true
	})

	for key, value := range keyValueMap {
		if _value, hasKey := testHashMap.Get(key); !hasKey || _value != value {
			fmt.Printf("testHashMap.Get(%v), not has key or store value %v not equal to origin value %v\n", key, _value, value)
			panic(fmt.Sprintf("testHashMap.Get(%v), not has key or store value %v not equal to origin value %v\n", key, _value, value))
		} else {
			// fmt.Printf("testHashMap.Get(%v), key and store value equal to origin value %v\n", key, value)
		}
	}

	// for key, value := range keyValueMap {
	// 	fmt.Printf("testHashMap.Del(%v)\n", key)
	// 	debugData.delSlice = append(debugData.delSlice, key)
	// 	if _value, hasKey := testHashMap.Del(key); !hasKey || _value != value {
	// 		fmt.Printf("testHashMap.Del(%v), not has key or store value %v not equal to origin value %v\n", key, _value, value)
	// 		debugData.outputFile()
	// 		return
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
