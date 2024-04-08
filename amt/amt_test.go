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
package amt

import (
	"testing"
)

func identicalTries(n1, n2 *Node) bool {
	if n1 == nil && n2 == nil {
		return true
	}
	if n1 != nil && n2 != nil {
		if (n1.prefix == n2.prefix) &&
			(n1.bitmap == n2.bitmap) &&
			(n1.value == n2.value) &&
			(n1.isLeaf == n2.isLeaf) &&
			(n1.isAliased == n2.isAliased) &&
			(len(n1.children) == len(n2.children)) {
			for i := range n1.children {
				if !(identicalTries(n1.children[i], n2.children[i])) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func TestRemoveEntry(t *testing.T) {
	for _, tt := range []struct {
		expected []int
		nodes    [][]byte
		inputs   [][]byte
	}{
		{
			nodes: [][]byte{
				{0xFA, 0x45, 0x01},
				{0xFB, 0x45, 0x01},
				{0x30, 0x45, 0x01},
				{0x30, 0x45, 0x02}},
			inputs: [][]byte{
				{0xFA, 0x45, 0x01},
				{0xFA, 0x45, 0x01},
				{0x30, 0x45, 0x02},
				{0x30, 0x45, 0x03},
				{0x30, 0x46, 0x01},
				{0x30, 0x55, 0x01},
				{0x31, 0x45, 0x01},
				{0x40, 0x45, 0x01},
				{0x30, 0x45, 0x01},
				{0xFB, 0x45, 0x01},
				{0x38, 0x46, 0x01},
			},
			expected: []int{
				5, 0, 1, 0, 0, 0, 0, 0, 6, 6, 0,
			},
		},
	} {
		amt := InitAMT()
		for _, ip := range tt.nodes {
			amt.Insert(ip)
		}
		for i, ip := range tt.inputs {
			numDeletedNodes := amt.RemoveEntry(ip)
			if numDeletedNodes != tt.expected[i] {
				t.Errorf("Wrong number of deleted nodes for (0x%X): given: %d - expected: %d", ip, numDeletedNodes, tt.expected[i])
			}
		}
	}
}

func TestGetPath(t *testing.T) {
	for _, tt := range []struct {
		expected [][]*Node
		nodes    [][]byte
		inputs   [][]byte
	}{
		{
			nodes: [][]byte{
				{0xFA, 0x45, 0x01},
				{0xFB, 0x45, 0x01},
				{0x30, 0x45, 0x01},
				{0x30, 0x45, 0x02}},
			inputs: [][]byte{
				{0xFA, 0x45, 0x01},
				{0x30, 0x45, 0x02},
				{0x30, 0x45, 0x03},
				{0x30, 0x45, 0x11},
				{0x30, 0x46, 0x01},
				{0x30, 0x55, 0x01},
				{0x31, 0x45, 0x01},
				{0x40, 0x45, 0x01},
			},
			expected: [][]*Node{
				{
					{
						bitmap: 0b1000000000001000,
						children: []*Node{
							{
								prefix: 4,
								value:  0x03,
								bitmap: 0b0000000000000001,
								children: []*Node{
									{
										prefix: 8,
										value:  0x00,
										bitmap: 0b0000000000010000,
										children: []*Node{
											{
												prefix: 12,
												value:  0x04,
												bitmap: 0b0000000000100000,
												children: []*Node{
													{
														prefix: 16,
														value:  0x05,
														bitmap: 0b0000000000000001,
														children: []*Node{
															{
																prefix: 20,
																value:  0x00,
																bitmap: 0b0000000000000110,
																children: []*Node{
																	{
																		prefix: 24,
																		value:  0x01,
																		bitmap: 0b0000000000000000,
																		isLeaf: true,
																	},
																	{
																		prefix: 24,
																		value:  0x02,
																		bitmap: 0b0000000000000000,
																		isLeaf: true,
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							{
								prefix: 4,
								value:  0x0F,
								bitmap: 0b0000110000000000,
								children: []*Node{
									{
										prefix: 8,
										value:  0x0A,
										bitmap: 0b0000000000010000,
										children: []*Node{
											{
												prefix: 12,
												value:  0x04,
												bitmap: 0b0000000000100000,
												children: []*Node{
													{
														prefix: 16,
														value:  0x05,
														bitmap: 0b0000000000000001,
														children: []*Node{
															{
																prefix: 20,
																value:  0x00,
																bitmap: 0b0000000000000010,
																children: []*Node{
																	{
																		prefix: 24,
																		value:  0x01,
																		bitmap: 0b0000000000000000,
																		isLeaf: true,
																	},
																},
															},
														},
													},
												},
											},
										},
									},
									{
										prefix: 8,
										value:  0x0B,
										bitmap: 0b0000000000010000,
										children: []*Node{
											{
												prefix: 12,
												value:  0x04,
												bitmap: 0b0000000000100000,
												children: []*Node{
													{
														prefix: 16,
														value:  0x05,
														bitmap: 0b0000000000000001,
														children: []*Node{
															{
																prefix: 20,
																value:  0x00,
																bitmap: 0b0000000000000010,
																children: []*Node{
																	{
																		prefix: 24,
																		value:  0x01,
																		bitmap: 0b0000000000000000,
																		isLeaf: true,
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						prefix: 4,
						value:  0x0F,
						bitmap: 0b0000110000000000,
						children: []*Node{
							{
								prefix: 8,
								value:  0x0A,
								bitmap: 0b0000000000010000,
								children: []*Node{
									{
										prefix: 12,
										value:  0x04,
										bitmap: 0b0000000000100000,
										children: []*Node{
											{
												prefix: 16,
												value:  0x05,
												bitmap: 0b0000000000000001,
												children: []*Node{
													{
														prefix: 20,
														value:  0x00,
														bitmap: 0b0000000000000010,
														children: []*Node{
															{
																prefix: 24,
																value:  0x01,
																bitmap: 0b0000000000000000,
																isLeaf: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
							{
								prefix: 8,
								value:  0x0B,
								bitmap: 0b0000000000010000,
								children: []*Node{
									{
										prefix: 12,
										value:  0x04,
										bitmap: 0b0000000000100000,
										children: []*Node{
											{
												prefix: 16,
												value:  0x05,
												bitmap: 0b0000000000000001,
												children: []*Node{
													{
														prefix: 20,
														value:  0x00,
														bitmap: 0b0000000000000010,
														children: []*Node{
															{
																prefix: 24,
																value:  0x01,
																bitmap: 0b0000000000000000,
																isLeaf: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						prefix: 8,
						value:  0x0A,
						bitmap: 0b0000000000010000,
						children: []*Node{
							{
								prefix: 12,
								value:  0x04,
								bitmap: 0b0000000000100000,
								children: []*Node{
									{
										prefix: 16,
										value:  0x05,
										bitmap: 0b0000000000000001,
										children: []*Node{
											{
												prefix: 20,
												value:  0x00,
												bitmap: 0b0000000000000010,
												children: []*Node{
													{
														prefix: 24,
														value:  0x01,
														bitmap: 0b0000000000000000,
														isLeaf: true,
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						prefix: 12,
						value:  0x04,
						bitmap: 0b0000000000100000,
						children: []*Node{
							{
								prefix: 16,
								value:  0x05,
								bitmap: 0b0000000000000001,
								children: []*Node{
									{
										prefix: 20,
										value:  0x00,
										bitmap: 0b0000000000000010,
										children: []*Node{
											{
												prefix: 24,
												value:  0x01,
												bitmap: 0b0000000000000000,
												isLeaf: true,
											},
										},
									},
								},
							},
						},
					},
					{
						prefix: 16,
						value:  0x05,
						bitmap: 0b0000000000000001,
						children: []*Node{
							{
								prefix: 20,
								value:  0x00,
								bitmap: 0b0000000000000010,
								children: []*Node{
									{
										prefix: 24,
										value:  0x01,
										bitmap: 0b0000000000000000,
										isLeaf: true,
									},
								},
							},
						},
					},
					{
						prefix: 20,
						value:  0x00,
						bitmap: 0b0000000000000010,
						children: []*Node{
							{
								prefix: 24,
								value:  0x01,
								bitmap: 0b0000000000000000,
								isLeaf: true,
							},
						},
					},
					{
						prefix: 24,
						value:  0x01,
						bitmap: 0b0000000000000000,
						isLeaf: true,
					},
				},
				{
					{
						bitmap: 0b1000000000001000,
						children: []*Node{
							{
								prefix: 4,
								value:  0x03,
								bitmap: 0b0000000000000001,
								children: []*Node{
									{
										prefix: 8,
										value:  0x00,
										bitmap: 0b0000000000010000,
										children: []*Node{
											{
												prefix: 12,
												value:  0x04,
												bitmap: 0b0000000000100000,
												children: []*Node{
													{
														prefix: 16,
														value:  0x05,
														bitmap: 0b0000000000000001,
														children: []*Node{
															{
																prefix: 20,
																value:  0x00,
																bitmap: 0b0000000000000110,
																children: []*Node{
																	{
																		prefix: 24,
																		value:  0x01,
																		bitmap: 0b0000000000000000,
																		isLeaf: true,
																	},
																	{
																		prefix: 24,
																		value:  0x02,
																		bitmap: 0b0000000000000000,
																		isLeaf: true,
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							{
								prefix: 4,
								value:  0x0F,
								bitmap: 0b0000110000000000,
								children: []*Node{
									{
										prefix: 8,
										value:  0x0A,
										bitmap: 0b0000000000010000,
										children: []*Node{
											{
												prefix: 12,
												value:  0x04,
												bitmap: 0b0000000000100000,
												children: []*Node{
													{
														prefix: 16,
														value:  0x05,
														bitmap: 0b0000000000000001,
														children: []*Node{
															{
																prefix: 20,
																value:  0x00,
																bitmap: 0b0000000000000010,
																children: []*Node{
																	{
																		prefix: 24,
																		value:  0x01,
																		bitmap: 0b0000000000000000,
																		isLeaf: true,
																	},
																},
															},
														},
													},
												},
											},
										},
									},
									{
										prefix: 8,
										value:  0x0B,
										bitmap: 0b0000000000010000,
										children: []*Node{
											{
												prefix: 12,
												value:  0x04,
												bitmap: 0b0000000000100000,
												children: []*Node{
													{
														prefix: 16,
														value:  0x05,
														bitmap: 0b0000000000000001,
														children: []*Node{
															{
																prefix: 20,
																value:  0x00,
																bitmap: 0b0000000000000010,
																children: []*Node{
																	{
																		prefix: 24,
																		value:  0x01,
																		bitmap: 0b0000000000000000,
																		isLeaf: true,
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						prefix: 4,
						value:  0x03,
						bitmap: 0b0000000000000001,
						children: []*Node{
							{
								prefix: 8,
								value:  0x00,
								bitmap: 0b0000000000010000,
								children: []*Node{
									{
										prefix: 12,
										value:  0x04,
										bitmap: 0b0000000000100000,
										children: []*Node{
											{
												prefix: 16,
												value:  0x05,
												bitmap: 0b0000000000000001,
												children: []*Node{
													{
														prefix: 20,
														value:  0x00,
														bitmap: 0b0000000000000110,
														children: []*Node{
															{
																prefix: 24,
																value:  0x01,
																bitmap: 0b0000000000000000,
																isLeaf: true,
															},
															{
																prefix: 24,
																value:  0x02,
																bitmap: 0b0000000000000000,
																isLeaf: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						prefix: 8,
						value:  0x00,
						bitmap: 0b0000000000010000,
						children: []*Node{
							{
								prefix: 12,
								value:  0x04,
								bitmap: 0b0000000000100000,
								children: []*Node{
									{
										prefix: 16,
										value:  0x05,
										bitmap: 0b0000000000000001,
										children: []*Node{
											{
												prefix: 20,
												value:  0x00,
												bitmap: 0b0000000000000110,
												children: []*Node{
													{
														prefix: 24,
														value:  0x01,
														bitmap: 0b0000000000000000,
														isLeaf: true,
													},
													{
														prefix: 24,
														value:  0x02,
														bitmap: 0b0000000000000000,
														isLeaf: true,
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						prefix: 12,
						value:  0x04,
						bitmap: 0b0000000000100000,
						children: []*Node{
							{
								prefix: 16,
								value:  0x05,
								bitmap: 0b0000000000000001,
								children: []*Node{
									{
										prefix: 20,
										value:  0x00,
										bitmap: 0b0000000000000110,
										children: []*Node{
											{
												prefix: 24,
												value:  0x01,
												bitmap: 0b0000000000000000,
												isLeaf: true,
											},
											{
												prefix: 24,
												value:  0x02,
												bitmap: 0b0000000000000000,
												isLeaf: true,
											},
										},
									},
								},
							},
						},
					},
					{
						prefix: 16,
						value:  0x05,
						bitmap: 0b0000000000000001,
						children: []*Node{
							{
								prefix: 20,
								value:  0x00,
								bitmap: 0b0000000000000110,
								children: []*Node{
									{
										prefix: 24,
										value:  0x01,
										bitmap: 0b0000000000000000,
										isLeaf: true,
									},
									{
										prefix: 24,
										value:  0x02,
										bitmap: 0b0000000000000000,
										isLeaf: true,
									},
								},
							},
						},
					},
					{
						prefix: 20,
						value:  0x00,
						bitmap: 0b0000000000000110,
						children: []*Node{
							{
								prefix: 24,
								value:  0x01,
								bitmap: 0b0000000000000000,
								isLeaf: true,
							},
							{
								prefix: 24,
								value:  0x02,
								bitmap: 0b0000000000000000,
								isLeaf: true,
							},
						},
					},
					{
						prefix: 24,
						value:  0x02,
						bitmap: 0b0000000000000000,
						isLeaf: true,
					},
				},
				{}, {}, {}, {}, {}, {},
			},
		},
	} {
		amt := InitAMT()
		for _, ip := range tt.nodes {
			amt.Insert(ip)
		}
		for i, ip := range tt.inputs {
			path := amt.GetPath(ip)
			for j := range path {
				if !(identicalTries(path[j], tt.expected[i][j])) {
					t.Errorf("Does not match! Input: 0x%X", ip)
				}
			}
		}
	}
}

func TestFind(t *testing.T) {
	for _, tt := range []struct {
		expected []bool
		nodes    [][]byte
		inputs   [][]byte
	}{
		{
			nodes: [][]byte{
				{0xFA, 0x45, 0x01},
				{0xFB, 0x45, 0x01},
				{0x30, 0x45, 0x01},
				{0x30, 0x45, 0x02}},
			inputs: [][]byte{
				{0xFA, 0x45, 0x01},
				{0xFB, 0x45, 0x01},
				{0x30, 0x45, 0x01},
				{0x30, 0x45, 0x02},
				{0x30, 0x45, 0x03},
				{0x30, 0x45, 0x11},
				{0x30, 0x46, 0x01},
				{0x30, 0x55, 0x01},
				{0x31, 0x45, 0x01},
				{0x40, 0x45, 0x01},
			},
			expected: []bool{
				true,
				true,
				true,
				true,
				false,
				false,
				false,
				false,
				false,
				false,
			},
		},
	} {
		amt := InitAMT()
		for _, ip := range tt.nodes {
			amt.Insert(ip)
		}
		for i, ip := range tt.inputs {
			if tt.expected[i] != amt.Find(ip) {
				t.Errorf("Does not match! Input: 0x%X", ip)
			}
		}
	}
}

func TestInsert(t *testing.T) {
	for _, tt := range []struct {
		expected AMT
		nodes    [][]byte
	}{
		{ //Test Case 1
			nodes: [][]byte{
				{0xFA, 0x45, 0x01},
				{0xFB, 0x45, 0x01},
				{0x30, 0x45, 0x01},
				{0x30, 0x45, 0x02}},
			expected: AMT{root: &Node{
				bitmap: 0b1000000000001000,
				children: []*Node{
					{
						prefix: 4,
						value:  0x03,
						bitmap: 0b0000000000000001,
						children: []*Node{
							{
								prefix: 8,
								value:  0x00,
								bitmap: 0b0000000000010000,
								children: []*Node{
									{
										prefix: 12,
										value:  0x04,
										bitmap: 0b0000000000100000,
										children: []*Node{
											{
												prefix: 16,
												value:  0x05,
												bitmap: 0b0000000000000001,
												children: []*Node{
													{
														prefix: 20,
														value:  0x00,
														bitmap: 0b0000000000000110,
														children: []*Node{
															{
																prefix: 24,
																value:  0x01,
																bitmap: 0b0000000000000000,
																isLeaf: true,
															},
															{
																prefix: 24,
																value:  0x02,
																bitmap: 0b0000000000000000,
																isLeaf: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						prefix: 4,
						value:  0x0F,
						bitmap: 0b0000110000000000,
						children: []*Node{
							{
								prefix: 8,
								value:  0x0A,
								bitmap: 0b0000000000010000,
								children: []*Node{
									{
										prefix: 12,
										value:  0x04,
										bitmap: 0b0000000000100000,
										children: []*Node{
											{
												prefix: 16,
												value:  0x05,
												bitmap: 0b0000000000000001,
												children: []*Node{
													{
														prefix: 20,
														value:  0x00,
														bitmap: 0b0000000000000010,
														children: []*Node{
															{
																prefix: 24,
																value:  0x01,
																bitmap: 0b0000000000000000,
																isLeaf: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
							{
								prefix: 8,
								value:  0x0B,
								bitmap: 0b0000000000010000,
								children: []*Node{
									{
										prefix: 12,
										value:  0x04,
										bitmap: 0b0000000000100000,
										children: []*Node{
											{
												prefix: 16,
												value:  0x05,
												bitmap: 0b0000000000000001,
												children: []*Node{
													{
														prefix: 20,
														value:  0x00,
														bitmap: 0b0000000000000010,
														children: []*Node{
															{
																prefix: 24,
																value:  0x01,
																bitmap: 0b0000000000000000,
																isLeaf: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}},
		},
	} {
		amt := InitAMT()
		for _, ip := range tt.nodes {
			amt.Insert(ip)
		}
		if !(identicalTries(amt.root, tt.expected.root)) {
			t.Errorf("Does not match!")
		}
	}
}
