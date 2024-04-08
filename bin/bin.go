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
// Based on zgrab2/bin/bin.go
package bin

import (
	"aliasv6"
	"aliasv6/radix"
	"aliasv6/stats"
	"aliasv6/stress"
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	flags "github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

func check(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}

// Get the value of the ALIASV6_MEMPROFILE variable (or the empty string).
// This may include {TIMESTAMP} or {NANOS}, which should be replaced using
// getFormattedFile().
func getMemProfileFile() string {
	return os.Getenv("ALIASV6_MEMPROFILE")
}

// Get the value of the ALIASV6_CPUPROFILE variable (or the empty string).
// This may include {TIMESTAMP} or {NANOS}, which should be replaced using
// getFormattedFile().
func getCPUProfileFile() string {
	return os.Getenv("ALIASV6_CPUPROFILE")
}

// Replace instances in formatString of {TIMESTAMP} with when formatted as
// YYYYMMDDhhmmss, and {NANOS} as the decimal nanosecond offset.
func getFormattedFile(formatString string, when time.Time) string {
	timestamp := when.Format("20060102150405")
	nanos := fmt.Sprintf("%d", when.Nanosecond())
	ret := strings.Replace(formatString, "{TIMESTAMP}", timestamp, -1)
	ret = strings.Replace(ret, "{NANOS}", nanos, -1)
	return ret
}

// If memory profiling is enabled (ZGRAB2_MEMPROFILE is not empty), perform a GC
// then write the heap profile to the profile file.
func dumpHeapProfile() {
	if file := getMemProfileFile(); file != "" {
		now := time.Now()
		fullFile := getFormattedFile(file, now)
		f, err := os.Create(fullFile)
		if err != nil {
			log.Fatal("could not create heap profile: ", err)
		}
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write heap profile: ", err)
		}
		f.Close()
	}
}

// If CPU profiling is enabled (ALIASV6_CPUPROFILE is not empty), start tracking
// CPU profiling in the configured file. Caller is responsible for invoking
// stopCPUProfile() when finished.
func startCPUProfile() {
	if file := getCPUProfileFile(); file != "" {
		now := time.Now()
		fullFile := getFormattedFile(file, now)
		f, err := os.Create(fullFile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
	}
}

// If CPU profiling is enabled (ALIASV6_CPUPROFILE is not empty), stop profiling
// CPU usage.
func stopCPUProfile() {
	if getCPUProfileFile() != "" {
		pprof.StopCPUProfile()
	}
}

// Test is called when the user puts the program into test mode.
func Test(testType, constructInputFile, inputFile, outputFile string, stepSize int) {
	switch testType {
	case "radix":
		radix.Tester(constructInputFile, inputFile, stepSize)
	case "stats":
		stats.StatsRadix(constructInputFile, stepSize)
		stats.Stats(constructInputFile, outputFile, stepSize)
	case "stress":
		stress.StressTest(constructInputFile)
	}
}

func constructRadixTree(constructInputFile string, checkpointBaseName string, checkpointFrequency float32) *radix.Radix {

	fin, err := os.Open(constructInputFile)
	check(err)
	defer fin.Close()

	l := radix.InitRadix()
	l.SetCheckpointBaseName(checkpointBaseName)
	l.SetCheckpointFrequency(checkpointFrequency)

	scanner := bufio.NewScanner(fin)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		parsedIP, parsedNetwork, err := net.ParseCIDR(scanner.Text())
		check(err)
		if parsedIP == nil || parsedNetwork == nil {
			log.Warnf("couldn't parse input %s", scanner.Text())
			continue
		}
		l.Insert(parsedNetwork)
	}
	l.SetChange(false)
	if l.CheckConstructionNewAliasFound() {
		exportTime := time.Now()
		log.Infof("found new aliases while constructing the tree. exporting new prefixes to checkpoint-%s", exportTime.Format(time.RFC3339))
		l.ExportCheckpoint(exportTime)
	}
	return l
}

// AliasV6Main should be called by func main() in a binary. The caller is
// responsible for importing any modules in use. This allows clients to easily
// include custom sets of scan modules by creating new main packages with custom
// sets of AliasV6 modules imported with side-effects.
func AliasV6Main() {
	startCPUProfile()
	defer stopCPUProfile()
	defer dumpHeapProfile()

	_, config, err := aliasv6.ParseCommandLine(os.Args[1:])
	// Blanked arg is positional arguments
	if err != nil {
		// Outputting help is returned as an error. Exit successfuly on help output.
		flagsErr, ok := err.(*flags.Error)
		if ok && flagsErr.Type == flags.ErrHelp {
			return
		}

		// Didn't output help. Unknown parsing error.
		check(err)
	}

	if config.Test {
		start := time.Now()
		log.Infof("started test at %s", start.Format(time.RFC3339))
		Test(config.TestType, config.ConstructInputFile, config.TestOutputFile, config.TestInputFile, config.TestStepSize)
		end := time.Now()
		log.Infof("finished test at %s", end.Format(time.RFC3339))
		return
	}

	// Need an operator to receive commands from Generator
	// Operator Tasks:
	// 1. Construct the tree from the input file and initiate the timer for checkpoint checks
	// 2. Set up a monitor to keep track of successes and failures
	// 3. Initiate the Output Encoder to send results to stdout (Generator)
	// 4. Initiate workers to perform lookup when there is something in the
	// 	  processQueue. Operator should insert IPs which come from either
	// 	  Generator or Scanner to perform dealising on them.
	// 5. Initiate a worker to parse commands from Generator. Commands may
	//    include inserting a new aliased region etc. Commands would come from
	//    stdin.

	// Construct the tree from the input file
	l := constructRadixTree(config.ConstructInputFile, config.CheckpointBaseName, config.CheckpointFrequency)

	// Set up a monitor to keep track of successes and failures
	wg := sync.WaitGroup{}
	monitor := aliasv6.MakeMonitor(config.NumLookUpWorkers*4, &wg)
	monitor.Callback = func(_ string) {
		dumpHeapProfile()
	}

	start := time.Now()
	log.Infof("started dealiasing at %s", start.Format(time.RFC3339))

	processQueue := make(chan aliasv6.Command, config.NumLookUpWorkers*4)
	outputQueue := make(chan []byte, config.NumLookUpWorkers*4)

	mutex := &sync.RWMutex{}

	// Create wait groups
	var lookupWorkerDone sync.WaitGroup
	var outputDone sync.WaitGroup
	lookupWorkerDone.Add(config.NumLookUpWorkers)
	outputDone.Add(1)

	// Initiate the checkpoint goroutine
	var timerWorkerDone sync.WaitGroup
	timerWorkerDone.Add(1)
	ticker := time.NewTicker(time.Duration(l.GetCheckpointFrequency() * float32(time.Second)))
	quitTimerChannel := make(chan struct{})
	go func(mux *sync.RWMutex) {
		for {
			select {
			case <-ticker.C:
				ticker.Stop()
				mux.RLock()
				if l.IsChanged() {
					checkpointTime := time.Now()
					log.Infof("detected changes in the tree, creating a checkpoint at %s", checkpointTime.Format(time.RFC3339))
					l.ExportCheckpoint(checkpointTime)
				}
				mux.RUnlock()
				ticker.Reset(time.Duration(l.GetCheckpointFrequency() * float32(time.Second)))
			case <-quitTimerChannel:
				ticker.Stop()
				timerWorkerDone.Done()
				return
			}
		}
	}(mutex)

	// Start the output encoder
	go func() {
		defer outputDone.Done()
		if err := config.OutputResults(outputQueue); err != nil {
			log.Fatal(err)
		}
	}()

	// Start all the lookup workers
	for i := 0; i < config.NumLookUpWorkers; i++ {
		go func(mux *sync.RWMutex) {
			for obj := range processQueue {
				if obj.Type == "lookup" {
					mux.RLock()
					raw := aliasv6.RunLookUp(l, monitor, obj.ParsedData.(net.IP), config.Expanded)
					mux.RUnlock()
					result, err := json.Marshal(raw)
					if err != nil {
						log.Fatalf("unable to marshal data: %s", err)
					}
					outputQueue <- result
				} else if obj.Type == "insert" {
					mux.Lock()
					log.Infof("inserting %s", obj.ParsedData.(*net.IPNet))
					l.Insert(obj.ParsedData.(*net.IPNet))
					mux.Unlock()
				} else if obj.Type == "quit" {
					break
				}
			}
			lookupWorkerDone.Done()
		}(mutex)
	}

	if err := config.InputTargets(processQueue); err != nil {
		log.Fatal(err)
	}
	close(processQueue)
	lookupWorkerDone.Wait()
	close(outputQueue)
	outputDone.Wait()
	close(quitTimerChannel)
	timerWorkerDone.Wait()

	end := time.Now()
	log.Infof("finished dealiasing at %s", end.Format(time.RFC3339))

	monitor.Stop()
	wg.Wait()
	s := Summary{
		Status:    monitor.GetStatus(),
		StartTime: start.Format(time.RFC3339),
		EndTime:   end.Format(time.RFC3339),
		Duration:  end.Sub(start).String(),
	}
	enc := json.NewEncoder(aliasv6.GetMetaFile())
	if err := enc.Encode(&s); err != nil {
		log.Fatalf("unable to write summary: %s", err.Error())
	}
}
