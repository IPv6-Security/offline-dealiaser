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
package radix

import (
	"bufio"
	"container/list"
	"fmt"
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// Label contains the lookup label results
type Label struct {
	Aliased  bool   `json:"aliased"`
	Metadata string `json:"metadata,omitempty"`
}

type Node struct {
	isLeaf      bool
	startPrefix uint8
	endPrefix   uint8
	length      uint8
	value       net.IP
	children    []*Node
}

type Radix struct {
	root                      *Node
	isChanged                 bool
	constructionNewAliasFound bool
	checkpointBaseName        string
	checkpointFrequency       float32
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func dupIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

func (t *Radix) traverseBFSRadix() {
	nodeCounter := 0
	leafCounter := 0
	queue := list.New()
	queue.PushBack(t.root)
	log.Info("BFS:")
	for queue.Len() != 0 {
		front := queue.Front()
		if front.Value.(*Node).isLeaf {
			leafCounter++
		}
		for i := range front.Value.(*Node).children {
			queue.PushBack(front.Value.(*Node).children[i])
		}
		nodeCounter++
		// fmt.Println(front.Value.(*Node))
		queue.Remove(front)
	}
	log.Infof("Leaf Count: %d\n\n", leafCounter)
	log.Infof("Node Count: %d\n\n", nodeCounter)
}

func (t *Radix) createLabel() Label {
	return Label{
		Aliased:  false,
		Metadata: "",
	}
}

func (t *Radix) lookup(ip net.IP) Label {
	label := t.createLabel()
	current := t.root
	ipBytes := ip.To16()
	var i, j, k int
	for i = 0; i < 128; {
		matchIndex := -1
		matchCounter := 0
		for j = 0; j < len(current.children); j++ {
			matchCounter = 0
			if i != int(current.children[j].startPrefix) {
				log.Fatalln("start Prefix don't match with current bit index")
				continue
			}
			for k = i; k < int(current.children[j].endPrefix); k++ {
				fb := ipBytes[k/8] & (128 >> (k % 8))
				sb := current.children[j].value[k/8] & (128 >> (k % 8))
				if fb != sb {
					break
				}
				matchCounter++
				matchIndex = j
			}
			if matchIndex != -1 {
				if matchCounter != int(current.children[j].length) {
					// Partial matches does not indicate alias
					// Go back and check other kids.
					matchIndex = -1
					matchCounter = 0
					continue
				}
				break
			}
		}
		if matchIndex == -1 {
			// No match could be found with any of the children.
			label.Aliased = false
			label.Metadata = ""
			return label
		} else {
			if current.children[matchIndex].isLeaf {
				// Fully matched, label it as aliased and return the matched prefix
				label.Aliased = true
				label.Metadata = fmt.Sprintf("%s/%d", current.children[matchIndex].value, current.children[matchIndex].endPrefix)
				return label
			} else {
				// We can go deeper in the tree, look for further matches until we arrive at a leaf node.
				i = i + matchCounter
				current = current.children[matchIndex]
			}
		}
	}
	return label
}

func (n *Node) exportCheckpointDFS(buf *bufio.Writer) {
	if n.isLeaf {
		if _, err := buf.WriteString(fmt.Sprintf("%s/%d", n.value, n.endPrefix)); err != nil {
			log.Fatal(err)
		}
		if err := buf.WriteByte('\n'); err != nil {
			log.Fatal(err)
		}
		buf.Flush()
	} else {
		for i := 0; i < len(n.children); i++ {
			n.children[i].exportCheckpointDFS(buf)
		}
	}
}

func (t *Radix) ExportCheckpoint(checkpointTime time.Time) {
	var err error
	var checkpointFile *os.File
	if checkpointFile, err = os.Create(fmt.Sprintf("%s-%s", t.checkpointBaseName, checkpointTime.Format(time.RFC3339))); err != nil {
		log.Fatal(err)
	}
	buf := bufio.NewWriter(checkpointFile)
	t.root.exportCheckpointDFS(buf)
	t.setChange(false)
	t.constructionNewAliasFound = false
}

func (t *Radix) remove(n *Node) {
	if n.isLeaf {
		n.value = nil
	} else {
		for i := 0; i < len(n.children); i++ {
			t.remove(n.children[i])
			n.children[i] = nil
		}
		n.value = nil
		n.children = nil
	}
}

func (t *Radix) insert(ip *net.IPNet) {
	current := t.root
	ones, _ := ip.Mask.Size()
	if current.isLeaf {
		newNode := &Node{
			isLeaf:      true,
			startPrefix: 0,
			endPrefix:   uint8(ones),
			length:      uint8(ones),
			value:       ip.IP.To16(),
		}
		current.children = append(current.children, newNode)
		current.isLeaf = false
		newNode = nil
		t.isChanged = true
		return
	} else {
		ipBytes := ip.IP.To16()
		ipEndPrefix := uint8(ones)
		var i, j, k int
		for i = 0; i < ones; {
			matchIndex := -1
			matchCounter := 0
			for j = 0; j < len(current.children); j++ {
				matchCounter = 0
				if i != int(current.children[j].startPrefix) {
					log.Fatal("start Prefix don't match with current bit index")
				}
				endPrefix := current.children[j].endPrefix
				if endPrefix > ipEndPrefix {
					endPrefix = ipEndPrefix
				}
				if i == int(endPrefix) && current.children[j].isLeaf {
					// Either startPrefix == endPrefix (which means it is a leaf)
					// or ipEndPrefix equals to where we left, which is equal to saying that
					// we reach to the end already. So, we already have something here!
					break
				}
				for k = i; k < int(endPrefix); k++ {
					fb := ipBytes[k/8] & (128 >> (k % 8))
					sb := current.children[j].value[k/8] & (128 >> (k % 8))
					if fb != sb {
						break
					}
					matchCounter++
					matchIndex = j
				}
				if matchIndex != -1 {
					break
				}
			}
			if matchIndex != -1 {
				// We have some prefix match
				// 1. We didn't fully cover our prefix range:
				// 		1.1. If the matched node is a leaf, then split.
				// 		1.2. If not, check if we reached to the end of
				// 			 matched node. If so, just move on to its children. OW,
				// 			 perform a split operation.
				// 2. We fully covered our prefix range:
				// 		Just move on to the children, next iteration will end the loop.
				if i+matchCounter < int(ipEndPrefix) {
					// We did not fully match our prefix, but we have some common
					// bits. It should be a split case if the current node is a leaf node.
					// If it is not a leaf node, we should follow to its children.
					// Splitting:
					// 		1. Create a parent node which covers the common prefix
					// 		2. Add that to current parent node as a child
					// 		3. Create a new child for the existing node to cover its rest
					// 		4. Create a new child for the new node.
					// 		5. Add these as children to the new parent node
					// 		6. Delete the reference for the previous leaf node.
					if i+matchCounter == int(current.children[matchIndex].endPrefix) {
						if current.children[matchIndex].isLeaf {
							return
						} else {
							// It is not a leaf node, we have to find more in depth matches and check
							// the children if we reached its end prefix.
							current = current.children[matchIndex]
							i = i + matchCounter
						}
					} else if current.children[matchIndex].isLeaf {
						if current.children[matchIndex].endPrefix == ipEndPrefix {
							if ipEndPrefix-uint8(i+matchCounter) == 1 {
								j := i + matchCounter
								current.children[matchIndex].value[j/8] = current.children[matchIndex].value[j/8] & ^(128 >> (j % 8))
								newNode := &Node{
									isLeaf:      true,
									startPrefix: current.children[matchIndex].startPrefix,
									endPrefix:   uint8(i + matchCounter),
									length:      uint8(i+matchCounter) - current.children[matchIndex].startPrefix,
									value:       dupIP(current.children[matchIndex].value),
								}
								t.remove(current.children[matchIndex])
								current.children[matchIndex] = newNode
								newNode = nil
								t.isChanged = true
								t.constructionNewAliasFound = true
								return
							}
						}
						newNode := &Node{
							isLeaf:      false,
							startPrefix: uint8(i),
							endPrefix:   uint8(i + matchCounter),
							length:      uint8(matchCounter),
							value:       current.children[matchIndex].value,
						}
						current.children = append(current.children, newNode)
						newNode2 := &Node{
							isLeaf:      true,
							startPrefix: uint8(i + matchCounter),
							endPrefix:   current.children[matchIndex].endPrefix,
							length:      current.children[matchIndex].endPrefix - uint8(i+matchCounter),
							value:       current.children[matchIndex].value,
						}
						newNode.children = append(newNode.children, newNode2)
						newNode3 := &Node{
							isLeaf:      true,
							startPrefix: uint8(i + matchCounter),
							endPrefix:   ipEndPrefix,
							length:      ipEndPrefix - uint8(i+matchCounter),
							value:       ip.IP.To16(),
						}
						newNode.children = append(newNode.children, newNode3)
						current.children[matchIndex] = nil
						current.children = append(current.children[:matchIndex], current.children[matchIndex+1:]...)
						newNode = nil
						newNode2 = nil
						newNode3 = nil
						t.isChanged = true
						return
					} else {
						newNode := &Node{
							isLeaf:      false,
							startPrefix: uint8(i),
							endPrefix:   uint8(i + matchCounter),
							length:      uint8(matchCounter),
							value:       current.children[matchIndex].value,
						}
						newNode2 := &Node{
							isLeaf:      true,
							startPrefix: uint8(i + matchCounter),
							endPrefix:   ipEndPrefix,
							length:      ipEndPrefix - uint8(i+matchCounter),
							value:       ip.IP.To16(),
						}
						newNode.children = append(newNode.children, newNode2)

						current.children[matchIndex].startPrefix = uint8(i + matchCounter)
						current.children[matchIndex].length = current.children[matchIndex].endPrefix - uint8(i+matchCounter)
						newNode.children = append(newNode.children, current.children[matchIndex])
						current.children[matchIndex] = newNode

						newNode = nil
						newNode2 = nil
						t.isChanged = true
						return
					}
				} else {
					// Matches all of the expected prefix, so they are identical and there is
					// already a node for that. No need to insert a new one (if their length is the same).
					// Else, prune (matched nodes's endPrefix > ipEndPrefix).
					// Prune Algorithm:
					// 		1. Create a new node (prefix: /matchedStartPrefix - inputEndPrefix).
					// 		Using input's end prefix is essential since we consumed it. It should be less
					// 		than matched end prefix. Use matched node's start prefix as we will replace
					// 		it with the new node.
					// 		2. Remove the node. Apply DFS for that. DFS is required for garbage collection.
					// 		Set all sub-tree nodes to nil (their value and children are slices).
					// 		3. Replace the pointer of matched node to the new node. Now the matched node is
					// 		lost, and cannot be referenced with any other material.
					// Also if the matched node is not a leaf:
					// In this case, their end prefix matches, but the matched node is not a leaf. Which means that, our
					// input has higher alias prefix compared the matched node (it should go deeper).
					if current.children[matchIndex].endPrefix != ipEndPrefix || !current.children[matchIndex].isLeaf {
						newNode := &Node{
							isLeaf:      true,
							startPrefix: current.children[matchIndex].startPrefix,
							endPrefix:   ipEndPrefix,
							length:      ipEndPrefix - current.children[matchIndex].startPrefix,
							value:       dupIP(ip.IP),
						}
						t.remove(current.children[matchIndex])
						current.children[matchIndex] = newNode
						newNode = nil
						t.isChanged = true
						t.constructionNewAliasFound = true
					}
					return
				}
			} else {
				// No match at all, just create a new child.
				newNode := &Node{
					isLeaf:      true,
					startPrefix: uint8(i),
					endPrefix:   uint8(ones),
					length:      uint8(ones - i),
					value:       ip.IP.To16(),
				}
				current.children = append(current.children, newNode)
				current.isLeaf = false
				newNode = nil
				t.isChanged = true
				return
			}
		}
	}
}

func (t *Radix) setCheckpointFrequency(checkpointFrequency float32) {
	t.checkpointFrequency = checkpointFrequency
}

func (t *Radix) setCheckpointBaseName(checkpointBaseName string) {
	t.checkpointBaseName = checkpointBaseName
}

func (t *Radix) setChange(val bool) {
	t.isChanged = val
}

func (n *Node) String() string {
	return fmt.Sprintf("{Mem-Location: %p, Value: %s/%d-%d, Length: %d, isLeaf: %t, Num-Children: %d}",
		n, n.value, n.startPrefix, n.endPrefix, n.length, n.isLeaf, len(n.children))
}

func InitRadix() *Radix {
	return &Radix{
		root: &Node{
			isLeaf:      true,
			startPrefix: 0,
			endPrefix:   0,
		},
		isChanged:                 false,
		constructionNewAliasFound: false,
		checkpointBaseName:        "checkpoint",
		checkpointFrequency:       1.0,
	}
}

func (t *Radix) SetCheckpointFrequency(checkpointFrequency float32) {
	t.setCheckpointFrequency(checkpointFrequency)
}

func (t *Radix) SetCheckpointBaseName(checkpointBaseName string) {
	t.setCheckpointBaseName(checkpointBaseName)
}

func (t *Radix) SetChange(val bool) {
	t.setChange(val)
}

func (t *Radix) GetCheckpointFrequency() float32 {
	return t.checkpointFrequency
}

func (t *Radix) CheckConstructionNewAliasFound() bool {
	return t.constructionNewAliasFound
}

func (t *Radix) IsChanged() bool {
	return t.isChanged
}

func (t *Radix) TraverseBFSRadix() {
	t.traverseBFSRadix()
}

func (t *Radix) Remove(n *Node) {
	t.remove(n)
}

func (t *Radix) LookUp(ip net.IP) Label {
	return t.lookup(ip)
}

func (t *Radix) Insert(ip *net.IPNet) {
	t.insert(ip)
}

func Tester(prefixFile, testInput string, stepSize int) {
	fin, err := os.Open(prefixFile)
	check(err)
	defer fin.Close()

	scanner := bufio.NewScanner(fin)
	scanner.Split(bufio.ScanLines)

	counter := 0
	radix := InitRadix()

	for scanner.Scan() {
		parsedIP, parsedNetwork, err := net.ParseCIDR(scanner.Text())
		check(err)
		if parsedIP == nil || parsedNetwork == nil {
			fmt.Println(scanner.Text())
			fmt.Println(counter)
			continue
		}
		radix.Insert(parsedNetwork)
		counter = counter + 1
		if counter%stepSize == 0 {
			fmt.Printf("Progress: %d\n", counter)
		}
	}
	fmt.Printf("Progress: %d\n", counter)

	fin, err = os.Open(testInput)
	check(err)
	defer fin.Close()

	scanner = bufio.NewScanner(fin)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		parsedIP := net.ParseIP(scanner.Text())
		if parsedIP == nil {
			continue
		}
		fmt.Printf("Lookup for: %s -> %v\n", parsedIP, radix.LookUp(parsedIP))
	}
}
