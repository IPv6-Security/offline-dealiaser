/*
Copyright 2024 Georgia Institute of Technology

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
/*
Uses 16-bit long bitmaps for 4 bit values in each node.
*/

package amt

import (
	"container/list"
	"fmt"
	"log"
	"math/bits"
	"os"
	"runtime"
	"unsafe"
)

type Node struct {
	isLeaf    bool
	isAliased bool
	value     byte
	prefix    uint8
	bitmap    uint16
	children  []*Node
}

type AMT struct {
	root              *Node
	currentCheckpoint int
	counter           []map[uint8]map[string]float64
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func (n *Node) String() string {
	return fmt.Sprintf("{Mem-Location: %p, Value: 0x%X, Prefix: %d, Bitmap: %016b, isLeaf: %t, isAliased: %t, Num-Children: %d}",
		n, n.value, n.prefix, n.bitmap, n.isLeaf, n.isAliased, len(n.children))
}

// nibbleOffset uses big-endian (HIGH_NIBBLE is the first 4 bytes)
func getNibble(value byte, nibbleOffset uint8) byte {
	nibble := ((value >> nibbleOffset) & 0x0F)
	return nibble
}

func (n *Node) getIndex(nibble byte) int {
	bitPosition := 1 << nibble
	if (n.bitmap & uint16(bitPosition)) == 0 {
		return -1
	}
	return int(bits.OnesCount16(n.bitmap & (uint16(bitPosition - 1))))
}

// func (n *Node) insertChildAt(index int, child *Node) {
// 	if len(n.children) == index {
// 		n.children = append(n.children, child)
// 	} else {
// 		// n.children[index] = nil
// 		n.children = append(n.children[:index+1], n.children[index:]...)
// 		n.children[index] = child
// 	}
// 	child = nil
// }

func (t *AMT) getPath(data []byte) []*Node {
	path := make([]*Node, 0)
	dataLength := len(data)
	current := t.root
	for i := 0; i < dataLength*2; i++ {
		nibbleOffset := uint8(4 - (4 * (i % 2)))
		nibble := getNibble(data[i/2], nibbleOffset)
		nibbleIndex := current.getIndex(nibble)
		if nibbleIndex == -1 {
			return []*Node{}
		}
		path = append(path, current)
		current = current.children[nibbleIndex]
	}
	path = append(path, current)
	current = nil
	return path
}

func (t *AMT) insert(data []byte) {
	dataLength := len(data)
	current := t.root
	for i := 0; i < dataLength*2; i++ {
		nibbleOffset := uint8(4 - (4 * (i % 2)))
		nibble := getNibble(data[i/2], nibbleOffset)
		nibbleIndex := current.getIndex(nibble)
		if nibbleIndex == -1 {
			current.bitmap = current.bitmap | (1 << nibble)
			nibbleIndex = current.getIndex(nibble)
			if len(current.children) == nibbleIndex {
				current.children = append(current.children, &Node{
					value:  nibble,
					prefix: current.prefix + 4,
				})
			} else {
				current.children = append(current.children[:nibbleIndex+1], current.children[nibbleIndex:]...)
				current.children[nibbleIndex] = &Node{
					value:  nibble,
					prefix: current.prefix + 4,
				}
			}
		}
		current = current.children[nibbleIndex]
	}
	current.isLeaf = true
	current.isAliased = true
	current = nil
}

func (t *AMT) insertWithCheckpoint(data []byte) {
	dataLength := len(data)
	current := t.root
	for i := 0; i < dataLength*2; i++ {
		nibbleOffset := uint8(4 - (4 * (i % 2)))
		nibble := getNibble(data[i/2], nibbleOffset)
		nibbleIndex := current.getIndex(nibble)
		if nibbleIndex == -1 {
			current.bitmap = current.bitmap | (1 << nibble)
			nibbleIndex = current.getIndex(nibble)
			if len(current.children) == nibbleIndex {
				current.children = append(current.children, &Node{
					value:  nibble,
					prefix: current.prefix + 4,
				})
			} else {
				current.children = append(current.children[:nibbleIndex+1], current.children[nibbleIndex:]...)
				current.children[nibbleIndex] = &Node{
					value:  nibble,
					prefix: current.prefix + 4,
				}
			}
			numChildren := len(current.children)
			// First update this prefix level as one more child inserted
			// 1. Update the average
			// 2. Check max and min again
			if _, ok := t.counter[t.currentCheckpoint][current.prefix]; ok {
				t.counter[t.currentCheckpoint][current.prefix]["totalChildren"]++
				t.counter[t.currentCheckpoint][current.prefix]["avg"] = t.counter[t.currentCheckpoint][current.prefix]["totalChildren"] / t.counter[t.currentCheckpoint][current.prefix]["totalNodes"]
				if t.counter[t.currentCheckpoint][current.prefix]["totalNodes"] == 1.0 {
					t.counter[t.currentCheckpoint][current.prefix]["max"] = float64(numChildren)
					t.counter[t.currentCheckpoint][current.prefix]["maxValue"] = float64(current.value)
					t.counter[t.currentCheckpoint][current.prefix]["min"] = float64(numChildren)
					t.counter[t.currentCheckpoint][current.prefix]["minValue"] = float64(current.value)
				} else {
					if t.counter[t.currentCheckpoint][current.prefix]["max"] < float64(numChildren) {
						t.counter[t.currentCheckpoint][current.prefix]["max"] = float64(numChildren)
						t.counter[t.currentCheckpoint][current.prefix]["maxValue"] = float64(current.value)
					}
					if t.counter[t.currentCheckpoint][current.prefix]["min"] > float64(numChildren) {
						t.counter[t.currentCheckpoint][current.prefix]["min"] = float64(numChildren)
						t.counter[t.currentCheckpoint][current.prefix]["minValue"] = float64(current.value)
					}
				}
			} else {
				// This is the first entry in this prefix level.
				t.counter[t.currentCheckpoint][current.prefix] = map[string]float64{
					"avg":           float64(numChildren),
					"max":           float64(numChildren),
					"min":           float64(numChildren),
					"minValue":      float64(current.value),
					"maxValue":      float64(current.value),
					"totalNodes":    1.0,
					"totalChildren": float64(numChildren),
				}
			}
			// Second, update the next prefix level since the child is created under current.prefix + 4 level.
			// 1. Check if that prefix level already has an entry or not.
			// 2. If so, just update the fields (check min, max again)
			// 3. If not, create a new dictionary entry.
			if _, ok := t.counter[t.currentCheckpoint][current.prefix+4]; ok {
				t.counter[t.currentCheckpoint][current.prefix+4]["totalNodes"]++
				t.counter[t.currentCheckpoint][current.prefix+4]["avg"] = t.counter[t.currentCheckpoint][current.prefix+4]["totalChildren"] / t.counter[t.currentCheckpoint][current.prefix+4]["totalNodes"]

				if t.counter[t.currentCheckpoint][current.prefix+4]["max"] < 0.0 {
					t.counter[t.currentCheckpoint][current.prefix+4]["max"] = 0.0
					t.counter[t.currentCheckpoint][current.prefix+4]["maxValue"] = float64(nibble)
				}
				if t.counter[t.currentCheckpoint][current.prefix+4]["min"] > 0.0 {
					t.counter[t.currentCheckpoint][current.prefix+4]["min"] = 0.0
					t.counter[t.currentCheckpoint][current.prefix+4]["minValue"] = float64(nibble)
				}
			} else {
				t.counter[t.currentCheckpoint][current.prefix+4] = map[string]float64{
					"avg":           0.0,
					"max":           0.0,
					"min":           0.0,
					"minValue":      float64(nibble),
					"maxValue":      float64(nibble),
					"totalNodes":    1.0,
					"totalChildren": 0.0,
				}
			}
		}
		current = current.children[nibbleIndex]
	}
	current.isLeaf = true
	current.isAliased = true
	current = nil
}

func (t *AMT) AddCheckpoint() {
	totalCheckpoints := len(t.counter)
	t.counter = append(t.counter, map[uint8]map[string]float64{
		0: {
			"avg":           0.0,
			"max":           -1.0,
			"min":           17.0,
			"minValue":      -1.0,
			"maxValue":      -1.0,
			"totalNodes":    0.0,
			"totalChildren": 0.0,
		},
	})
	for k, v := range t.counter[totalCheckpoints-1] {
		t.counter[totalCheckpoints][k] = map[string]float64{
			"avg":           0.0,
			"max":           -1.0,
			"min":           17.0,
			"minValue":      -1.0,
			"maxValue":      -1.0,
			"totalNodes":    0.0,
			"totalChildren": 0.0,
		}
		for k2, v2 := range v {
			t.counter[totalCheckpoints][k][k2] = v2
		}
	}
	t.currentCheckpoint++
}

func (t *AMT) ExportStats(exportStats string) {
	fout, err := os.OpenFile(exportStats, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	check(err)
	defer fout.Close()

	currentStats := t.counter[t.currentCheckpoint]
	totalNodes := 0
	for k, v := range currentStats {
		fout.WriteString(fmt.Sprintf("Prefix=%d\tNumNodes=%f\tAvgChildren=%f\tMinChildren=%f\tMinValue=%04b\tMaxChildren=%f\tMaxValue=%04b\n",
			k,
			v["totalNodes"],
			v["avg"],
			v["min"],
			int(v["minValue"]),
			v["max"],
			int(v["maxValue"])))
		totalNodes += int(v["totalNodes"])
	}
	fout.WriteString(fmt.Sprintf("TotalNodes=%d", totalNodes))
}

func (t *AMT) find(data []byte) bool {
	dataLength := len(data)
	current := t.root
	for i := 0; i < dataLength*2; i++ {
		nibbleOffset := uint8(4 - (4 * (i % 2)))
		nibble := getNibble(data[i/2], nibbleOffset)
		nibbleIndex := current.getIndex(nibble)
		if nibbleIndex == -1 {
			return false
		}
		current = current.children[nibbleIndex]
	}
	return current.isLeaf
}

func (n *Node) removeChild(value byte) int {
	index := n.getIndex(value)
	if index != -1 {
		n.children[index] = nil                                          // Set the child to nil pointer
		n.children = append(n.children[:index], n.children[index+1:]...) // Rearrange the children
		bitPosition := 1 << value                                        // Set the bit for the child in bitmap
		n.bitmap = n.bitmap & (^uint16(bitPosition))                     // Clear the bit for the child
		if len(n.children) == 0 {
			n.children = nil
			return 0
		}
		return 1
	}
	return -1
}

func (t *AMT) removeEntry(data []byte) int {
	deletedNodeCounter := 0
	path := t.getPath(data)
	if len(path) > 0 {
		i := len(path) - 1
		prevValue := path[i].value
		for ; i >= 0; i-- {
			node := path[i]
			if node.isLeaf {
				prevValue = node.value
				node = nil
				path[i] = nil
				deletedNodeCounter++
			} else {
				stats := node.removeChild(prevValue)
				if stats == -1 || stats == 1 { // Value not found in node's children or node has more children left after deletion
					break
				} else if stats == 0 && (node != t.root) { // Node does not have any more children left after deleting the previous node (child)
					prevValue = node.value
					node = nil
					path[i] = nil
					deletedNodeCounter++
				}
			}
		}
	}
	return deletedNodeCounter
}

func (n *Node) pruneChildren() int {
	if n.isLeaf {
		return 0
	}
	numDeletedNodes := 0
	numChildren := len(n.children)
	for i := range n.children {
		numDeletedNodes = numDeletedNodes + n.children[i].pruneChildren()
		n.children[i] = nil
		numDeletedNodes++
	}
	n.bitmap = 0
	n.children = nil
	n.isLeaf = true
	if numChildren > numDeletedNodes { // Something should be wrong in this case
		log.Panicln("panic: deleted nodes should be at least equal to number of children")
	}
	return numDeletedNodes
}

func (t *AMT) traverseBFSRadix(checkpoint int) []int {
	nodeCounter := 0
	normalNodeCounter := 0
	queue := list.New()
	queue.PushBack(t.root)
	for queue.Len() != 0 {
		front := queue.Front()
		for i := range front.Value.(*Node).children {
			queue.PushBack(front.Value.(*Node).children[i])
		}
		nodeCounter++
		normalNodeCounter++
		bitmap := front.Value.(*Node).bitmap
		numChildren := bits.OnesCount16(bitmap)
		if numChildren == 1 && front.Value.(*Node) != t.root {
			nodeCounter--
		}
		queue.Remove(front)
	}
	fmt.Printf("Checkpoint:%d - TotalNodes:%d - RadixNodes:%d - Diff:%d - Ratio: %f\n", checkpoint, normalNodeCounter, nodeCounter, normalNodeCounter-nodeCounter, float64(nodeCounter)/float64(normalNodeCounter))
	return []int{normalNodeCounter, nodeCounter, normalNodeCounter - nodeCounter}
}

func (t *AMT) traverseBFS(exportStats string) {
	fout, err := os.OpenFile(exportStats, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	check(err)
	defer fout.Close()

	nodeCounter := 0
	statsTable := map[uint8]map[string]float64{}
	queue := list.New()
	queue.PushBack(t.root)
	lastPrefix := t.root.prefix
	statsTable[lastPrefix] = map[string]float64{
		"avg":           0.0,
		"max":           -1.0,
		"min":           17.0,
		"minValue":      -1.0,
		"maxValue":      -1.0,
		"totalNodes":    0.0,
		"totalChildren": 0.0,
	}
	for queue.Len() != 0 {
		front := queue.Front()
		for i := range front.Value.(*Node).children {
			queue.PushBack(front.Value.(*Node).children[i])
		}
		nodeCounter++
		prefix := front.Value.(*Node).prefix
		value := front.Value.(*Node).value
		bitmap := front.Value.(*Node).bitmap
		numChildren := bits.OnesCount16(bitmap)
		if lastPrefix == prefix {
			statsTable[prefix]["totalNodes"] = statsTable[prefix]["totalNodes"] + 1
			statsTable[prefix]["totalChildren"] = statsTable[prefix]["totalChildren"] + float64(numChildren)
			if float64(numChildren) < statsTable[prefix]["min"] {
				statsTable[prefix]["min"] = float64(numChildren)
				statsTable[prefix]["minValue"] = float64(value)
			}
			if float64(numChildren) > statsTable[prefix]["max"] {
				statsTable[prefix]["max"] = float64(numChildren)
				statsTable[prefix]["maxValue"] = float64(value)
			}
		} else {
			statsTable[lastPrefix]["avg"] = statsTable[lastPrefix]["totalChildren"] / statsTable[lastPrefix]["totalNodes"]
			fout.WriteString(fmt.Sprintf("Summary\tPrefix=%d\tNumNodes=%f\tAvgChildren=%f\tMinChildren=%f\tMinValue=%04b\tMaxChildren=%f\tMaxValue=%04b\n",
				lastPrefix,
				statsTable[lastPrefix]["totalNodes"],
				statsTable[lastPrefix]["avg"],
				statsTable[lastPrefix]["min"],
				int(statsTable[lastPrefix]["minValue"]),
				statsTable[lastPrefix]["max"],
				int(statsTable[lastPrefix]["maxValue"])))
			lastPrefix = prefix
			statsTable[lastPrefix] = map[string]float64{
				"avg":           float64(numChildren),
				"min":           float64(numChildren),
				"max":           float64(numChildren),
				"minValue":      float64(value),
				"maxValue":      float64(value),
				"totalNodes":    1.0,
				"totalChildren": float64(numChildren),
			}
		}
		// fout.WriteString(fmt.Sprintf("Prefix=%d\tValue=%04b\tBitmap=%016b\tNumChildren=%d\n", prefix, value, bitmap, numChildren))
		queue.Remove(front)
	}
	fout.WriteString(fmt.Sprintf("Summary\tPrefix=%d\tNumNodes=%f\tAvgChildren=%f\tMinChildren=%f\tMinValue=%04b\tMaxChildren=%f\tMaxValue=%04b\n",
		lastPrefix,
		statsTable[lastPrefix]["totalNodes"],
		statsTable[lastPrefix]["avg"],
		statsTable[lastPrefix]["min"],
		int(statsTable[lastPrefix]["minValue"]),
		statsTable[lastPrefix]["max"],
		int(statsTable[lastPrefix]["maxValue"])))
	fout.WriteString(fmt.Sprintf("TotalNodes:%d\n", nodeCounter))
}

func InitAMT() *AMT {
	return &AMT{
		root: &Node{},
		counter: []map[uint8]map[string]float64{
			{
				0: {
					"avg":           0.0,
					"max":           -1.0,
					"min":           17.0,
					"minValue":      -1.0,
					"maxValue":      -1.0,
					"totalNodes":    1.0,
					"totalChildren": 0.0,
				},
			},
		},
		currentCheckpoint: 0,
	}
}

func (t *AMT) GetPath(data []byte) []*Node {
	return t.getPath(data)
}

func (t *AMT) Insert(data []byte) {
	t.insert(data)
}

func (t *AMT) InsertWithCheckpoint(data []byte) {
	t.insertWithCheckpoint(data)
}

func (t *AMT) Find(data []byte) bool {
	return t.find(data)
}

// Returns the number of deleted nodes.
func (t *AMT) RemoveEntry(data []byte) int {
	return t.removeEntry(data)
}

// Returns number of pruned children nodes under the sub-trie.
func (n *Node) PruneChildren() int {
	return n.pruneChildren()
}

func (t *AMT) TraverseBFS(exportStats string) {
	t.traverseBFS(exportStats)
}

func (t *AMT) TraverseBFSRadix(checkpoint int) []int {
	return t.traverseBFSRadix(checkpoint)
}

func MemCheckForEachElement() {
	var mem1, mem2 runtime.MemStats
	fmt.Println("memory init...")
	runtime.ReadMemStats(&mem1)
	log.Println(mem1.Alloc)
	log.Println(mem1.TotalAlloc)
	log.Println(mem1.HeapAlloc)
	log.Println(mem1.HeapSys)
	n1 := Node{}
	n2 := Node{}
	fmt.Println("memory comparison...")
	runtime.ReadMemStats(&mem2)
	log.Println(float64(mem2.Alloc - mem1.Alloc))
	log.Println(float64(mem2.TotalAlloc - mem1.TotalAlloc))
	log.Println(float64(mem2.HeapAlloc - mem1.HeapAlloc))
	log.Println(float64(mem2.HeapSys - mem1.HeapSys))

	fmt.Printf("n1 Location: %p\n", &n1)
	fmt.Printf("n2 Location: %p\n", &n2)

	fmt.Printf("n1 unsafe sizeof: %d\n", unsafe.Sizeof(n1))
	fmt.Printf("n2 unsafe sizeof: %d\n", unsafe.Sizeof(n2))

	fmt.Printf("n1.isLeaf Location: %p\n", &n1.isLeaf)
	fmt.Printf("n1.isAliased Location: %p\n", &n1.isAliased)
	fmt.Printf("n1.value Location: %p\n", &n1.value)
	fmt.Printf("n1.prefix Location: %p\n", &n1.prefix)
	fmt.Printf("n1.bitmap Location: %p\n", &n1.bitmap)
	fmt.Printf("n1.children Location: %p\n", &n1.children)
	fmt.Printf("n1.children unsafe sizeof: %d\n", unsafe.Sizeof(n1.children))
	fmt.Printf("n1.children Len: %d\n", len(n1.children))
	fmt.Printf("n1.children Cap: %d\n", cap(n1.children))
}
