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
package stress

import (
	"aliasv6/amt"
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

var memprofile = flag.String("memprofile", "", "write memory profile to this file")
var gcprofile = flag.String("gcprofile", "", "write garbage collector time profile to this file")
var haprofile = flag.String("haprofile", "", "write HeapAlloc profile to this file")

func StressTest(filename string) {
	flag.Parse()
	// if *cpuprofile != "" {
	// 	f, err := os.Create(*cpuprofile)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	pprof.StartCPUProfile(f)
	// 	defer pprof.StopCPUProfile()
	// }
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	var gf *os.File
	if *gcprofile != "" {
		gt, err := os.Create(*gcprofile)
		if err != nil {
			log.Fatal(err)
		}
		gf = gt
		defer gf.Close()
		gf.WriteString("[")
		// return
	}

	var ha *os.File
	if *haprofile != "" {
		hf, err := os.Create(*haprofile)
		if err != nil {
			log.Fatal(err)
		}
		ha = hf
		defer ha.Close()
		ha.WriteString("[")
		// return
	}

	var mem1, mem2 runtime.MemStats

	fmt.Println("memory baseline...")

	runtime.ReadMemStats(&mem1)
	log.Println(mem1.Alloc)
	log.Println(mem1.TotalAlloc)
	log.Println(mem1.HeapAlloc)
	log.Println(mem1.HeapSys)

	fin, err := os.Open(filename)
	check(err)
	defer fin.Close()

	scanner := bufio.NewScanner(fin)
	scanner.Split(bufio.ScanLines)

	counter := 0

	// amt.MemCheckForEachElement()

	amt := amt.InitAMT()
	// gcFlag := false

	for scanner.Scan() {
		parsedIP := net.ParseIP(scanner.Text()).To16()
		if parsedIP == nil {
			log.Println(scanner.Text())
			log.Println(counter)
			continue
		}
		amt.Insert(parsedIP)
		counter = counter + 1
		// if counter == 1000 {
		// 	break
		// }
		// gcFlag = false
		if counter%100000 == 0 {
			fmt.Printf("Progress: %d\n", counter)
			// runtime.ReadMemStats(&mem3)
			// log.Println(mem3.HeapInuse)
			// runtime.GC()
			log.Println("Running GC...")
			t1 := time.Now()
			debug.FreeOSMemory()
			t2 := time.Now()
			gf.WriteString(fmt.Sprintf("%f, ", t2.Sub(t1).Seconds()))
			log.Println("Done.")
			log.Println("Exporting Memory Profile...")
			if *memprofile != "" {
				f, err := os.Create(fmt.Sprintf("%s-%d.mprof", *memprofile, counter))
				if err != nil {
					log.Fatal(err)
				}
				pprof.WriteHeapProfile(f)
				f.Close()
				// return
			}
			log.Println("Done.")
			runtime.ReadMemStats(&mem2)
			log.Println(float64(mem2.Alloc-mem1.Alloc) / float64(1024*1024))
			log.Println(float64(mem2.TotalAlloc-mem1.TotalAlloc) / float64(1024*1024))
			log.Println(float64(mem2.HeapAlloc-mem1.HeapAlloc) / float64(1024*1024))
			log.Println(float64(mem2.HeapSys-mem1.HeapSys) / float64(1024*1024))
			log.Println("----")
			ha.WriteString(fmt.Sprintf("%f, ", float64(mem2.HeapAlloc-mem1.HeapAlloc)/float64(1024*1024)))
		}
		// if counter%1e7 == 0 {
		// 	log.Println("Running GC...")
		// 	debug.FreeOSMemory()
		// 	log.Println("Done.")
		// 	gcFlag = true
		// }
	}
	gf.WriteString("]\n")
	gf.Close()
	ha.WriteString("]\n")
	ha.Close()
	// if !gcFlag {
	// 	log.Println("Running GC...")
	// 	debug.FreeOSMemory()
	// 	log.Println("Done.")
	// }
	fmt.Printf("Progress: %d\n", counter)
	if *memprofile != "" {
		f, err := os.Create(fmt.Sprintf("%s-end.mprof", *memprofile))
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		// return
	}
	// time.Sleep(10 * time.Second)
	fmt.Println("memory comparison...")

	runtime.ReadMemStats(&mem2)
	log.Println(float64(mem2.Alloc-mem1.Alloc) / float64(1024*1024))
	log.Println(float64(mem2.TotalAlloc-mem1.TotalAlloc) / float64(1024*1024))
	log.Println(float64(mem2.HeapAlloc-mem1.HeapAlloc) / float64(1024*1024))
	log.Println(float64(mem2.HeapSys-mem1.HeapSys) / float64(1024*1024))
}
