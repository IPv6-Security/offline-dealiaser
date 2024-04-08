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
package stats

import (
	"fmt"
	"testing"
)

func TestStatsRadix(t *testing.T) {
	for _, tt := range []struct {
		testIPs  string
		stepSize int
		expected [][]int
	}{
		{
			testIPs: "testIPs5.txt",
			expected: [][]int{
				{33, 2, 31},
				{34, 4, 30},
				{35, 5, 30},
				{52, 7, 45},
				{68, 9, 59},
			},
			stepSize: 1,
		},
		{
			testIPs: "testIPs10.txt",
			expected: [][]int{
				{33, 2, 31},
				{49, 4, 45},
				{66, 6, 60},
				{67, 8, 59},
				{68, 9, 59},
				{69, 10, 59},
				{71, 12, 59},
				{89, 14, 75},
				{121, 15, 106},
				{139, 17, 122},
			},
			stepSize: 1,
		},
	} {
		results := StatsRadix(tt.testIPs, tt.stepSize)
		for i, checkpoint := range results {
			for j := range checkpoint {
				if checkpoint[j] != tt.expected[i][j] {
					fmt.Println(checkpoint)
					fmt.Println(tt.expected[i])
					t.Fatalf("output does not match on %s-%d-%d", tt.testIPs, i, j)
				}
			}
		}
	}
}
