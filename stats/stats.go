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
	"aliasv6/amt"
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func StatsRadix(filename string, stepSize int) [][]int {
	fin, err := os.Open(filename)
	check(err)
	defer fin.Close()

	scanner := bufio.NewScanner(fin)
	scanner.Split(bufio.ScanLines)

	counter := 0
	amt := amt.InitAMT()

	stats := make([][]int, 0)

	for scanner.Scan() {
		parsedIP := net.ParseIP(scanner.Text()).To16()
		if parsedIP == nil {
			continue
		}
		amt.Insert(parsedIP)
		counter = counter + 1
		if counter%stepSize == 0 {
			fmt.Printf("Progress: %d\n", counter)
			stats = append(stats, amt.TraverseBFSRadix(counter))
		}
	}
	fmt.Printf("Progress: %d\n", counter)
	stats = append(stats, amt.TraverseBFSRadix(counter))
	return stats
}

func Stats(filename, exportStats string, stepSize int) {
	fin, err := os.Open(filename)
	check(err)
	defer fin.Close()

	scanner := bufio.NewScanner(fin)
	scanner.Split(bufio.ScanLines)

	counter := 0
	amt := amt.InitAMT()

	for scanner.Scan() {
		parsedIP := net.ParseIP(scanner.Text()).To16()
		if parsedIP == nil {
			log.Println(scanner.Text())
			log.Println(counter)
			continue
		}
		amt.InsertWithCheckpoint(parsedIP)
		counter = counter + 1
		if counter%stepSize == 0 {
			fmt.Printf("Progress: %d\n", counter)
			amt.ExportStats(fmt.Sprintf("%s-%d.txt", exportStats, counter))
			amt.AddCheckpoint()
		}
	}
	fmt.Printf("Progress: %d\n", counter)
	amt.ExportStats(fmt.Sprintf("%s-%d.txt", exportStats, counter))
	amt.AddCheckpoint()
}
